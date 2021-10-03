package cmd

import (
	"github.com/innatical/pax/util"
	apkg "github.com/innatical/apkg/util"
	"github.com/urfave/cli/v2"
)

func List(c *cli.Context) error {
	if err := apkg.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer apkg.UnlockDatabase(c.String("root"))

	if err := util.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer util.UnlockDatabase(c.String("root"))

	db, err := apkg.ReadDatabase(c.String("root"))

	if err != nil {
		return err
	}

	for k, d := range db.Packages {
		println(k + "@" + d.Package.Version)
	}

	return nil
}