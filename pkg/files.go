package pkg

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// DependencyFile represents all of the behavior required to manage project
// dependency files.
type DependencyFile interface {
	Filename() string
	Fullpath() string
	Dirpath() string
	Dependencies(withDev bool) []Package
}

var _ DependencyFile = (*lockfile)(nil)

type lockfile struct {
	Packages    []Package `json:"packages"`
	PackagesDev []Package `json:"packages-dev"`
	filename    string
	fullpath    string
}

func (f *lockfile) Filename() string {
	return f.filename
}

func (f *lockfile) Fullpath() string {
	return f.fullpath
}

func (f *lockfile) Dirpath() string {
	return strings.TrimRight(f.fullpath, f.filename)
}

func (f *lockfile) Dependencies(withDev bool) []Package {
	if withDev {
		packages := append(make([]Package, 0), f.Packages...)
		return append(packages, f.PackagesDev...)
	}
	return f.Packages
}

type newLockfileOptions struct {
	skipLoading bool
}

func newLockfile(path string, options ...newLockfileOptions) (*lockfile, error) {
	// Just incase we didn't actually pass in a fullpath.
	fullpath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	// @todo this only works on osx/linux?
	filename := fullpath[strings.LastIndex(fullpath, "/")+1:]

	lf := &lockfile{
		filename: filename,
		fullpath: fullpath,
	}

	// This is present for testing. There is probably a better way to handle this.
	if len(options) > 0 && options[0].skipLoading {
		return lf, nil
	}

	contents, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return nil, err
	}
	var staticFile = struct {
		Location string `json:"location"`
		Contents struct {
			Packages    []Package `json:"packages"`
			PackagesDev []Package `json:"packages-dev"`
		} `json:"contents"`
	}{}
	err = json.Unmarshal(contents, &staticFile.Contents)
	if err != nil {
		return nil, err
	}

	lf.Packages = staticFile.Contents.Packages
	lf.PackagesDev = staticFile.Contents.PackagesDev

	return lf, nil
}

// LoadFile will generate a Dependency file from a given path.
func LoadFile(path string, force bool) (DependencyFile, error) {
	fullpath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(fullpath)
	if err != nil {
		return nil, err
	}

	// @todo This grouping of conditionals will need to be reworked for readability
	// and maintainability. Just working to make this simple and testable at the moment.
	if !info.IsDir() {
		if force {
			file, err := newLockfile(fullpath)
			if err != nil {
				return nil, err
			}
			return file, nil
		} else if strings.HasSuffix(fullpath, "composer.lock") {
			file, err := newLockfile(fullpath)
			if err != nil {
				return nil, err
			}
			return file, nil
		}

		return nil, errors.Errorf("No valid composer.lock file found at %s", fullpath)
	}

	// Find a dependency file from the current directory path.
	exists, err := pathExists(filepath.Join(fullpath, "composer.lock"))
	if exists && err == nil {
		file, err := newLockfile(filepath.Join(fullpath, "composer.lock"))
		if err != nil {
			return nil, err
		}
		return file, nil
	}
	if force {
		exists, err = pathExists(filepath.Join(fullpath, "composer.json"))
		if exists && err == nil {
			file, err := newLockfile(filepath.Join(fullpath, "composer.json"))
			if err != nil {
				return nil, err
			}
			return file, nil
		}
	}

	return nil, errors.Errorf("No valid composer.lock or composer.json file found at %s", fullpath)
}

func pathExists(fullpath string) (bool, error) {
	_, err := os.Stat(fullpath)
	return !os.IsNotExist(err), err
}
