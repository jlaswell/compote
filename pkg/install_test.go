package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstall(t *testing.T) {
	tests := map[string]struct {
		fullpath         string
		expectedPackages []string
	}{
		"test a single dependency is installed": {
			fullpath:         "../testdata/installCmd/single/composer.lock",
			expectedPackages: []string{"composer/semver"},
		},
		"test multiple dependencies are installed": {
			fullpath:         "../testdata/installCmd/multiple/composer.lock",
			expectedPackages: []string{"doctrine/cache", "dnoegel/php-xdg-base-dir", "sebastian/version"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var (
				vendors = make(map[string]bool)
				file    DependencyFile
				err     error
			)
			file, err = newLockfile(tc.fullpath)
			assert.Nil(t, err)
			// @todo convert the third to options
			err = InstallFile(file, false, true)
			assert.Nil(t, err, fmt.Sprintf("Check %s for possible undeleted .compote_ directories.", file.Fullpath()))
			err = filepath.Walk(file.Dirpath(), func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					vendorDir := strings.SplitAfter(path, "vendor/")
					if len(vendorDir) == 2 && len(strings.SplitAfter(vendorDir[1], "/")) == 2 {
						vendors[vendorDir[1]] = true
					}
				}
				return err
			})
			assert.Nil(t, err)
			for _, p := range tc.expectedPackages {
				assert.True(t, vendors[p])
			}
		})
	}
}
