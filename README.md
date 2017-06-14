# git-release

[![Go Report Card](https://goreportcard.com/badge/github.com/zcong1993/git-release)](https://goreportcard.com/report/github.com/zcong1993/git-release)
[![Build Status](https://travis-ci.org/zcong1993/git-release.svg?branch=master)](https://travis-ci.org/zcong1993/git-release)

> Generate github release with changelog in a single command

Based on [tcnksm/ghr](https://github.com/tcnksm/ghr)

## Install
Download binary from [release](https://github.com/zcong1993/git-release/releases) page and place it in `$PATH` directory.

Or build yourself
```bash
$ go get -v -u github.com/zcong1993/git-release
$ cd $GOPATH/github.com/zcong1993/git-release
$ make build
# then place ./bin/rls in `$PATH` directory.
```

## Usage
```bash
$ rls [option] TAG
```
`TAG` is required. TAG is the release tag. And must config github api token first.

### GitHub API Token

To get token, first, visit GitHub account settings page, then go to Applications for the user. Here you can create a token in the Personal access tokens section. For a private repository you need repo scope and for a public repository you need public_repo scope.

When using ghr, you can set it via GITHUB_TOKEN env var, -token command line option or github.token property in .gitconfig file.

For instance, to set it via environmental variable:
```bash
$ export GITHUB_TOKEN="....."
```
Or set it in github.token in gitconfig:

```bash
$ git config --global github.token "....."
```
Note that environmental variable takes priority over gitconfig value.

### GitHub Enterprise

You can use ghr for GitHub Enterprise. Change API endpoint via the enviromental variable.
```bash
$ export GITHUB_API=http://github.company.com/api/v3/
```

### Options

```bash
$ rls \
    -t TOKEN \        # Set Github API Token
    -u USERNAME \     # Set Github username
    -r REPO \         # Set repository name
    -c COMMIT \       # Set target commitish, branch or commit SHA
    -delete \         # Delete release and its git tag in advance if it exists
    -draft \          # Release as draft (Unpublish)
    -prerelease \     # Crate prerelease
    TAG PATH
```

## License

MIT &copy; zcong1993
