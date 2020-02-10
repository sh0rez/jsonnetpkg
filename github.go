package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cavaliercoder/grab"
	"github.com/google/go-github/v29/github"
	"github.com/jsonnet-bundler/jsonnet-bundler/pkg/jsonnetfile"
	"github.com/mholt/archiver"
	"github.com/otiai10/copy"
	"golang.org/x/oauth2"
)

const (
	// user, repo, commit-sha
	uriZipball = "https://codeload.github.com/%s/%s/tar.gz/%s"

	// user, repo, commit-sha, subdir
	uriRawfile = "https://raw.github.com/%s/%s/%s/%s/jsonnetfile.json"
)

type GitHubResolver struct {
	Token string
}

func (gh GitHubResolver) Deps(p Package) (*Pkgfile, error) {
	req, err := http.Get(fmt.Sprintf(uriRawfile, p.User, p.Repo, p.Commit, p.Subdir))
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

type GitHubInstaller struct{}

func (gh GitHubInstaller) Install(p Package, tmp, to string) error {
	fn := filepath.Join(tmp, "pkg.tar.gz")

	// download the archive
	uri := fmt.Sprintf(uriZipball, p.User, p.Repo, p.Commit)
	fmt.Print("GET ", uri)
	res, err := grab.Get(fn, uri)
	if err != nil {
		return err
	}
	fmt.Println("", res.HTTPResponse.StatusCode)

	target := res.HTTPResponse.Header.Get("Content-Disposition")
	target = strings.TrimPrefix(target, "attachment; filename=")
	target = strings.TrimSuffix(target, ".tar.gz")

	// untar it
	extracted := filepath.Join(tmp, "extracted")
	if err := archiver.Extract(fn, target, extracted); err != nil {
		return err
	}

	// move it in place
	loc := filepath.Join(extracted, target)
	return copy.Copy(loc, to)
}
