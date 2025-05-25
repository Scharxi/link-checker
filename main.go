package main

import (
	"fmt"
	"os"

	"bxfferoverflow.me/link-checker/linkchecker/cli"
)

// Build information (set by ldflags during build)
var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "unknown"
)

func main() {
	// Set version info for CLI
	cli.SetVersionInfo(version, buildTime, commit)

	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
