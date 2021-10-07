package main

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/innatical/pax/cmd"

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
		UsageText: "Pax [global options] command [command options] [arguments...]",
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
				Aliases:   []string{"i"},
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
<<<<<<< HEAD
				Usage:     "List all installed packages",
				UsageText: "alt list",
=======
				Usage:     "List All Installed Packages",
				UsageText: "apkg list",
>>>>>>> upstream
				Aliases:   []string{"l"},
				Action:    cmd.List,
			},
			{
				Name:      "search",
<<<<<<< HEAD
				Usage:     "Search all packages",
				UsageText: "alt search",
=======
				Usage:     "Search All Packages",
				UsageText: "apkg search",
>>>>>>> upstream
				Aliases:   []string{"s"},
				Action:    cmd.Search,
			},
			{
				Name:      "upgrade",
<<<<<<< HEAD
				Usage:     "upgrade all packages",
				UsageText: "alt upgrade [package names]",
=======
				Usage:     "Upgrade All Packages",
				UsageText: "apkg upgrade [package names]",
>>>>>>> upstream
				Aliases:   []string{"u"},
				Action:    cmd.Upgrade,
				Flags: []cli.Flag{
					&cli.BoolFlag{
<<<<<<< HEAD
						Name:    "latest",
						Value:   false,
						Usage:   "Grab the latest compatible, self semver breaking packages",
=======
						Name:  "latest",
						Value: false,
						Usage: "Grab the Latest Compatible, Self Semver Breaking Packages",
>>>>>>> upstream
						Aliases: []string{"L"},
					},
				},
			},
			{
				Name:      "info",
<<<<<<< HEAD
				Usage:     "Get the information for a package",
				UsageText: "alt info <package file|package name>",
=======
				Usage:     "Get the Information for a Package",
				UsageText: "apkg info <package file|package name>",
>>>>>>> upstream
				Aliases:   []string{"in"},
				Action:    cmd.Info,
			},
			{
				Name:      "repo",
				Usage:     "Manage Alt Repositories",
				UsageText: "alt repo <subcomamnd>",
				Aliases:   []string{"r"},
				Subcommands: []*cli.Command{
					// For Later
					{
						Name:      "import-key",
						Usage:     "Import Alt Repository GPG Key",
						UsageText: "alt repo import-key <keyfile>",
						Action: func(c *cli.Context) error {
							c.Args().First()
							println("Test this command, alt repo imoprt-key")
							return nil
						},
					},
					{
						Name:      "add",
						Usage:     "Add new Alt Repository",
						UsageText: "alt repo add <repository>",
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
