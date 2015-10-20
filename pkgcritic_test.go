package main

import (
	"testing"

	"github.com/bom-d-van/goutils/printutils"
	"github.com/rainycape/gondola/social/github"
)

func TestHierarchise(t *testing.T) {
	hierarchise([]Critique{
		{
			Path:           "github.com/original/csfw/storage/money",
			GitHubFullName: "original/csfw",
			forks: []github.Repository{
				FullName: "bom-d-van/csfw",
			},
		},
		{
			Path:           "github.com/bom-d-van/csfw/storage/money",
			GitHubFullName: "bom-d-van/csfw",
			forks:          []github.Repository{},
		},
	})

	printutils.PrettyPrint()
}
