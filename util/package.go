package util

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/charmbracelet/lipgloss"
	"github.com/cheggaaa/pb"
	apkg "github.com/innatical/apkg/v3/util"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type ResolvedPackage struct {
	Version string
	File    string
}

var downloadSemaphore = semaphore.NewWeighted(5)

var barPool = pb.NewPool()

func IsResolved(resolved map[string]ResolvedPackage, name, constraint string) (bool, error) {
	if _, ok := resolved[name]; !ok {
		return false, nil
	}

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, err
	}

	v, err := semver.NewVersion(resolved[name].Version)
	if err != nil {
		return false, err
	}

	if !c.Check(v) {
		return false, &apkg.ErrorString{S: "Errno 6: Package " + name + " is planned to be installed but does not match version constaint " + constraint}
	}

	return true, nil
}

func ResolveNeeded(root string, cache string, sources []Source, pkg string, installOptional bool, state map[string]ResolvedPackage, stateLock *sync.Mutex, group *errgroup.Group) {
	group.Go(func() error {
		parsed := strings.Split(pkg, "@")
		name := parsed[0]
		version := ""

		if len(parsed) == 2 {
			version = parsed[1]
		}

		var pkg map[string]string

		for _, source := range sources {
			if _, ok := source.Packages[name]; ok {
				pkg = source.Packages[name]
				break
			}
		}

		if pkg == nil {
			return &apkg.ErrorString{S: "Errno 4: Could not find package with name " + name}
		}

		v := ""

		if version == "" {
			v = GetLatest(pkg)
		} else {
			constraint, err := semver.NewConstraint(version)
			if err != nil {
				return err
			}

			for ver := range pkg {
				semver, err := semver.NewVersion(ver)
				if err != nil {
					return err
				}

				if constraint.Check(semver) {
					v = ver
					break
				}
			}
		}

		if v == "" {
			return &apkg.ErrorString{S: "Errno 5: Could not find package " + name + " with version " + version}
		}

		stateLock.Lock()
		resolved, err := IsResolved(state, name, v)
		if err != nil {
			stateLock.Unlock()
			return err
		}

		stateLock.Unlock()

		if resolved {
			return nil
		}

		installed, err := IsInstalled(root, name, v)
		if err != nil {
			return err
		}

		if installed {
			return nil
		}

		url := pkg[v]

		var f *os.File

		var fName = filepath.Join(cache, name+version+".apkg")

		var sourceSize int64

		if err := func() error {
			// TODO: It's 12am and I forgot what context does again
			downloadSemaphore.Acquire(context.Background(), 1)
			defer downloadSemaphore.Release(1)

			println("Downloading " + lipgloss.NewStyle().Bold(true).Render(name))

			resp, err := http.Get(url)
			if err != nil {
				return err
			}

			defer resp.Body.Close()

			i, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

			sourceSize = int64(i)

			if err := os.MkdirAll(cache, 0755); err != nil {
				return err
			}

			if _, err := os.Stat(fName); err != nil {
				f, err = os.OpenFile(fName, os.O_RDWR|os.O_CREATE, 0755)
				if err != nil {
					return err
				}

				bar := pb.New(int(sourceSize)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10).Prefix(lipgloss.NewStyle().Bold(true).Render(name))
				bar.ShowSpeed = true
				barPool.Add(bar)

				reader := bar.NewProxyReader(resp.Body)

				if _, err := io.Copy(f, reader); err != nil {
					return err
				}

				bar.Finish()
			} else {
				println("File exists, not redownloading")
				f, err = os.OpenFile(fName, os.O_RDWR|os.O_CREATE, 0755)
				if err != nil {
					return err
				}
				return err
			}

			return nil
		}(); err != nil {
			return err
		}

		pkgRoot, err := apkg.InspectPackage(f.Name())
		if err != nil {
			return err
		}

		for _, dep := range pkgRoot.Dependencies.Required {
			ResolveNeeded(root, cache, sources, dep, installOptional, state, stateLock, group)
		}

		if installOptional {
			for _, dep := range pkgRoot.Dependencies.Optional {
				ResolveNeeded(root, cache, sources, dep, installOptional, state, stateLock, group)
			}
		}

		stateLock.Lock()
		defer stateLock.Unlock()

		resolved, err = IsResolved(state, name, v)
		if err != nil {
			return err
		}

		if resolved {
			return nil
		}

		state[name] = ResolvedPackage{
			Version: v,
			File:    f.Name(),
		}

		return nil
	})
}

