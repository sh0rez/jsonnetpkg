package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v29/github"
	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"golang.org/x/oauth2"
)

type GitHubResolver struct {
	Token string
}

func (gh GitHubResolver) Deps(p Package) (*Pkgfile, error) {
	req, err := http.Get(fmt.Sprintf("https://raw.github.com/%s/%s/%s/%s/jsonnetfile.json", p.User, p.Repo, p.Commit, p.Subdir))
	if err != nil {
		return nil, err
	}
	switch req.StatusCode {
	case http.StatusNotFound:
		// not found: no transient deps
		return &Pkgfile{}, nil
	case http.StatusOK:
		break
	default:
		return nil, errors.New(fmt.Sprint(req.StatusCode))
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	jf, err := jsonnetfile.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	pkgfile := Pkgfile{
		Deps: make(map[string]Package),
	}

	for _, d := range jf.Dependencies {
		pkg := Package{
			User:    d.Source.GitSource.User,
			Repo:    d.Source.GitSource.Repo,
			Subdir:  d.Source.GitSource.Subdir,
			Host:    d.Source.GitSource.Host,
			Version: d.Version,
		}
		pkgfile.Deps[pkg.String()] = pkg
	}
	return &pkgfile, nil
}

func (gh GitHubResolver) Commit(p Package) (string, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gh.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	c := github.NewClient(tc)

	branch, _, err := c.Repositories.GetBranch(ctx, p.User, p.Repo, p.Version)
	if err != nil {
		return "", err
	}

	return branch.Commit.GetSHA(), nil
}
