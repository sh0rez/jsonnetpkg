package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
	yaml "gopkg.in/yaml.v3"
)

type Resolver interface {
	// Deps returns the Pkgfile of the package, by querying it from the remote
	Deps(p Package) (*Pkgfile, error)

	// Commit returns the latest commit for the package/version match
	Commit(p Package) (string, error)
}

type Installer interface {
	// Install should install the root of a package to the given path.
	// Installers must not take care of subdirs, etc.
	// They must however extract archives if required.
	Install(p Package, tmp, to string) error
}

func main() {
	log.SetFlags(0)

	file := Pkgfile{
		Deps: map[string]Package{
			"": {
				Host:    "github.com",
				User:    "grafana",
				Repo:    "jsonnet-libs",
				Subdir:  "prometheus-ksonnet",
				Version: "master",
			},
		},
	}

	data, err := yaml.Marshal(file)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Print("jsonnet.pkg:\n", string(data), "\n")

	// ---

	resolver := GitHubResolver{
		Token: os.Getenv("GITHUB_TOKEN"),
	}

	locks := make(Lockfile)
	for _, p := range file.Deps {
		if err := resolve(p, resolver, locks); err != nil {
			log.Fatalln(err)
		}
	}

	data, err = yaml.Marshal(locks)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Print("jsonnet.lock:\n", string(data))

	if err := ensure(locks, "vendor"); err != nil {
		log.Fatalln(err)
	}
}

// resolve takes a Package and adds it with resolved Dependencies to
// `locks`.
func resolve(p Package, r Resolver, locks Lockfile) error {
	// get the commit
	commit, err := r.Commit(p)
	if err != nil {
		return err
	}
	p.Commit = commit

	// get the dependencies
	deps, err := r.Deps(p)
	if err != nil {
		return err
	}

	transient := make(Lockfile)
	for _, d := range deps.Deps {
		if err := resolve(d, r, transient); err != nil {
			return err
		}
	}
	p.Deps = transient

	locks[p.String()] = p
	return nil
}

func ensure(locks Lockfile, dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	installer := GitHubInstaller{}

	for _, p := range locks.Packages() {
		// create tmp dir for this install
		slug := strings.Replace(p.Name(), "/", "-", -1)
		tmp, err := ioutil.TempDir("", "jpkg-"+slug)
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmp)

		// install to tmp location
		installLoc := filepath.Join(tmp, "pkg")
		installTmp := filepath.Join(tmp, "tmp")
		if err := installer.Install(p, installTmp, installLoc); err != nil {
			return err
		}

		// prepare final location and move there
		to := filepath.Join(dir, p.Locked())
		if err := os.MkdirAll(filepath.Dir(to), os.ModePerm); err != nil {
			return err
		}

		pkgDir := filepath.Join(installLoc, p.Subdir)
		if err := copy.Copy(pkgDir, to); err != nil {
			return err
		}
	}

	return nil
}
