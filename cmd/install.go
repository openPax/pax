package cmd

import (
	"strings"

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

	for _, name := range c.Args().Slice() {
		parsed := strings.Split(name, "@")
		installed, err := util.IsInstalledName(c.String("root"), parsed[0])

		if err != nil {
			return err
		}

		if installed {
			return &apkg.ErrorString{S: "Errno 1: Package " + name + " Already Installed"}
		}


		if len(parsed) == 1 {
			if err := util.Install(c.String("root"), parsed[0], "", c.Bool("optional")); err != nil {
				return err
			}
		} else {
			if err := util.Install(c.String("root"), parsed[0], parsed[1], c.Bool("optional")); err != nil {
				return err
			}
		}

		db, err := util.ReadDatabase(c.String("root"))

		if err != nil {
			return err
		}

		db.Packages = append(db.Packages, name)

		if err := util.WriteDatabase(c.String("root"), db); err != nil {
			return err
		}
	}

	return nil
}