package main

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/innatical/alt/cmd"

	"github.com/urfave/cli/v2"
)

var errorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000"))

func main() {
	usr, err := user.Current()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	app := &cli.App{
		Name:      "alt",
		Usage:     "The alt package manager",
		UsageText: "alt [global options] command [command options] [arguments...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "root",
				Value: filepath.Join(usr.HomeDir, "/.apkg"),
				Usage: "The root directory for the apkg package backend",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "install",
				Usage:     "Install a package",
				UsageText: "alt install <package names>",
				Aliases:   []string{"i"},
				Action:    cmd.Install,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "optional",
						Value: true,
						Usage: "Install optional dependencies",
					},
				},
			},
			{
				Name:      "remove",
				Usage:     "Remove a package",
				UsageText: "alt remove <package names>",
				Aliases:   []string{"r"},
				Action:    cmd.Remove,
			},
			{
				Name:      "list",
				Usage:     "List all installed packages",
				UsageText: "apkg list",
				Aliases:   []string{"l"},
				Action:    cmd.List,
			},
			{
				Name:      "search",
				Usage:     "Search all packages",
				UsageText: "apkg search",
				Aliases:   []string{"s"},
				Action:    cmd.Search,
			},
			{
				Name:      "upgrade",
				Usage:     "upgrade all packages",
				UsageText: "apkg upgrade [package names]",
				Aliases:   []string{"u"},
				Action:    cmd.Upgrade,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "latest",
						Value: false,
						Usage: "Grab the latest compatible, self semver breaking packages",
						Aliases: []string{"L"},
					},
				},
			},
			{
				Name:      "info",
				Usage:     "Get the information for a package",
				UsageText: "apkg info <package file|package name>",
				Aliases:   []string{"in"},
				Action:    cmd.Info,
			},
		},
	}
	
	if err := app.Run(os.Args); err != nil {
		println(errorStyle.Render("Error: ") + err.Error())
		os.Exit(1)
	}
}
