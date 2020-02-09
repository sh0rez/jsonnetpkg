package main

var testdataLockfile = Lockfile{
	"github.com/grafana/jsonnet-libs/prometheus-ksonnet@master": Package{
		Host:    "github.com",
		User:    "grafana",
		Repo:    "jsonnet-libs",
		Subdir:  "/prometheus-ksonnet",
		Version: "master",
		Commit:  "7ac7da1a0fe165b68cdb718b2521b560d51bd1f4",
		Deps: map[string]Package{
			"github.com/grafana/jsonnet-libs/ksonnet-util@master": {
				Host:    "github.com",
				User:    "grafana",
				Repo:    "jsonnet-libs",
				Subdir:  "/ksonnet-util",
				Version: "master",
				Deps:    map[string]Package{},
				Commit:  "7ac7da1a0fe165b68cdb718b2521b560d51bd1f4",
			},
			"github.com/grafana/jsonnet-libs/oauth2-proxy@master": {
				Host:    "github.com",
				User:    "grafana",
				Repo:    "jsonnet-libs",
				Subdir:  "/oauth2-proxy",
				Version: "master",
				Deps: map[string]Package{
					"github.com/grafana/jsonnet-libs/ksonnet-util@master": {
						Host:    "github.com",
						User:    "grafana",
						Repo:    "jsonnet-libs",
						Subdir:  "/ksonnet-util",
						Version: "master",
						Deps:    map[string]Package{},
						Commit:  "7ac7da1a0fe165b68cdb718b2521b560d51bd1f4",
					},
				},
				Commit: "7ac7da1a0fe165b68cdb718b2521b560d51bd1f4",
			},
			"github.com/kubernetes-monitoring/kubernetes-mixin@release-0.1": {
				Host:    "github.com",
				User:    "kubernetes-monitoring",
				Repo:    "kubernetes-mixin",
				Subdir:  "",
				Version: "release-0.1",
				Deps: map[string]Package{
					"github.com/grafana/grafonnet-lib/grafonnet@master": {
						Host:    "github.com",
						User:    "grafana",
						Repo:    "grafonnet-lib",
						Subdir:  "/grafonnet",
						Version: "master",
						Deps:    map[string]Package{},
						Commit:  "c459106d2d2b583dd3a83f6c75eb52abee3af764",
					},
					"github.com/kausalco/public/grafana-builder@master": {
						Host:    "github.com",
						User:    "kausalco",
						Repo:    "public",
						Subdir:  "/grafana-builder",
						Version: "master",
						Deps:    map[string]Package{},
						Commit:  "7ac7da1a0fe165b68cdb718b2521b560d51bd1f4",
					},
				},
				Commit: "0ee0da08a9691201a118d30f6b396381c2c8d017",
			},
		},
	},
}
