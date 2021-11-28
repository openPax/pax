package cmd

import (
	"strings"

	apkg "github.com/innatical/apkg/v2/util"
	"github.com/innatical/pax/v3/util"
	"github.com/urfave/cli/v2"
)

func Remove(c *cli.Context) error {
	if err := apkg.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer apkg.UnlockDatabase(c.String("root"))

	if err := util.LockDatabase(c.String("root")); err != nil {
		return err
	}

	defer util.UnlockDatabase(c.String("root"))

	for _, name := range c.Args().Slice() {
		dependents, err := util.GetNonOptionalDepdendents(c.String("root"), name)
		if err != nil {
			return err
		}

		if len(dependents) > 0 {
			return &apkg.ErrorString{S: "Errno 2: Package " + name + " Is Required By " + strings.Join(dependents, ", ")}
		}

		if err := util.Remove(c.String("root"), name); err != nil {
			return err
		}
	}

	return nil
}
