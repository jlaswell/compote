package pkg

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLockfile(t *testing.T) {
	tests := map[string]struct {
		filename string
		path     string
	}{
		"correctly implements DependencyFile interface": {
			filename: "composer.lock",
			path:     "../testdata/composer.lock",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				lf  DependencyFile
				err error
			)
			lf, err = newLockfile(tc.path, newLockfileOptions{skipLoading: true})
			assert.Nil(t, err)
			assert.Equal(t, tc.filename, lf.Filename())
			fullpath, err := filepath.Abs(tc.path)
			assert.Nil(t, err)
			assert.Equal(t, fullpath, lf.Fullpath())
			assert.Equal(t, strings.TrimRight(fullpath, tc.filename), lf.Dirpath())
		})
	}
}

func TestLoadFile(t *testing.T) {
	tests := map[string]struct {
		dependencies        []string
		missingDependencies []string
		withDev             bool
		filename            string
		path                string
		passes              bool
		forced              bool
	}{

		"no path": {},
		"non-existent path": {
			path: "../testdata/noWhere",
		},
		"testdata missing file path": {
			path: "../testdata/composer.missing",
		},
		"testdata directory path": {
			filename: "composer.lock",
			path:     "../testdata",
			passes:   true,
		},
		"testdata lockfile path": {
			filename: "composer.lock",
			path:     "../testdata/composer.lock",
			passes:   true,
		},
		"testdata unforced jsonfile path": {
			path:   "../testdata/composer.json",
			passes: false,
		},
		"testdata unforced unique path": {
			path:   "../testdata/composer.unique",
			passes: false,
		},
		"testdata single lockfile dependencies": {
			dependencies: []string{"composer/semver"},
			filename:     "composer.lock",
			path:         "../testdata/installCmd/single/composer.lock",
			passes:       true,
		},
		"testdata multiple lockfile dependencies": {
			dependencies:        []string{"doctrine/dbal", "laravel/framework"},
			missingDependencies: []string{"doctrine/instantiator"},
			filename:            "composer.lock",
			path:                "../testdata/installCmd/multiple/composer.lock",
			passes:              true,
		},
		"testdata multiple lockfile dependencies without dev": {
			dependencies: []string{"doctrine/dbal", "doctrine/instantiator", "laravel/framework"},
			withDev:      true,
			filename:     "composer.lock",
			path:         "../testdata/installCmd/multiple/composer.lock",
			passes:       true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			file, err := LoadFile(tc.path)
			if tc.passes {
				assert.NotNil(t, file)
				assert.Nil(t, err)
				assert.Equal(t, tc.filename, file.Filename())
				foundDependencies := map[string]bool{}
				for _, dependency := range tc.dependencies {
					foundDependencies[dependency] = false
				}
				for _, p := range file.Dependencies(tc.withDev) {
					foundDependencies[p.Name] = true
				}
				for dependency, found := range foundDependencies {
					assert.True(t, found, "Unable to resolve dependency: "+dependency)
				}
			} else {
				assert.Nil(t, file)
				assert.NotNil(t, err)
			}
		})
	}
}
