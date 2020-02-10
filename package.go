package main

import (
	"fmt"
	"path"
)

type Package struct {
	Host   string `yaml:"host"`
	User   string `yaml:"user"`
	Repo   string `yaml:"repo"`
	Subdir string `yaml:"subdir"`

	Version string `yaml:"version"`

	// lockfile contents
	Deps   map[string]Package `yaml:"-"`
	Sum    string             `yaml:"-"`
	Commit string             `yaml:"-"`
}

func (p Package) String() string {
	return fmt.Sprintf("%s@%s", p.Name(), p.Version)
}

func (p Package) Locked() string {
	return fmt.Sprintf("%s@%s", p.Name(), p.Commit)
}

func (p Package) Name() string {
	return path.Clean(fmt.Sprintf("%s/%s/%s/%s", p.Host, p.User, p.Repo, p.Subdir))
}
