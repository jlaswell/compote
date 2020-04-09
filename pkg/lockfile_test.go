package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLockfileLoading(t *testing.T) {
	tests := map[string]struct {
		path   string
		passes bool
	}{
		"no path": {},
		"non-existent path": {
			path: "../testdata/noWhere",
		},
		"testdata directory path": {
			path:   "../testdata",
			passes: true,
		},
		"testdata filename path": {
			path:   "../testdata/composer.lock",
			passes: true,
		},
		"testdata bad filename path": {
			path: "../testdata/composer.not-lock",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			lockfile, err := LoadLockfile(tc.path)
			if tc.passes {
				assert.NotNil(t, lockfile)
				// assert.NotEmpty(t, lockfile.Contents)
				var semvarIndex int
				for i, p := range lockfile.Contents.Packages {
					if p.Name == "composer/semver" {
						semvarIndex = i
					}
				}
				assert.Equal(t, "composer/semver", lockfile.Contents.Packages[semvarIndex].Name)
				assert.Equal(t, "1.5.0", lockfile.Contents.Packages[semvarIndex].Version)
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "No valid composer.lock file found")
			}
		})
	}
}
