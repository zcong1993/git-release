package main

import "fmt"

// Name is cli name
const Name = "rls"

// Version is cli current version
const Version string = "v0.1.0"

// GitCommit is cli current git commit hash
var GitCommit string

// ShowVersion is handler for version command
func ShowVersion() {
	version := fmt.Sprintf("%s version %s", Name, Version)
	if len(GitCommit) != 0 {
		version += fmt.Sprintf(" (%s)", GitCommit)
	}
	fmt.Println(version)
}
