package pkg

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/progress"
	"github.com/mholt/archiver"
	uuid "github.com/satori/go.uuid"
)

// @todo switch to a DependencyFile w/ a bool for lockfile

type Lockfile struct {
	Location string            `json:"location"`
	Contents *LockfileContents `json:"contents"`
}

type LockfileContents struct {
	Readme      []string  `json:"_readme,omitempty"`
	Packages    []Package `json:"packages"`
	PackagesDev []Package `json:"packages-dev"`
}

type Package struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Distribution Distribution `json:"dist"`
	Description  string       `json:"description"`
	Autoload     Autoload     `json:"autoload"`
}

type Distribution struct {
	Type      string `json:"type"`
	URL       string `json:"url"`
	Reference string `json:"reference"`
	Shasum    string `json:"shasum"`
}

// @doc https://engineering.bitnami.com/articles/dealing-with-json-with-non-homogeneous-types-in-go.html
type Autoload struct {
	Classmap            []string `json:"classmap,omitempty"`
	ExcludeFromClassmap []string `json:"exclude-from-classmap,omitempty"`
	Files               []string `json:"files,omitempty"`
	PSR0                FlexPSR  `json:"psr-0,omitempty"`
	PSR4                FlexPSR  `json:"psr-4,omitempty"`
}

func (autoload *Autoload) MarshalJSON() ([]byte, error) {
	type singleZeroSingleFour struct {
		Classmap            []string           `json:"classmap,omitempty"`
		ExcludeFromClassmap []string           `json:"exclude-from-classmap,omitempty"`
		Files               []string           `json:"files,omitempty"`
		PSR0                *map[string]string `json:"psr-0,omitempty"`
		PSR4                *map[string]string `json:"psr-4,omitempty"`
	}
	type multipleZeroSingleFour struct {
		Classmap            []string             `json:"classmap,omitempty"`
		ExcludeFromClassmap []string             `json:"exclude-from-classmap,omitempty"`
		Files               []string             `json:"files,omitempty"`
		PSR0                *map[string][]string `json:"psr-0,omitempty"`
		PSR4                *map[string]string   `json:"psr-4,omitempty"`
	}
	type singleZeroMultipleFour struct {
		Classmap            []string             `json:"classmap,omitempty"`
		ExcludeFromClassmap []string             `json:"exclude-from-classmap,omitempty"`
		Files               []string             `json:"files,omitempty"`
		PSR0                *map[string]string   `json:"psr-0,omitempty"`
		PSR4                *map[string][]string `json:"psr-4,omitempty"`
	}
	type multipleZeroMultipleFour struct {
		Classmap            []string             `json:"classmap,omitempty"`
		ExcludeFromClassmap []string             `json:"exclude-from-classmap,omitempty"`
		Files               []string             `json:"files,omitempty"`
		PSR0                *map[string][]string `json:"psr-0,omitempty"`
		PSR4                *map[string][]string `json:"psr-4,omitempty"`
	}
	if autoload.PSR0.Single != nil {
		if autoload.PSR4.Single != nil {
			return json.Marshal(&singleZeroSingleFour{
				Classmap:            autoload.Classmap,
				ExcludeFromClassmap: autoload.ExcludeFromClassmap,
				Files:               autoload.Files,
				PSR0:                autoload.PSR0.Single,
				PSR4:                autoload.PSR4.Single,
			})
		} else {
			return json.Marshal(&singleZeroMultipleFour{
				Classmap:            autoload.Classmap,
				ExcludeFromClassmap: autoload.ExcludeFromClassmap,
				Files:               autoload.Files,
				PSR0:                autoload.PSR0.Single,
				PSR4:                autoload.PSR4.Multiple,
			})
		}
	}
	if autoload.PSR4.Single != nil {
		return json.Marshal(&multipleZeroSingleFour{
			Classmap:            autoload.Classmap,
			ExcludeFromClassmap: autoload.ExcludeFromClassmap,
			Files:               autoload.Files,
			PSR0:                autoload.PSR0.Multiple,
			PSR4:                autoload.PSR4.Single,
		})
	}

	return json.Marshal(&multipleZeroMultipleFour{
		Classmap:            autoload.Classmap,
		ExcludeFromClassmap: autoload.ExcludeFromClassmap,
		Files:               autoload.Files,
		PSR0:                autoload.PSR0.Multiple,
		PSR4:                autoload.PSR4.Multiple,
	})
}

type FlexPSR struct {
	Single   *map[string]string   `json:",omitempty"`
	Multiple *map[string][]string `json:",omitempty"`
}

func (fpsr *FlexPSR) UnmarshalJSON(b []byte) error {
	var (
		single   = map[string]string{}
		multiple = map[string][]string{}
	)

	if bytes.Contains(b, []byte("[")) {
		if err := json.Unmarshal(b, &multiple); err != nil {
			return err
		}
		fpsr.Multiple = &multiple
	} else {
		if err := json.Unmarshal(b, &single); err != nil {
			return err
		}
		fpsr.Single = &single

	}

	return nil
}

