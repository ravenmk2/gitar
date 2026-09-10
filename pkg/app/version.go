package app

import (
	"fmt"
	"runtime/debug"

	"github.com/urfave/cli/v2"
)

var (
	Version = "dev"
	GitSHA  = ""
)

func NewVersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Show version information",
		Action: func(ctx *cli.Context) error {
			fmt.Println(VersionString())
			return nil
		},
	}
}

func VersionString() string {
	sha := GitSHA
	if sha == "" {
		sha = vcsRevision()
	}
	if sha == "" {
		sha = "unknown"
	}
	return fmt.Sprintf("%s %s (%s)", AppName, Version, sha)
}

func vcsRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	revision := ""
	modified := false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	if revision != "" && modified {
		revision += "-dirty"
	}
	return revision
}
