package util

import (
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	apkg "github.com/innatical/apkg/util"
)

func Install(root string, name string, version string, installOptional bool) error {
	list, err := ReadSourcesList(root)
	if err != nil {
		return err
	}

	sources, err := FetchSourcesList(list)
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
		return &apkg.ErrorString{S: "Could not find package with name " + name}
	}

	v := ""

	if version == "" {
		semver, err := GetLatest(pkg)
		if err != nil {
			return err
		}

		v = semver.String()
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
				v = semver.String()
				break
			}
		}
	}
	

	if v == "" {
		return &apkg.ErrorString{S: "Could not find package " + name + " with version " + version}
	}

	installed, err := IsInstalled(root, name, v)
	if err != nil {
		return err
	}

	if installed {
		return nil
	}

	url := pkg[v]

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	f, err := os.CreateTemp(os.TempDir(), "alt")
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	pkgRoot, err := apkg.InspectPackage(f.Name())
	if err != nil {
		return err
	}

	for _, dep := range pkgRoot.Dependencies.Required {
		parsed := strings.Split(dep, "@")
		if len(parsed) == 1 {
			if err := Install(root, parsed[0], "", installOptional); err != nil {
				return err
			}
		} else {
			if err := Install(root, parsed[0], parsed[1], installOptional); err != nil {
				return err
			}
		}
	}

	if installOptional {
		for _, dep := range pkgRoot.Dependencies.Optional {
			parsed := strings.Split(dep, "@")
			if len(parsed) == 1 {
				if err := Install(root, parsed[0], "", installOptional); err != nil {
					return err
				}
			} else {
				if err := Install(root, parsed[0], parsed[1], installOptional); err != nil {
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
		return false, &apkg.ErrorString{S: "Package " + name  + " installed but does not match version constaint " + constraint}
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

func Upgrade(root, name string, sv bool) error {
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

	list, err := ReadSourcesList(root)
	if err != nil {
		return err
	}


	sources, err := FetchSourcesList(list)
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
		return &apkg.ErrorString{S: "Could not find package with name " + name}
	}

	var versions []*semver.Version

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

		versions = append(versions, ver)

		failed:
	}

	sort.Sort(semver.Collection(versions))

	if len(versions) == 0 {
		return &apkg.ErrorString{S: "Could not find latest version for package  " + name }
	}

	chosen := versions[len(versions) - 1]

	if err := Remove(root, name); err != nil {
		return err
	}

	if err := Install(root, name, chosen.String(), false); err != nil {
		return err
	}

	for _, v := range db.Packages[name].Dependencies.Required {
		parsed := strings.Split(v, "@")
		if err := Upgrade(root, parsed[0], sv); err != nil {
			return err
		}
	}

	for _, v := range db.Packages[name].Dependencies.Optional {
		parsed := strings.Split(v, "@")
		if err := Upgrade(root, parsed[0], sv); err != nil {
			return err
		}
	}

	return nil
}