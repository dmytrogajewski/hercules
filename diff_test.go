package hercules

import (
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"unicode/utf8"
)

func fixtureFileDiff() *FileDiff {
	fd := &FileDiff{}
	fd.Initialize(testRepository)
	return fd
}

func TestFileDiffMeta(t *testing.T) {
	fd := fixtureFileDiff()
	assert.Equal(t, fd.Name(), "FileDiff")
	assert.Equal(t, len(fd.Provides()), 1)
	assert.Equal(t, fd.Provides()[0], DependencyFileDiff)
	assert.Equal(t, len(fd.Requires()), 2)
	assert.Equal(t, fd.Requires()[0], DependencyTreeChanges)
	assert.Equal(t, fd.Requires()[1], DependencyBlobCache)
	assert.Len(t, fd.ListConfigurationOptions(), 0)
	fd.Configure(nil)
}

func TestFileDiffRegistration(t *testing.T) {
	tp, exists := Registry.registered[(&FileDiff{}).Name()]
	assert.True(t, exists)
	assert.Equal(t, tp.Elem().Name(), "FileDiff")
	tps, exists := Registry.provided[(&FileDiff{}).Provides()[0]]
	assert.True(t, exists)
	assert.True(t, len(tps) >= 1)
	matched := false
	for _, tp := range tps {
		matched = matched || tp.Elem().Name() == "FileDiff"
	}
	assert.True(t, matched)
}

func TestFileDiffConsume(t *testing.T) {
	fd := fixtureFileDiff()
	deps := map[string]interface{}{}
	cache := map[plumbing.Hash]*object.Blob{}
	hash := plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe")
	cache[hash], _ = testRepository.BlobObject(hash)
	hash = plumbing.NewHash("334cde09da4afcb74f8d2b3e6fd6cce61228b485")
	cache[hash], _ = testRepository.BlobObject(hash)
	hash = plumbing.NewHash("dc248ba2b22048cc730c571a748e8ffcf7085ab9")
	cache[hash], _ = testRepository.BlobObject(hash)
	deps[DependencyBlobCache] = cache
	changes := make(object.Changes, 3)
	treeFrom, _ := testRepository.TreeObject(plumbing.NewHash(
		"a1eb2ea76eb7f9bfbde9b243861474421000eb96"))
	treeTo, _ := testRepository.TreeObject(plumbing.NewHash(
		"994eac1cd07235bb9815e547a75c84265dea00f5"))
	changes[0] = &object.Change{From: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeFrom,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("dc248ba2b22048cc730c571a748e8ffcf7085ab9"),
		},
	}, To: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeTo,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("334cde09da4afcb74f8d2b3e6fd6cce61228b485"),
		},
	}}
	changes[1] = &object.Change{From: object.ChangeEntry{}, To: object.ChangeEntry{
		Name: ".travis.yml",
		Tree: treeTo,
		TreeEntry: object.TreeEntry{
			Name: ".travis.yml",
			Mode: 0100644,
			Hash: plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe"),
		},
	},
	}
	changes[2] = &object.Change{From: object.ChangeEntry{
		Name: "rbtree.go",
		Tree: treeFrom,
		TreeEntry: object.TreeEntry{
			Name: "rbtree.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("14c3fa5a1cca103032f10379467a3a2f210e5f94"),
		},
	}, To: object.ChangeEntry{},
	}
	deps[DependencyTreeChanges] = changes
	res, err := fd.Consume(deps)
	assert.Nil(t, err)
	diffs := res[DependencyFileDiff].(map[string]FileDiffData)
	assert.Equal(t, len(diffs), 1)
	diff := diffs["analyser.go"]
	assert.Equal(t, diff.OldLinesOfCode, 307)
	assert.Equal(t, diff.NewLinesOfCode, 309)
	deletions := 0
	insertions := 0
	for _, edit := range diff.Diffs {
		switch edit.Type {
		case diffmatchpatch.DiffEqual:
			continue
		case diffmatchpatch.DiffInsert:
			insertions += utf8.RuneCountInString(edit.Text)
		case diffmatchpatch.DiffDelete:
			deletions += utf8.RuneCountInString(edit.Text)
		}
	}
	assert.Equal(t, deletions, 13)
	assert.Equal(t, insertions, 15)
}

