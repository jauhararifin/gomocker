package main

import (
	"runtime/debug"

	"github.com/jauhararifin/gomocker/gomocker/cmd"
)

var version = "v1.0.1"

func main() {
	if buildInfo, ok := debug.ReadBuildInfo(); ok && version == "unknown" {
		version = buildInfo.Main.Version
	}
	cmd.Execute(version)
}
