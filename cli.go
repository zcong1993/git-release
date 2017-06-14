package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/mitchellh/colorstring"
	"github.com/tcnksm/go-gitconfig"
	"github.com/tj/go-prompt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"html/template"
	"io"
	"log"
	"os"
	"strings"
)

const (
	//EnvGitHubToken is environmental var to set GitHub API token
	EnvGitHubToken = "GITHUB_TOKEN"
	// EnvGitHubAPI is environmental var to set GitHub API base endpoint.
	// This is used mainly by GitHub Enterprise user.
	EnvGitHubAPI   = "GITHUB_API"
	defaultBaseURL = "https://api.github.com/"
)

const (
	// ExitCodeOK is exit code 0
	ExitCodeOK int = 0

	// Errors start at 10

	// ExitCodeError is exit code 10
	ExitCodeError = 10 + iota
	// ExitCodeParseFlagsError is exit code 11
	ExitCodeParseFlagsError
	// ExitCodeBadArgs is exit code 12
	ExitCodeBadArgs
	// ExitCodeTokenNotFound is exit code 13
	ExitCodeTokenNotFound
	// ExitCodeOwnerNotFound is exit code 14
	ExitCodeOwnerNotFound
)

// CLI is the command line object
type CLI struct {
	outStream, errStream io.Writer
}

// Data is data type of commit log message user chose
type Data struct {
	Majors  []Commit
	Minors  []Commit
	Patches []Commit
	Ignore  []Commit
}

// Commit is type of simplify git commit log
type Commit struct {
	Message string
	Hash    plumbing.Hash
}

var tpl = `
{{if .Majors|len}}
### Major Changes
{{range .Majors}}
  - {{.Message|format}}: {{.Hash}}
{{end}}
{{end}}
{{if .Minors|len}}
### Minor Changes
{{range .Minors}}
  - {{.Message|format}}: {{.Hash}}
{{end}}
{{end}}
{{if .Patches|len}}
### Patches
{{range .Patches}}
  - {{.Message|format}}: {{.Hash}}
{{end}}
{{end}}
`

