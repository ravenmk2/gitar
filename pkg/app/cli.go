package app

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	AppName = "gitar"
)

func RunCliApp() error {
	app := NewCliApp()
	return app.Run(os.Args)
}

func NewCliApp() *cli.App {
	app := &cli.App{
		Name:        AppName,
		Usage:       AppName,
		Description: "Git Archive & Repository Tool",
		Commands: []*cli.Command{
			NewMirrorCommand(),
			NewDownloadCommand(),
		},
	}
	return app
}

func NewMirrorCommand() *cli.Command {
	return &cli.Command{
		Name:    "mirror",
		Aliases: []string{"mi"},
		Usage:   "Make a git repository mirror",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "debug", Required: false, Value: false},
			&cli.BoolFlag{Name: "mail", Aliases: []string{"m"}, Required: false, Value: false},
		},
		Action: func(ctx *cli.Context) error {
			debug := ctx.Bool("debug")
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			url := ctx.Args().First()
			return MirrorRepository(url, ctx.Bool("mail"))
		},
	}
}

func NewDownloadCommand() *cli.Command {
	return &cli.Command{
		Name:    "download",
		Aliases: []string{"dl"},
		Usage:   "Download a git archive",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "debug", Required: false, Value: false},
			&cli.BoolFlag{Name: "mail", Aliases: []string{"m"}, Required: false, Value: false},
		},
		Action: func(ctx *cli.Context) error {
			debug := ctx.Bool("debug")
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			url := ctx.Args().First()
			return DownloadArchive(url, ctx.Bool("mail"))
		},
	}
}
