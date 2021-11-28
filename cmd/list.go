package cmd

import (
	apkg "github.com/innatical/apkg/v3/util"
	"github.com/innatical/pax/v2/util"
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

	if len(db.Packages) == 0 {
		return &apkg.ErrorString{S: "No packages installed"}
	}

	table := make(map[string]string)
	maxWidth := 0

	for _, dbPackage := range db.Packages {
		table[dbPackage.Package.Name+"@"+dbPackage.Package.Version] = dbPackage.Hash

		lineWidth := len(dbPackage.Package.Name) + 1 + len(dbPackage.Package.Version) + 5 + len(dbPackage.Hash)
		if lineWidth > maxWidth {
			maxWidth = lineWidth
		}
	}

	println(apkg.RenderTable(table, maxWidth))

	return nil
}
