package util

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	apkg "github.com/innatical/apkg/v3/util"
)

type Database struct {
	Packages []string `toml:"packages"`
}

func ReadDatabase(root string) (*Database, error) {
	_, err := os.Stat(filepath.Join(root, "pax.toml"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		} else {
			file, err := os.Create(filepath.Join(root, "pax.toml"))
			if err != nil {
				return nil, err
			}

			file.Close()
		}
	}

	var db Database

	if _, err := toml.DecodeFile(filepath.Join(root, "pax.toml"), &db); err != nil {
		return nil, err
	}

	return &db, nil
}

func WriteDatabase(root string, db *Database) error {
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(filepath.Join(root, "pax.toml"), os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	if err := toml.NewEncoder(file).Encode(db); err != nil {
		return err
	}

	return nil
}

func LockDatabase(root string) error {
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}

	_, err := os.Stat(filepath.Join(root, "pax.lock"))
	if err == nil {
		return &apkg.ErrorString{S: "Database already locked"}
	}

	if !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(filepath.Join(root, "pax.lock"))
	if err != nil {
		return err
	}

	defer file.Close()

	return nil
}

func UnlockDatabase(root string) error {
	if err := os.Remove(filepath.Join(root, "pax.lock")); err != nil {
		return nil
	}

	return nil
}
