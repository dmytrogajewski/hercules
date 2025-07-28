package test

import (
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/storage/memory"
)

// Repository is a test repository used by tests
var Repository *git.Repository

func init() {
	// Initialize with an empty in-memory repository for tests
	var err error
	Repository, err = git.Init(memory.NewStorage(), nil)
	if err != nil {
		panic(err)
	}
}
