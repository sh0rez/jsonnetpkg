package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"sort"

	"github.com/google/go-github/v29/github"
	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"gopkg.in/yaml.v3"
)

type Package struct {
	Host   string
	User   string
	Repo   string
	Subdir string

	Version string
	Sum     string

	Dependencies []Package
}

func (p Package) MarshalYAML() (interface{}, error) {
	if len(p.Dependencies) == 0 {
		return p.String(), nil
	}

	sort.Slice(p.Dependencies, func(i int, j int) bool {
		return p.Dependencies[i].String() < p.Dependencies[j].String()
	})

	o := map[string]interface{}{
		p.String(): p.Dependencies,
	}

	return o, nil
}

func (p Package) String() string {
	return fmt.Sprintf("%s@%s", p.Name(), p.Version)
}

func (p Package) Name() string {
	return path.Clean(fmt.Sprintf("%s/%s/%s/%s", p.Host, p.User, p.Repo, p.Subdir))
}

func main() {
	p, err := resolveDependencies("grafana", "jsonnet-libs", "prometheus-ksonnet")
	if err != nil {
		log.Fatalln(err)
	}

	data, err := yaml.Marshal(p)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Print("jsonnet.pkg:\n", string(data), "\n")

	locks := map[string]string{}
	if err := lock(p, locks); err != nil {
		log.Fatalln(err)
	}

	lockData, err := yaml.Marshal(locks)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Print("jsonnet.lock:\n", string(lockData))
}

func resolveDependencies(owner, repo, subdir string) ([]Package, error) {
	r, err := http.Get(fmt.Sprintf("https://raw.github.com/%s/%s/master/%s/jsonnetfile.json", owner, repo, subdir))
	if err != nil {
		return nil, err
	}
	switch r.StatusCode {
	case http.StatusNotFound:
		return []Package{}, nil
	case http.StatusOK:
		break
	default:
		return nil, errors.New(fmt.Sprint(r.StatusCode))
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	jf, err := jsonnetfile.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	pkgs := make([]Package, 0, len(jf.Dependencies))

	results := make(chan []Package)
	errs := make(chan error)
	for _, d := range jf.Dependencies {
		go parallelResolve(d.Source.GitSource.User, d.Source.GitSource.Repo, d.Source.GitSource.Subdir, results, errs)
	}

	for _, d := range jf.Dependencies {
		select {
		case err := <-errs:
			return nil, err
		case transient := <-results:
			pkgs = append(pkgs, Package{
				Host:         d.Source.GitSource.Host,
				User:         d.Source.GitSource.User,
				Repo:         d.Source.GitSource.Repo,
				Subdir:       d.Source.GitSource.Subdir,
				Version:      d.Version,
				Dependencies: transient,
			})
		}
	}

	return pkgs, nil
}

func parallelResolve(owner, repo, subdir string, result chan []Package, e chan error) {
	pkgs, err := resolveDependencies(owner, repo, subdir)
	if err != nil {
		e <- err
		return
	}
	result <- pkgs
	return
}

func lock(pkgs []Package, locks map[string]string) error {
	if len(pkgs) == 0 {
		return nil
	}

	results := make(chan lockResult)
	errs := make(chan error)

	for _, p := range pkgs {
		if _, ok := locks[p.String()]; ok {
			continue
		}

		go resolveVersion(p, results, errs)
	}

	for range pkgs {
		select {
		case err := <-errs:
			return err
		case r := <-results:
			locks[r.key] = r.value
		}
	}
	return nil
}

type lockResult struct {
	key, value string
}

func resolveVersion(p Package, result chan lockResult, e chan error) {
	c := github.NewClient(nil)
	branch, _, err := c.Repositories.GetBranch(context.Background(), p.User, p.Repo, p.Version)
	if err != nil {
		e <- err
		return
	}

	result <- lockResult{
		key:   p.Name(),
		value: branch.Commit.GetSHA(),
	}
	return
}