func InstallMultiple(root string, cache string, packages []string, installOptional bool) error {
	list, err := ReadReposList(root)
	if err != nil {
		return err
	}

	sources, err := FetchSourcesList(list.Repos)
	if err != nil {
		return err
	}

	resolved := make(map[string]ResolvedPackage)
	group := new(errgroup.Group)
	var resolvedLock sync.Mutex

	for _, pkg := range packages {
		pkg := pkg

		ResolveNeeded(root, cache, sources, pkg, installOptional, resolved, &resolvedLock, group)
		barPool.Start()
	}
	barPool.Stop()

	if err := group.Wait(); err != nil {
		resolvedLock.Lock()
		for _, res := range resolved {
			os.Remove(res.File)
		}
		resolvedLock.Unlock()

		return err
	}

	var files []string

	for _, res := range resolved {
		files = append(files, res.File)
	}

	if err := apkg.InstallMultiple(root, files); err != nil {
		return err
	}

	return nil
}

func Install(root string, cache string, name string, version string, installOptional bool) error {
	list, err := ReadReposList(root)
	if err != nil {
		return err
	}

	sources, err := FetchSourcesList(list.Repos)
	if err != nil {
		return err
	}

	var pkg map[string]string

	for _, source := range sources {
		if _, ok := source.Packages[name]; ok {
			pkg = source.Packages[name]
			break
		}
	}

	if pkg == nil {
		return &apkg.ErrorString{S: "Errno 4: Could not find package with name " + name}
	}

	v := ""

	if version == "" {
		v = GetLatest(pkg)
	} else {
		constraint, err := semver.NewConstraint(version)
		if err != nil {
			return err
		}

		for ver := range pkg {
			semver, err := semver.NewVersion(ver)
			if err != nil {
				return err
			}

			if constraint.Check(semver) {
				v = ver
				break
			}
		}
	}

	if v == "" {
		return &apkg.ErrorString{S: "Errno 5: Could not find package " + name + " with version " + version}
	}

	installed, err := IsInstalled(root, name, v)
	if err != nil {
		return err
	}

	if installed {
		return nil
	}

	url := pkg[v]

	var f *os.File

	var sourceSize int64

	if err := func() error {
		// TODO: It's 12am and I forgot what context does again
		downloadSemaphore.Acquire(context.Background(), 1)
		defer downloadSemaphore.Release(1)

		println("Downloading " + lipgloss.NewStyle().Bold(true).Render(name))

		resp, err := http.Get(url)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		i, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

		sourceSize = int64(i)

		var fName = filepath.Join(cache, name+version+".apkg")

		if err := os.MkdirAll(cache, 0755); err != nil {
			return err
		}

		if _, err := os.Stat(fName); err != nil {
			f, err = os.OpenFile(fName, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				return err
			}

			bar := pb.New(int(sourceSize)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10).Prefix(lipgloss.NewStyle().Bold(true).Render(name))
			bar.ShowSpeed = true
			barPool.Add(bar)

			reader := bar.NewProxyReader(resp.Body)

			if _, err := io.Copy(f, reader); err != nil {
				return err
			}

			bar.Finish()
		} else {
			println("File exists, not redownloading")
			f, err = os.OpenFile(fName, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				return err
			}
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	pkgRoot, err := apkg.InspectPackage(f.Name())
	if err != nil {
		return err
	}

	for _, dep := range pkgRoot.Dependencies.Required {
		parsed := strings.Split(dep, "@")
		if len(parsed) == 1 {
			if err := Install(root, cache, parsed[0], "", installOptional); err != nil {
				return err
			}
		} else {
			if err := Install(root, cache, parsed[0], parsed[1], installOptional); err != nil {
				return err
			}
		}
	}

	if installOptional {
		for _, dep := range pkgRoot.Dependencies.Optional {
			parsed := strings.Split(dep, "@")
			if len(parsed) == 1 {
				if err := Install(root, cache, parsed[0], "", installOptional); err != nil {
					return err
				}
			} else {
				if err := Install(root, cache, parsed[0], parsed[1], installOptional); err != nil {
					return err
				}
			}
		}
	}

	if err := apkg.Install(root, f.Name()); err != nil {
		return err
	}

	return nil
}

func IsInstalled(root, name, constraint string) (bool, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return false, err
	}

	db, err := apkg.ReadDatabase(root)

	if err != nil {
		return false, err
	}

	if _, ok := db.Packages[name]; !ok {
		return false, nil
	}

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, err
	}

	v, err := semver.NewVersion(db.Packages[name].Package.Version)
	if err != nil {
		return false, err
	}

	if !c.Check(v) {
		return false, &apkg.ErrorString{S: "Errno 6: Package " + name + " installed but does not match version constaint " + constraint}
	}

	return true, nil
}

