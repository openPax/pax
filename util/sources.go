package util

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver"
)

func ReadSourcesList(root string) ([]string, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}

	_, err := os.Stat(filepath.Join(root, "altsources.list"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		} else {
			file, err := os.Create(filepath.Join(root, "altsources.list"))
			if err != nil {
				return nil, err
			}

			file.Close()
		}
	}

	file, err := os.Open(filepath.Join(root, "altsources.list"))

	if err != nil {
		return nil, err
	}

	
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)
	var sources []string

	for scanner.Scan() {
		sources = append(sources, scanner.Text())
	}

	return sources, nil
}

type Source struct {
	Packages map[string]map[string]string `toml:"packages"`
}

func FetchSourcesList(list []string) ([]Source, error) {
	var sources []Source

	for i := range list {
		resp, err := http.Get(list[i])

		if err != nil {
			return nil, err
		}

		var source Source

		if _, err := toml.DecodeReader(resp.Body, &source); err != nil {
			return nil, err
		}

		sources = append(sources, source)
	}

	return sources, nil
}

func GetLatest(versions map[string]string) (*semver.Version, error) {
	var vs []*semver.Version

	for s := range versions {
		version, err := semver.NewVersion(s)
		if err != nil {
			return nil, err
		}

		vs = append(vs, version)
	}

	sort.Sort(semver.Collection(vs))

	return vs[len(vs)-1], nil
}

func Search(root string, query string) ([]string, error) {
	list, err := ReadSourcesList(root)
	if err != nil {
		return nil, err
	}

	sourcesList, err := FetchSourcesList(list)
	if err != nil {
		return nil, err
	}

	var found []string

	for i := range sourcesList {
		for s := range sourcesList[i].Packages {
			if strings.Contains(s, query) {
				latest, err := GetLatest(sourcesList[i].Packages[s])
				if err != nil {
					return nil, err
				}

				found = append(found, s+"@"+latest.String())
			}
		}
	}

	return found, nil
}
