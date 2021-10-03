package cmd

import (
	"github.com/innatical/pax/util"
	"github.com/urfave/cli/v2"
)

func Search(c *cli.Context) error {
	if err := util.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer util.UnlockDatabase(c.String("root"))


	list, err := util.Search(c.String("root"), c.Args().First())
	if err != nil {
		return err
	}

	for i := range list {
		println(list[i])
	}

	return nil
}
