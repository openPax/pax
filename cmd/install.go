package cmd

import (
	apkg "github.com/innatical/apkg/v2/util"
	"github.com/innatical/pax/v2/util"
	"github.com/urfave/cli/v2"
)

func Install(c *cli.Context) error {
	if err := apkg.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer apkg.UnlockDatabase(c.String("root"))

	if err := util.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer util.UnlockDatabase(c.String("root"))

	if err := util.InstallMultiple(c.String("root"), c.Args().Slice(), c.Bool("optional")); err != nil {
		return err
	}

	db, err := util.ReadDatabase(c.String("root"))

	if err != nil {
		return err
	}

	db.Packages = append(db.Packages, c.Args().Slice()...)

	if err := util.WriteDatabase(c.String("root"), db); err != nil {
		return err
	}

	return nil
}
