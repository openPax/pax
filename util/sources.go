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

type ReposList struct {
	Repos map[string]string `toml:"repos"`
}

func ReadReposList(root string) (*ReposList, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}

	_, err := os.Stat(filepath.Join(root, "repos.toml"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		} else {
			file, err := os.Create(filepath.Join(root, "repos.toml"))
			if err != nil {
				return nil, err
			}

			file.Close()
		}
	}

	var repos ReposList

	if _, err := toml.DecodeFile(filepath.Join(root, "repos.toml"), &repos); err != nil {
		return nil, err
	}

	if repos.Repos == nil {
		repos.Repos = make(map[string]string)
	}

	return &repos, nil
}

func UpdateSourcesList(root string, source string) error {
	println("Adding " + source + " to sources list")

	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}

	_, err := os.Stat(filepath.Join(root, "paxsources.list"))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		} else {
			file, err := os.Create(filepath.Join(root, "paxsources.list"))
			if err != nil {
				return err
			}

			file.Close()
		}
	}

	file, err := os.OpenFile(filepath.Join(root, "paxsources.list"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	if _, err := file.WriteString(source); err != nil {
		return err
		//return &util.ErrorString{S: "Error Saving File"}
	}

	defer file.Close()

	return nil
}

func ReadSourcesList(root string) ([]string, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}

	_, err := os.Stat(filepath.Join(root, "paxsources.list"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		} else {
			file, err := os.Create(filepath.Join(root, "paxsources.list"))
			if err != nil {
				return nil, err
			}

			file.Close()
		}
	}

	file, err := os.Open(filepath.Join(root, "paxsources.list"))

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

type BySemver []string

func (s BySemver) Len() int {
    return len(s)
}

func (s BySemver) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}

func (s BySemver) Less(i, j int) bool {
	iVersion, err := semver.NewVersion(s[i])
	if err != nil {
		return false
	}

	jVersion, err := semver.NewVersion(s[j])
	if err != nil {
		return false
	}

    return iVersion.LessThan(jVersion)
}

func GetLatest(versions map[string]string) string {
	var vs []string

	for s := range versions {
		vs = append(vs, s)
	}

	sort.Sort(BySemver(vs))

	return vs[len(vs) - 1]
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
				latest := GetLatest(sourcesList[i].Packages[s])

				found = append(found, s + "@" + latest)
			}
		}
	}

	return found, nil
}
