package main

import "fmt"

const Name = "rls"
const Version string = "v0.1.0"

var GitCommit string

func ShowVersion() {
	version := fmt.Sprintf("%s version %s", Name, Version)
	if len(GitCommit) != 0 {
		version += fmt.Sprintf(" (%s)", GitCommit)
	}
	fmt.Println(version)
}