var types = []string{"Major Change", "Minor Change", "Patch", "Ignore", "End"}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		owner      string
		repo       string
		token      string
		help       bool
		version    bool
		commitish  string
		draft      bool
		prerelease bool
		recreate   bool
	)
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)
	flags.Usage = func() {
		fmt.Fprint(cli.errStream, helpText)
	}

	flags.StringVar(&owner, "username", "", "")
	flags.StringVar(&owner, "owner", "", "")
	flags.StringVar(&owner, "u", "", "")

	flags.StringVar(&repo, "repository", "", "")
	flags.StringVar(&repo, "r", "", "")

	flags.StringVar(&token, "token", os.Getenv(EnvGitHubToken), "")
	flags.StringVar(&token, "t", os.Getenv(EnvGitHubToken), "")

	flags.BoolVar(&help, "help", false, "")
	flags.BoolVar(&help, "h", false, "")

	flags.BoolVar(&version, "version", false, "")
	flags.BoolVar(&version, "v", false, "")

	flags.StringVar(&commitish, "commitish", "", "")
	flags.StringVar(&commitish, "c", "", "")

	flags.BoolVar(&draft, "draft", false, "")
	flags.BoolVar(&prerelease, "prerelease", false, "")

	flags.BoolVar(&recreate, "delete", false, "")
	flags.BoolVar(&recreate, "recreate", false, "")
	flags.BoolVar(&recreate, "update", false, "")

	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeParseFlagsError
	}

	if help {
		fmt.Fprint(cli.errStream, helpText)
		return 0
	}
	if version {
		ShowVersion()
		return 0
	}
	parsedArgs := flags.Args()
	if len(parsedArgs) != 2 {
		PrintRedf(cli.errStream,
			"Invalid argument: you must set TAG and PATH name.")
		return ExitCodeBadArgs
	}
	tag, path := parsedArgs[0], parsedArgs[1]

	if len(owner) == 0 {
		var err error
		owner, err = gitconfig.GithubUser()
		if err != nil {
			owner, err = gitconfig.Username()
		}

		if err != nil {
			PrintRedf(cli.errStream,
				"Failed to set up rls: repository owner name not found\n")
			fmt.Fprintf(cli.errStream,
				"Please set it via `-u` option.\n\n"+
					"You can set default owner name in `github.username` or `user.name`\n"+
					"in `~/.gitconfig` file")
			return ExitCodeOwnerNotFound
		}
	}

	if len(repo) == 0 {
		var err error
		repo, err = gitconfig.Repository()
		if err != nil {
			PrintRedf(cli.errStream,
				"Failed to set up rls: repository name not found\n")
			fmt.Fprintf(cli.errStream,
				"rls reads it from `.git/config` file. Change directory to \n"+
					"repository root directory or setup git repository.\n"+
					"Or set it via `-r` option.\n")
			return ExitCodeOwnerNotFound
		}
	}

	if len(token) == 0 {
		var err error
		token, err = gitconfig.GithubToken()
		if err != nil {
			PrintRedf(cli.errStream, "Failed to set up rls: token not found\n")
			fmt.Fprintf(cli.errStream,
				"To use rls, you need a GitHub API token.\n"+
					"Please set it via `%s` env var or `-t` option.\n\n"+
					"If you don't have one, visit official doc (goo.gl/jSnoI)\n"+
					"and get it first.\n",
				EnvGitHubToken)
			return ExitCodeTokenNotFound
		}
	}
	baseURLStr := defaultBaseURL
	if urlStr := os.Getenv(EnvGitHubAPI); len(urlStr) != 0 {
		baseURLStr = urlStr
	}
	gitHubClient, err := NewGitHubClient(owner, repo, token, baseURLStr)
	if err != nil {
		PrintRedf(cli.errStream, "Failed to construct GitHub client: %s", err)
		return ExitCodeError
	}
	rls := RLS{
		GitHub:    gitHubClient,
		outStream: cli.outStream,
	}

	g, err := git.PlainOpen(path)
	checkErr(err)
	logs, err := g.Log(&git.LogOptions{})
	checkErr(err)
	Data := inquired(logs)
	var bf bytes.Buffer
	compile(Data, &bf)
	req := &github.RepositoryRelease{
		Name:            github.String(tag),
		TagName:         github.String(tag),
		Prerelease:      github.Bool(prerelease),
		Draft:           github.Bool(draft),
		TargetCommitish: github.String(commitish),
		Body:            github.String(bf.String()),
	}
	ctx := context.TODO()
	release, err := rls.CreateRelease(ctx, req, recreate)
	if err != nil {
		PrintRedf(cli.errStream, "Failed to create GitHub release page: %s", err)
		return ExitCodeError
	}
	PrintBluef(cli.outStream, "\nRelease success! %s", *release.HTMLURL)
	return ExitCodeOK
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func inquired(logs object.CommitIter) *Data {
	var Data Data
	for true {
		log, err := logs.Next()
		if err != nil {
			break
		}
		query := fmt.Sprintf("Commit '%s' is a change of: ", strings.Replace(log.Message, "\n", "", 1))
		i := prompt.Choose(query, types)
		if i == 0 {
			Data.Majors = append(Data.Majors, Commit{log.Message, log.Hash})
		}
		if i == 1 {
			Data.Minors = append(Data.Minors, Commit{log.Message, log.Hash})
		}
		if i == 2 {
			Data.Patches = append(Data.Patches, Commit{log.Message, log.Hash})
		}
		if i == 4 {
			break
		}
	}
	return &Data
}

func formatString(s string) string {
	l := 33
	s = strings.Replace(s, "\n", "", 1)
	b := []byte(s)
	s = strings.ToUpper(string(b[0])) + string(b[1:])
	if len(s) < l {
		return s
	}
	return string(b[0:l-3]) + "..."
}

func compile(data *Data, w io.Writer) {
	t := template.New("tpl")
	t = t.Funcs(template.FuncMap{"format": formatString})
	t = template.Must(t.Parse(tpl))
	t.Execute(w, data)
}

// PrintRedf is helper function printf with color red
func PrintRedf(w io.Writer, format string, args ...interface{}) {
	format = fmt.Sprintf("[red]%s[reset]", format)
	fmt.Fprint(w,
		colorstring.Color(fmt.Sprintf(format, args...)))
}

// PrintBluef is helper function printf with color blue
func PrintBluef(w io.Writer, format string, args ...interface{}) {
	format = fmt.Sprintf("[blue]%s[reset]", format)
	fmt.Fprint(w,
		colorstring.Color(fmt.Sprintf(format, args...)))
}

var helpText = `

	Usage: rls [options...] TAG PATH

rls is a tool to create Release on Github.

You must specify tag (e.g., v1.0.0) and PATH to local git workspace folder.

And you also must provide GitHub API token which has enough permission
(For a private repository you need the 'repo' scope and for a public
repository need 'public_repo' scope). You can get token from GitHub's
account setting page.

You can use rls on GitHub Enterprise. Set base URL via GITHUB_API
environment variable.

Options:
  -username, -u      Github repository onwer name. By default, rls
                     extracts it from global gitconfig value.
  -repository, -r    GitHub repository name. By default, rls extracts
                     repository name from current directory's .git/config.
  -token, -t         GitHub API Token. By default, rls reads it from
                     'GITHUB_TOKEN' env var.
  -recreate          Recreate release if it already exists. If want to
                     upload to same release and replace use '-replace'.

`
