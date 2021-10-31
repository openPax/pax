package cmd

import (
	apkg "github.com/innatical/apkg/v2/util"
	"github.com/innatical/pax/v2/util"
	"github.com/urfave/cli/v2"
)

func Add(c *cli.Context) error {
	if err := apkg.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer apkg.UnlockDatabase(c.String("root"))

	if err := util.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer util.UnlockDatabase(c.String("root"))

	util.UpdateSourcesList(c.String("root"), c.Args().First())

	return nil
}
