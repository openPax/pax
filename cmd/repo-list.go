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

	for k, d := range repos.Repos {
		println(k + ": " + d)
	}

	return nil
}
