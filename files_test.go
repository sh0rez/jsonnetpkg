package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPkgfileRemarshal(t *testing.T) {
	pkgfile := Pkgfile{
		Deps: map[string]Package{
			"github.com/grafana/jsonnet-libs/ksonnet-util@master": {
				Host:    "github.com",
				User:    "grafana",
				Repo:    "jsonnet-libs",
				Subdir:  "/ksonnet-util",
				Version: "master",
			},
			"github.com/grafana/jsonnet-libs/oauth2-proxy@master": {
				Host:    "github.com",
				User:    "grafana",
				Repo:    "jsonnet-libs",
				Subdir:  "/oauth2-proxy",
				Version: "master",
			},
		},
	}

	data, err := yaml.Marshal(pkgfile)
	require.NoError(t, err)

	var result Pkgfile
	err = yaml.Unmarshal(data, &result)
	require.NoError(t, err)

	assert.Equal(t, pkgfile, result)
}

func TestLockfileRemarshal(t *testing.T) {
	data, err := yaml.Marshal(testdataLockfile)
	require.NoError(t, err)

	var result Lockfile
	err = yaml.Unmarshal(data, &result)
	require.NoError(t, err)

	if diff := cmp.Diff(testdataLockfile, result); diff != "" {
		t.Error(diff)
	}
}
