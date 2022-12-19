package main

import "github.com/thiagozs/githubpal/cmd"

var (
	Version string
)

func main() {
	cmd.Version = Version
	cmd.Execute()
}
