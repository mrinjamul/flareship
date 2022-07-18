package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Configured via -ldflags during build
// Version is the version of the binary
var Version = "dev"

// GitCommit is the git commit hash of the binary
var GitCommit string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints version.",
	Run: func(cmd *cobra.Command, args []string) {
		shortCommit := shortGitCommit(GitCommit)
		version := fmt.Sprintf("Version: %s %s", Version, shortCommit)
		fmt.Println(version)
	},
}

// shortGitCommit returns the short form of the git commit hash
func shortGitCommit(fullGitCommit string) string {
	shortCommit := ""
	if len(fullGitCommit) >= 7 {
		shortCommit = fullGitCommit[0:7]
	}

	return shortCommit
}
