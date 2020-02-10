package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jsonnet-bundler/jsonnet-bundler/spec/deps"
)

// Pkgfile records all direct dependencies, without any version locks, etc.
// It essentially states what's required, without being interested in details.
type Pkgfile struct {
	Deps map[string]Package `yaml:"dependencies"`
}

func (p Pkgfile) MarshalYAML() (interface{}, error) {
	pkgs := []string{}
	for _, pkg := range p.Deps {
		pkgs = append(pkgs, pkg.String())
	}
	return pkgs, nil
}

func (p *Pkgfile) UnmarshalYAML(unmarshal func(interface{}) error) error {
	strs := []string{}
	if err := unmarshal(&strs); err != nil {
		return err
	}

	pkgs := make(map[string]Package)

	for _, s := range strs {
		d := deps.Parse("", s)

		pkg := Package{
			Host:    d.Source.GitSource.Host,
			User:    d.Source.GitSource.User,
			Repo:    d.Source.GitSource.Repo,
			Subdir:  d.Source.GitSource.Subdir,
			Version: d.Version,
		}

		pkgs[pkg.String()] = pkg
	}

	p.Deps = pkgs
	return nil
}

// Lockfile holds the entire dependency graph, including absolute versions and
// content checksums
type Lockfile map[string]Package

const (
	LockFormat       = "%s commit:%s"
	LockPrefixCommit = "commit:"
	LockPrefixSum    = "sum:"
)

// Packages returns the flat list of packages
func (l Lockfile) Packages() map[string]Package {
	pkgs := make(map[string]Package)
	for _, p := range l {
		pkgs[p.Locked()] = p

		deps := Lockfile(p.Deps)
		for _, d := range deps.Packages() {
			pkgs[d.Locked()] = d
		}
	}

	return pkgs
}

func (l Lockfile) MarshalYAML() (interface{}, error) {
	locks := []interface{}{}

	for _, p := range l {
		locks = append(locks, renderLock(p))
	}

	return locks, nil
}

func lockStr(p Package) string {
	return fmt.Sprintf(LockFormat, p.String(), p.Commit)
}

func renderLock(p Package) interface{} {
	if len(p.Deps) == 0 {
		return lockStr(p)
	}

	transient := []interface{}{}
	for _, d := range p.Deps {
		transient = append(transient, renderLock(d))
	}

	return map[string]interface{}{
		lockStr(p): transient,
	}
}

func (l *Lockfile) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var ds interface{}
	if err := unmarshal(&ds); err != nil {
		return err
	}

	if *l == nil {
		*l = make(map[string]Package)
	}

	return walkLock(ds, *l)
}

func walkLock(ptr interface{}, locks Lockfile) error {
	switch t := ptr.(type) {
	case []interface{}:
		for _, p := range t {
			walkLock(p, locks)
		}

	// A lockStr as a key, so a package with dependencies. Example:
	//
	// github.com/prometheus/prometheus/documentation/prometheus-mixin@master commit:0a8acb654e8c040c8e67c6df41d85cf93b56df29:
	//   - github.com/grafana/jsonnet-libs/grafana-builder@master commit:7ac7da1a0fe165b68cdb718b2521b560d51bd1f4
	//   - github.com/grafana/grafonnet-lib/grafonnet@master commit:c459106d2d2b583dd3a83f6c75eb52abee3af764
	case map[string]interface{}:
		for k, v := range t {
			pkg := parseLock(k)

			transient := make(map[string]Package)
			walkLock(v, transient)
			pkg.Deps = transient

			locks[pkg.String()] = pkg
		}

	// A flat lockStr, so a package without dependencies. Example:
	//
	// - github.com/grafana/jsonnet-libs/grafana-builder@master commit:7ac7da1a0fe165b68cdb718b2521b560d51bd1f4
	case string:
		pkg := parseLock(t)
		locks[pkg.String()] = pkg

	// Something that should not have happened
	default:
		return fmt.Errorf("unexpected type: %T", t)
	}
	return nil
}

var space = regexp.MustCompile(`\s+`)

func parseLock(str string) Package {
	str = space.ReplaceAllString(str, " ")
	elems := strings.Split(str, " ")

	d := deps.Parse("", elems[0])
	pkg := Package{
		Host:    d.Source.GitSource.Host,
		User:    d.Source.GitSource.User,
		Repo:    d.Source.GitSource.Repo,
		Subdir:  addPrefix(d.Source.GitSource.Subdir, "/"),
		Deps:    make(map[string]Package),
		Version: d.Version,
	}

	// parse commit and sum
	for _, s := range elems[1:] {
		switch {
		case strings.HasPrefix(s, LockPrefixCommit):
			pkg.Commit = strings.TrimPrefix(s, LockPrefixCommit)
		case strings.HasPrefix(s, LockPrefixSum):
			pkg.Sum = strings.TrimPrefix(s, LockPrefixSum)
		}
	}

	return pkg
}

func addPrefix(s, prefix string) string {
	if s == "" {
		return s
	}
	if strings.HasPrefix(s, prefix) {
		return s
	}
	return prefix + s
}
