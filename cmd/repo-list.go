package cmd

import (
	apkg "github.com/innatical/apkg/v2/util"
	"github.com/innatical/pax/v2/util"
	"github.com/urfave/cli/v2"
)

func RepoList(c *cli.Context) error {
	if err := apkg.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer apkg.UnlockDatabase(c.String("root"))

	if err := util.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer util.UnlockDatabase(c.String("root"))

	repos, err := util.ReadReposList(c.String("root"))

	if err != nil {
		return err
	}

	if len(repos.Repos) == 0 {
		return &apkg.ErrorString{S: "No repositories added"}
	}

	table := make(map[string]string)
	maxWidth := 0

	for k, d := range repos.Repos {
		table[k] = d

		lineWidth := len(k) + 5 + len(d)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	println(apkg.RenderTable(table, maxWidth))

	return nil
}