func TestFileDiffConsumeInvalidBlob(t *testing.T) {
	fd := fixtureFileDiff()
	deps := map[string]interface{}{}
	cache := map[plumbing.Hash]*object.Blob{}
	hash := plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe")
	cache[hash], _ = testRepository.BlobObject(hash)
	hash = plumbing.NewHash("334cde09da4afcb74f8d2b3e6fd6cce61228b485")
	cache[hash], _ = testRepository.BlobObject(hash)
	hash = plumbing.NewHash("dc248ba2b22048cc730c571a748e8ffcf7085ab9")
	cache[hash], _ = testRepository.BlobObject(hash)
	deps[DependencyBlobCache] = cache
	changes := make(object.Changes, 1)
	treeFrom, _ := testRepository.TreeObject(plumbing.NewHash(
		"a1eb2ea76eb7f9bfbde9b243861474421000eb96"))
	treeTo, _ := testRepository.TreeObject(plumbing.NewHash(
		"994eac1cd07235bb9815e547a75c84265dea00f5"))
	changes[0] = &object.Change{From: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeFrom,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("ffffffffffffffffffffffffffffffffffffffff"),
		},
	}, To: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeTo,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("334cde09da4afcb74f8d2b3e6fd6cce61228b485"),
		},
	}}
	deps[DependencyTreeChanges] = changes
	res, err := fd.Consume(deps)
	assert.Nil(t, res)
	assert.NotNil(t, err)
	changes[0] = &object.Change{From: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeFrom,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("dc248ba2b22048cc730c571a748e8ffcf7085ab9"),
		},
	}, To: object.ChangeEntry{
		Name: "analyser.go",
		Tree: treeTo,
		TreeEntry: object.TreeEntry{
			Name: "analyser.go",
			Mode: 0100644,
			Hash: plumbing.NewHash("ffffffffffffffffffffffffffffffffffffffff"),
		},
	}}
	res, err = fd.Consume(deps)
	assert.Nil(t, res)
	assert.NotNil(t, err)
}

func TestCountLines(t *testing.T) {
	blob, _ := testRepository.BlobObject(
		plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe"))
	lines, err := CountLines(blob)
	assert.Equal(t, lines, 12)
	assert.Nil(t, err)
	lines, err = CountLines(nil)
	assert.Equal(t, lines, -1)
	assert.NotNil(t, err)
	blob, _ = createDummyBlob(plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe"), true)
	lines, err = CountLines(blob)
	assert.Equal(t, lines, -1)
	assert.NotNil(t, err)
	// test_data/blob
	blob, err = testRepository.BlobObject(
		plumbing.NewHash("c86626638e0bc8cf47ca49bb1525b40e9737ee64"))
	assert.Nil(t, err)
	lines, err = CountLines(blob)
	assert.Equal(t, lines, -1)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "binary")
}

func TestBlobToString(t *testing.T) {
	blob, _ := testRepository.BlobObject(
		plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe"))
	str, err := BlobToString(blob)
	assert.Nil(t, err)
	assert.Equal(t, str, `language: go

go:
  - 1.7

go_import_path: gopkg.in/src-d/hercules.v1
`+"  "+`
script:
  - go test -v -cpu=1,2 ./...

notifications:
  email: false
`)
	str, err = BlobToString(nil)
	assert.Equal(t, str, "")
	assert.NotNil(t, err)
	blob, _ = createDummyBlob(plumbing.NewHash("291286b4ac41952cbd1389fda66420ec03c1a9fe"), true)
	str, err = BlobToString(blob)
	assert.Equal(t, str, "")
	assert.NotNil(t, err)
}
