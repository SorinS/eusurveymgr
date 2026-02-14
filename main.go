package main

import "eusurveymgr/cmd"

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, buildDate)
	cmd.Execute()
}