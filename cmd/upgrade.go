package cmd

import (
	apkg "github.com/innatical/apkg/v2/util"
	"github.com/innatical/pax/v3/util"
	"github.com/urfave/cli/v2"
)

func Upgrade(c *cli.Context) error {
	if err := apkg.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer apkg.UnlockDatabase(c.String("root"))

	if err := util.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer util.UnlockDatabase(c.String("root"))

	packages := c.Args().Slice()

	if len(packages) == 0 {
		db, err := apkg.ReadDatabase(c.String("root"))
		if err != nil {
			return err
		}

		for k := range db.Packages {
			deps, err := util.GetDepdendents(c.String("root"), k)
			if err != nil {
				return err
			}

			if len(deps) == 0 {
				packages = append(packages, k)
			}
		}
	}

	if len(packages) == 0 {
		return &apkg.ErrorString{S: "Errno 3: No Packages to Upgrade"}
	}

	for _, v := range packages {
		util.Upgrade(c.String("root"), c.String("cache"), v, !c.Bool("latest"))
	}

	return nil
}
