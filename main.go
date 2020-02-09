package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Resolver interface {
	Deps(p Package) (*Pkgfile, error)
	Commit(p Package) (string, error)
}

type Installer interface {
	Install(p Package, to string) error
}

func main() {
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
