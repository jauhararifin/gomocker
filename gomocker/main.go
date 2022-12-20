package main

import (
	"runtime/debug"

	"github.com/jauhararifin/gomocker/gomocker/cmd"
)

func main() {
	version := "unknown"
	info, ok := debug.ReadBuildInfo()
	if ok {
		version = info.Main.Version
	}
	cmd.Execute(version)
}
