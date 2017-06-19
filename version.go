package main

import (
	"bytes"
	"fmt"
	latest "github.com/tcnksm/go-latest"
	"time"
)

// Name is cli name
const Name = "rls"

// Version is cli current version
const Version string = "v0.2.0"

// GitCommit is cli current git commit hash
var GitCommit string

const defaultCheckTimeout = 2 * time.Second

// ShowVersion is handler for version command
func ShowVersion() {
	version := fmt.Sprintf("%s version %s", Name, Version)
	if len(GitCommit) != 0 {
		version += fmt.Sprintf(" (%s)", GitCommit)
	}
	fmt.Println(version)
	var buf bytes.Buffer
	verCheckCh := make(chan *latest.CheckResponse)
	go func() {
		fixFunc := latest.DeleteFrontV()
		githubTag := &latest.GithubTag{
			Owner:             "zcong1993",
			Repository:        "git-release",
			FixVersionStrFunc: fixFunc,
		}

		res, err := latest.Check(githubTag, fixFunc(Version))
		if err != nil {
			// Don't return error
			return
		}
		verCheckCh <- res
	}()

	select {
	case <-time.After(defaultCheckTimeout):
	case res := <-verCheckCh:
		if res.Outdated {
			fmt.Fprintf(&buf,
				"Latest version of rls is v%s, please upgrade!\n",
				res.Current)
		}
	}
	fmt.Print(buf.String())
}
