package pkg

import (
	"archive/zip"
	"encoding/json"
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

	"github.com/mholt/archiver"
	uuid "github.com/satori/go.uuid"
)

// @todo rename to Install
// @todo convert the third to options
func InstallFile(file DependencyFile, skipDev bool, quiet bool) error {
	var packages = make(map[string]Package)
	pkgs := file.Dependencies(!skipDev)
	for _, p := range pkgs {
		packages[p.Name] = p
	}
	buffer := new(sync.WaitGroup)
	buffer.Add(len(packages))
	if !quiet {
		fmt.Printf("Installing %d direct dependencies\n", len(packages))
	}
	start := time.Now()

	dir, err := ioutil.TempDir(file.Dirpath(), ".compote_")
	if err != nil {
		return err
	}
	for _, p := range packages {
		// @todo rename this to installPackage
		go installPackageRename(buffer, dir, p)
	}
	buffer.Wait()

	vendorDir := filepath.Join(file.Dirpath(), "vendor")
	os.RemoveAll(vendorDir)
	err = os.Rename(dir, vendorDir)
	if err != nil {
		return err
	}
	if !quiet {
		fmt.Printf("\nInstalled %d packages in %s\n", len(packages), time.Since(start))
	}

	// Add the installed.json file for autoloading
	err = os.Mkdir(filepath.Join(vendorDir, "composer"), 0755)
	if !os.IsExist(err) && err != nil {
		return err
	}

	var installedFile *os.File
	installedFile, err = os.Open(filepath.Join(vendorDir, "composer", "installed.json"))
	if !os.IsNotExist(err) && err != nil {
		return err
	} else if os.IsNotExist(err) {
		installedFile, err = os.Create(filepath.Join(vendorDir, "composer", "installed.json"))
	}
	defer installedFile.Close()
	if err != nil {
		return err
	}
	installedFileJson, err := json.MarshalIndent(pkgs, "", "    ")
	if err != nil {
		return err
	}
	fmt.Fprintf(installedFile, string(installedFileJson)+"\n")

	return nil
}

// @todo need a way to handle errors here
func installPackageRename(buffer *sync.WaitGroup, dir string, p Package) error {
	defer buffer.Done()

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
	err = os.MkdirAll(filepath.Join(dir, packageName[0]), os.ModePerm)
	if err != nil {
		return err
	}

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

	// fmt.Print(".")

	return nil
}
