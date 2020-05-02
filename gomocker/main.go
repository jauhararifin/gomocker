package main

import (
	"runtime/debug"

	"github.com/jauhararifin/gomocker/gomocker/cmd"
)

var version = "unknown"

func main() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok && version == "unknown" {
		version = buildInfo.Main.Version
	}
	cmd.Execute(version)
}
