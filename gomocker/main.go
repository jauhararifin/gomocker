package main

import (
	"runtime/debug"

	"github.com/jauhararifin/gomocker/gomocker/cmd"
)

var version = "v1.1.0"

func main() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		version = buildInfo.Main.Version
	}
	cmd.Execute(version)
}
