package cmd

import (
	apkg "github.com/innatical/apkg/v2/util"
	"github.com/innatical/pax/v3/util"
	"github.com/urfave/cli/v2"
)

func RepoAdd(c *cli.Context) error {
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

	repos.Repos[c.Args().Get(0)] = c.Args().Get(1)

	if err := util.WriteReposList(c.String("root"), repos); err != nil {
		return err
	}

	return nil
}