func IsInstalledName(root, name string) (bool, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return false, err
	}

	db, err := apkg.ReadDatabase(root)

	if err != nil {
		return false, err
	}

	if _, ok := db.Packages[name]; !ok {
		return false, nil
	}

	return true, nil
}

func GetDepdendents(root, name string) ([]string, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}

	db, err := apkg.ReadDatabase(root)

	if err != nil {
		return nil, err
	}

	var packages []string

	for pkgName, pkg := range db.Packages {
		for _, value := range pkg.Dependencies.Required {
			parsed := strings.Split(value, "@")
			if parsed[0] == name {
				packages = append(packages, pkgName)
			}
		}

		for _, value := range pkg.Dependencies.Optional {
			parsed := strings.Split(value, "@")
			if parsed[0] == name {
				packages = append(packages, pkgName)
			}
		}
	}

	return packages, nil
}

func GetNonOptionalDepdendents(root, name string) ([]string, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}

	db, err := apkg.ReadDatabase(root)

	if err != nil {
		return nil, err
	}

	var packages []string

	for pkgName, pkg := range db.Packages {
		for _, value := range pkg.Dependencies.Required {
			parsed := strings.Split(value, "@")
			if parsed[0] == name {
				packages = append(packages, pkgName)
			}
		}
	}

	return packages, nil
}

func Remove(root, name string) error {
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}

	db, err := apkg.ReadDatabase(root)

	if err != nil {
		return err
	}

	if err := apkg.Remove(root, name); err != nil {
		return err
	}

	for _, v := range db.Packages[name].Dependencies.Required {
		parsed := strings.Split(v, "@")
		dependents, err := GetDepdendents(root, parsed[0])
		if err != nil {
			return err
		}

		if len(dependents) == 0 {
			if err := Remove(root, parsed[0]); err != nil {
				return err
			}
		}
	}

	for _, v := range db.Packages[name].Dependencies.Optional {
		parsed := strings.Split(v, "@")
		dependents, err := GetDepdendents(root, parsed[0])
		if err != nil {
			return err
		}

		if len(dependents) == 0 {
			if err := Remove(root, parsed[0]); err != nil {
				return err
			}
		}
	}

	return nil
}

func Upgrade(root, cache string, name string, sv bool) error {
	db, err := apkg.ReadDatabase(root)

	if err != nil {
		return err
	}

	var constaints []*semver.Constraints

	for _, pkg := range db.Packages {
		for _, value := range pkg.Dependencies.Required {
			parsed := strings.Split(value, "@")
			if parsed[0] == name {
				ver, err := semver.NewConstraint(parsed[1])

				if err != nil {
					return err
				}

				constaints = append(constaints, ver)
			}
		}

		for _, value := range pkg.Dependencies.Optional {
			parsed := strings.Split(value, "@")
			if parsed[0] == name {
				ver, err := semver.NewConstraint(parsed[1])

				if err != nil {
					return err
				}

				constaints = append(constaints, ver)
			}
		}
	}

	if sv {
		ver, err := semver.NewConstraint("^" + db.Packages[name].Package.Version)
		if err != nil {
			return err
		}

		constaints = append(constaints, ver)
	}

	list, err := ReadReposList(root)
	if err != nil {
		return err
	}

	sources, err := FetchSourcesList(list.Repos)

	if err != nil {
		return err
	}

	var pkg map[string]string

	for _, source := range sources {
		if _, ok := source.Packages[name]; ok {
			pkg = source.Packages[name]
			break
		}
	}

	if pkg == nil {
		return &apkg.ErrorString{S: "Errno 4: Could not find package with name " + name}
	}

	var versions []string

	for k := range pkg {
		ver, err := semver.NewVersion(k)
		if err != nil {
			return err
		}

		for _, c := range constaints {
			if !c.Check(ver) {
				goto failed
			}
		}

		versions = append(versions, k)

	failed:
	}

	sort.Sort(BySemver(versions))

	if len(versions) == 0 {
		return &apkg.ErrorString{S: "Errno 7: Could not find latest version for package  " + name}
	}

	chosen := versions[len(versions)-1]

	if chosen != db.Packages[name].Package.Version {
		if err := apkg.Remove(root, name); err != nil {
			return err
		}

		if err := Install(root, cache, name, chosen, false); err != nil {
			return err
		}
	}

	for _, v := range db.Packages[name].Dependencies.Required {
		parsed := strings.Split(v, "@")
		if err := Upgrade(root, cache, parsed[0], sv); err != nil {
			return err
		}
	}

	for _, v := range db.Packages[name].Dependencies.Optional {
		parsed := strings.Split(v, "@")
		if err := Upgrade(root, cache, parsed[0], sv); err != nil {
			return err
		}
	}

	return nil
}