// parsePath will attempt to find the path of the composer.lock file moving in priority of:
// 1. if the path is an existing composer.lock file, use that file
// 2: if the path is a directory, look for a composer.lock file in the passed directory
// 3. if the path is a non-existant directory or non-"composer.lock" file, fail
func parsePath(path string) (string, error) {
	var (
		err      error
		info     os.FileInfo
		fullpath string
	)

	fullpath, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if exists, err := fileExist(fullpath); !exists {
		return "", err
	}

	info, err = os.Stat(fullpath)
	if err != nil {
		return "", err
	}
	if info.IsDir() && !strings.HasSuffix(fullpath, "composer.lock") {
		fullpath = filepath.Join(fullpath, "composer.lock")
	}

	if exists, err := fileExist(fullpath); !exists {
		return "", err
	}

	return fullpath, err
}

func fileExist(fullpath string) (exists bool, err error) {
	_, err = os.Stat(fullpath)
	if os.IsNotExist(err) {
		msg := fmt.Sprintf("No valid composer.lock file found at %s", fullpath)
		return false, errors.New(msg)
	}

	return true, nil
}

func removeFilename(fullpath string) string {
	return string(fullpath[:len(fullpath)-13])
}

// LoadLockfile will find and parse a .lock file from the given path.
func LoadLockfile(path string) (*Lockfile, error) {
	var (
		err      error
		fullpath string
	)

	fullpath, err = parsePath(path)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return nil, err
	}
	lockfile := new(Lockfile)
	lockfile.Location = removeFilename(fullpath)
	lockfile.Contents = new(LockfileContents)
	err = json.Unmarshal(contents, lockfile.Contents)
	if err != nil {
		return nil, err
	}

	return lockfile, err
}

type progressTracker struct {
	pw      progress.Writer
	tracker *progress.Tracker
	wg      *sync.WaitGroup
}

func Install(lockfile *Lockfile, withDev bool) error {
	location := (*lockfile).Location
	dir, err := ioutil.TempDir(location, ".compote_")
	if err != nil {
		return err
	}
	// defer os.RemoveAll(dir)

	// @todo make quiet
	// log.Println(viper.GetBool("quiet"))
	pw := progress.NewWriter()
	total := len(lockfile.Contents.Packages)
	if withDev {
		total += len(lockfile.Contents.PackagesDev)
	}
	msg := fmt.Sprintf("installing %d packages", total)
	progTracker := &progressTracker{
		pw:      pw,
		tracker: &progress.Tracker{Message: msg, Total: int64(total)},
		wg:      new(sync.WaitGroup),
	}
	progTracker.wg.Add(total)
	go progTracker.pw.Render()
	progTracker.pw.AppendTracker(progTracker.tracker)
	start := time.Now()

	// @todo remove
	// log.Println(msg)
	for _, p := range lockfile.Contents.Packages {
		// @todo This should push errors into a queue to prevent from breaking
		// the whole process. There should probably be a reasonable error
		// threshold as well as a timeout too.
		go installPackage(progTracker, dir, p)
	}
	if withDev {
		for _, p := range lockfile.Contents.PackagesDev {
			// @todo This should push errors into a queue to prevent from breaking
			// the whole process. There should probably be a reasonable error
			// threshold as well as a timeout too.
			go installPackage(progTracker, dir, p)
		}
	}

	progTracker.wg.Wait()
	progTracker.tracker.MarkAsDone()

	vendorDir := filepath.Join(location, "vendor")
	os.RemoveAll(vendorDir)
	err = os.Rename(dir, vendorDir)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(location)
	fmt.Printf("done in %s", time.Since(start))

	return nil
}

func installPackage(pt *progressTracker, dir string, p Package) error {
	defer pt.wg.Done()

	archive := filepath.Join(dir, uuid.NewV4().String()+".zip")
	out, err := os.Create(archive)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(p.Distribution.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)

	var (
		first    string
		firstSet bool
	)
	// @todo Would be nice to only have to walk the first file to get the dir name.
	err = archiver.Walk(archive, func(f archiver.File) error {
		zfh, ok := f.Header.(zip.FileHeader)
		if ok {
			// fmt.Println("Filename:", zfh.Name)
		}
		if !firstSet {
			first = zfh.Name
			firstSet = true
		}
		return nil
	})
	err = archiver.Unarchive(archive, dir)
	if err != nil {
		return err
	}
	packagePath := filepath.Join(dir, p.Name)
	packageName := strings.Split(p.Name, "/")
	// vendor
	// err = os.Mkdir(filepath.Join(dir, packageName[0]))
	// err = os.MkdirAll(packagePath, os.ModePerm)
	err = os.MkdirAll(filepath.Join(dir, packageName[0]), os.ModePerm)
	if err != nil {
		return err
	}

	// log.Println(filepath.Join(dir, first))
	// log.Println(packageName)
	// log.Println(packagePath)
	err = os.Rename(filepath.Join(dir, first), packagePath)
	if err != nil {
		log.Println(err)
		return err
	}
	err = os.Remove(archive)
	if err != nil {
		log.Println(err)
		return err
	}

	pt.tracker.Increment(int64(1))

	return nil
}
