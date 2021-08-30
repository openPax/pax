package util

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	apkg "github.com/innatical/apkg/util"
)

type Database struct {
	Packages []string `toml:"packges"`
}

func ReadDatabase(root string) (*Database, error) {
	_, err := os.Stat(filepath.Join(root, "alt.toml"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		} else {
			file, err := os.Create(filepath.Join(root, "alt.toml"))
			if err != nil {
				return nil, err
			}

			file.Close()
		}
	}

	var db Database

	if _, err := toml.DecodeFile(filepath.Join(root, "alt.toml"), &db); err != nil {
		return nil, err
	}


	return &db, nil
}

func WriteDatabase(root string, db *Database) error {
	file, err := os.OpenFile(filepath.Join(root, "alt.toml"), os.O_TRUNC|os.O_WRONLY, 0644)
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
	_, err := os.Stat(filepath.Join(root, "alt.lock"))
	if err == nil {
		return &apkg.ErrorString{S: "Database already locked"}
	}

	if !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(filepath.Join(root, "alt.lock"))
	if err != nil {
		return err
	}

	defer file.Close()

	return nil
}

func UnlockDatabase(root string) error {
	if err := os.Remove(filepath.Join(root, "alt.lock")); err != nil {
		return nil
	}

	return nil
}
