package cmd

import (
	"github.com/charmbracelet/lipgloss"
	apkg "github.com/innatical/apkg/v2/util"
	"github.com/innatical/pax/v2/util"
	"github.com/urfave/cli/v2"
)

func Search(c *cli.Context) error {

	if err := util.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer util.UnlockDatabase(c.String("root"))

	println("Searching for " + lipgloss.NewStyle().Bold(true).Render(c.Args().First()))

	list, err := util.Search(c.String("root"), c.Args().First())
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return &apkg.ErrorString{S: "No package " + lipgloss.NewStyle().Bold(true).Render(c.Args().First()) + " found"}
	}

	table := make(map[string]string)
	maxWidth := 0

	table[lipgloss.NewStyle().Bold(true).Underline(true).Render("Package")] = lipgloss.NewStyle().Bold(true).Underline(true).Render("Version")

	for k, d := range list {

		table[k] = d

		lineWidth := len(k) + 10 + len(d)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	println(apkg.RenderTable(table, maxWidth))

	return nil
}
