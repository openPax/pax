package main

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/innatical/pax/v2/cmd"

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
		Name:      "pax",
		Usage:     "Pax Package Manager",
		UsageText: "pax [global options] command [command options] [arguments...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "root",
				Value: filepath.Join(usr.HomeDir, "/.apkg"),
				Usage: "The Root Directory for the apkg Package Backend",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "install",
				Usage:     "Install a Package",
				UsageText: "pax install <package names>",
				Aliases:   []string{"i", "in"},
				Action:    cmd.Install,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "optional",
						Value: true,
						Usage: "Install Optional Dependencies",
					},
				},
			},
			{
				Name:      "remove",
				Usage:     "Remove a Package",
				UsageText: "pax remove <package names>",
				Aliases:   []string{"r"},
				Action:    cmd.Remove,
			},
			{
				Name:      "list",
				Usage:     "List All Installed Packages",
				UsageText: "pax list",
				Aliases:   []string{"l"},
				Action:    cmd.List,
			},
			{
				Name:      "search",
				Usage:     "Search All Packages",
				UsageText: "pax search",
				Aliases:   []string{"s"},
				Action:    cmd.Search,
			},
			{
				Name:      "upgrade",
				Usage:     "Upgrade All Packages",
				UsageText: "pax upgrade [package names]",
				Aliases:   []string{"u"},
				Action:    cmd.Upgrade,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "latest",
						Value:   false,
						Usage:   "Grab the Latest Compatible, Self Semver Breaking Packages",
						Aliases: []string{"L"},
					},
				},
			},
			{
				Name:      "info",
				Usage:     "Get the Information for a Package",
				UsageText: "pax info <package file|package name>",
				Action:    cmd.Info,
			},
			{
				Name:      "repo",
				Usage:     "Manage Pax Repositories",
				UsageText: "pax repo <subcomamnd>",
				Aliases:   []string{"re"},
				Subcommands: []*cli.Command{
					{
						Name:      "add",
						Usage:     "Add new Pax Repository",
						UsageText: "pax repo add <repository>",
						Action:    cmd.Add,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		println(errorStyle.Render("Error: ") + err.Error())
		os.Exit(1)
	}
}
