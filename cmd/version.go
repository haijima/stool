package cmd

import (
	"runtime/debug"
)

// Version is set in build step
var Version = ""

func version() string {
	if Version != "" {
		return Version
	}
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		return buildInfo.Main.Version
	}
	return "(devel)"
}
