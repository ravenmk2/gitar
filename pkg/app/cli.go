package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	AppName = "gitar"
)

func RunCliApp() error {
	app := NewCliApp()
	return app.Run(reorderArgs(os.Args))
}

// valueFlags 是需要跟随一个值的 flag（其余均视为 bool flag）
var valueFlags = map[string]bool{
	"retry": true,
}

// reorderArgs 将 flag 参数移到位置参数之前。
// urfave/cli/v2 底层使用标准库 flag 解析，遇到第一个位置参数即停止解析，
// 导致 `gitar mi <url> --retry 3` 中 --retry 被忽略，这里统一调整顺序。
func reorderArgs(args []string) []string {
	if len(args) <= 2 {
		return args
	}

	head := args[:2] // 程序名 + 子命令
	var flags, positionals []string
	rest := args[2:]
	for i := 0; i < len(rest); i++ {
		arg := rest[i]
		if arg == "--" { // -- 之后全部为位置参数
			positionals = append(positionals, rest[i+1:]...)
			break
		}
		if !isFlagArg(arg) {
			positionals = append(positionals, arg)
			continue
		}
		flags = append(flags, arg)
		name := strings.TrimLeft(arg, "-")
		if idx := strings.IndexByte(name, '='); idx >= 0 {
			continue // --flag=value 形式，值已包含在同一参数中
		}
		if valueFlags[name] && i+1 < len(rest) {
			i++
			flags = append(flags, rest[i])
		}
	}
	return append(head, append(flags, positionals...)...)
}

func isFlagArg(arg string) bool {
	return len(arg) > 1 && arg[0] == '-'
}

func NewCliApp() *cli.App {
	cli.VersionPrinter = func(ctx *cli.Context) {
		fmt.Println(VersionString())
	}
	app := &cli.App{
		Name:        AppName,
		Usage:       AppName,
		Description: "Git Archive & Repository Tool",
		Version:     Version,
		Commands: []*cli.Command{
			NewMirrorCommand(),
			NewDownloadCommand(),
			NewVersionCommand(),
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
			&cli.BoolFlag{Name: "ssh", Required: false, Value: false},
			&cli.BoolFlag{Name: "mail", Aliases: []string{"m"}, Required: false, Value: false},
			&cli.IntFlag{Name: "retry", Required: false, Value: 20, Usage: "max retry attempts"},
		},
		Action: func(ctx *cli.Context) error {
			debug := ctx.Bool("debug")
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			url := ctx.Args().First()
			return MirrorRepository(url, ctx.Bool("ssh"), ctx.Bool("mail"), ctx.Int("retry"))
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
			&cli.IntFlag{Name: "retry", Required: false, Value: 20, Usage: "max retry attempts"},
		},
		Action: func(ctx *cli.Context) error {
			debug := ctx.Bool("debug")
			if debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			url := ctx.Args().First()
			return DownloadArchive(url, ctx.Bool("mail"), ctx.Int("retry"))
		},
	}
}
