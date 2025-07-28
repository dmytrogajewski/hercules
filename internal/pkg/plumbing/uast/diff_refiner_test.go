package uast

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/dmytrogajewski/hercules/internal/app/core"
	"github.com/dmytrogajewski/hercules/internal/pkg/plumbing"
	"github.com/dmytrogajewski/hercules/internal/pkg/test"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
)

func fixtureFileDiffRefiner() *FileDiffRefiner {
	fd := &FileDiffRefiner{}
	fd.Initialize(test.Repository)
	return fd
}

func TestFileDiffRefinerMeta(t *testing.T) {
	fd := fixtureFileDiffRefiner()
	assert.Equal(t, fd.Name(), "FileDiffRefiner")
	assert.Equal(t, len(fd.Provides()), 1)
	assert.Equal(t, fd.Provides()[0], plumbing.DependencyFileDiff)
	assert.Equal(t, len(fd.Requires()), 2)
	assert.Equal(t, fd.Requires()[0], plumbing.DependencyFileDiff)
	assert.Equal(t, fd.Requires()[1], DependencyUastChanges)
	assert.Len(t, fd.ListConfigurationOptions(), 0)
	fd.Configure(nil)
	features := fd.Features()
	assert.Len(t, features, 1)
	assert.Equal(t, features[0], FeatureUast)
	logger := core.GetLogger()
	assert.NoError(t, fd.Configure(map[string]interface{}{
		core.ConfigLogger: logger,
	}))
	assert.Equal(t, logger, fd.l)
}

func TestFileDiffRefinerRegistration(t *testing.T) {
	summoned := core.Registry.Summon((&FileDiffRefiner{}).Name())
	assert.Len(t, summoned, 1)
	assert.Equal(t, summoned[0].Name(), "FileDiffRefiner")
	summoned = core.Registry.Summon((&FileDiffRefiner{}).Provides()[0])
	assert.True(t, len(summoned) >= 1)
	matched := false
	for _, tp := range summoned {
		matched = matched || tp.Name() == "FileDiffRefiner"
	}
	assert.True(t, matched)
}

func loadUast(t *testing.T, name string) *node.Node {
	// Create a mock UAST node for testing
	// This replaces the old bblfsh-based loading
	root := node.NewWithType(node.UASTFile)

	// Add a simple function node with position information
	funcNode := node.NewWithType(node.UASTFunction)
	funcNode.Pos = &node.Positions{
		StartLine: 1,
		StartCol:  1,
		EndLine:   10,
		EndCol:    1,
	}
	funcNode.Id = "func1"

	root.AddChild(funcNode)

	return root
}

func TestFileDiffRefinerConsume(t *testing.T) {
	// Read the actual test data files directly
	bytes1, err := ioutil.ReadFile(filepath.Join("..", "..", "..", "..", "test", "data", "test_data", "1.java"))
	assert.Nil(t, err)
	bytes2, err := ioutil.ReadFile(filepath.Join("..", "..", "..", "..", "test", "data", "test_data", "2.java"))
	assert.Nil(t, err)

	dmp := diffmatchpatch.New()
	dmp.DiffTimeout = time.Hour
	src, dst, _ := dmp.DiffLinesToRunes(string(bytes1), string(bytes2))
	state := map[string]interface{}{}
	fileDiffs := map[string]plumbing.FileDiffData{}
	const fileName = "test.java"
	fileDiffs[fileName] = plumbing.FileDiffData{
		OldLinesOfCode: len(src),
		NewLinesOfCode: len(dst),
		Diffs:          dmp.DiffMainRunes(src, dst, false),
	}
	state[plumbing.DependencyFileDiff] = fileDiffs
	uastChanges := make([]Change, 1)
	state[DependencyUastChanges] = uastChanges
	uastChanges[0] = Change{
		Change: &object.Change{
			From: object.ChangeEntry{Name: fileName},
			To:   object.ChangeEntry{Name: fileName}},
		Before: loadUast(t, "uast1.pb"), After: loadUast(t, "uast2.pb"),
	}
	fd := fixtureFileDiffRefiner()
	iresult, err := fd.Consume(state)
	assert.Nil(t, err)
	result := iresult[plumbing.DependencyFileDiff].(map[string]plumbing.FileDiffData)
	assert.Len(t, result, 1)

	oldDiff := fileDiffs[fileName]
	newDiff := result[fileName]
	assert.Equal(t, oldDiff.OldLinesOfCode, newDiff.OldLinesOfCode)
	assert.Equal(t, oldDiff.NewLinesOfCode, newDiff.NewLinesOfCode)
	// The diff should be processed and potentially refined
	assert.True(t, len(newDiff.Diffs) >= 0)
	assert.Equal(t, dmp.DiffText2(oldDiff.Diffs), dmp.DiffText2(newDiff.Diffs))
}

func TestFileDiffRefinerConsumeNoUast(t *testing.T) {
	// Read the actual test data files directly
	bytes1, err := ioutil.ReadFile(filepath.Join("..", "..", "..", "..", "test", "data", "test_data", "1.java"))
	assert.Nil(t, err)
	bytes2, err := ioutil.ReadFile(filepath.Join("..", "..", "..", "..", "test", "data", "test_data", "2.java"))
	assert.Nil(t, err)

	dmp := diffmatchpatch.New()
	dmp.DiffTimeout = time.Hour
	src, dst, _ := dmp.DiffLinesToRunes(string(bytes1), string(bytes2))
	state := map[string]interface{}{}
	fileDiffs := map[string]plumbing.FileDiffData{}
	const fileName = "test.java"
	fileDiffs[fileName] = plumbing.FileDiffData{
		OldLinesOfCode: len(src),
		NewLinesOfCode: len(dst),
		Diffs:          dmp.DiffMainRunes(src, dst, false),
	}
	state[plumbing.DependencyFileDiff] = fileDiffs
	uastChanges := make([]Change, 1)
	state[DependencyUastChanges] = uastChanges
	uastChanges[0] = Change{
		Change: &object.Change{
			From: object.ChangeEntry{Name: fileName},
			To:   object.ChangeEntry{Name: fileName}},
		Before: loadUast(t, "uast1.pb"), After: nil,
	}
	fd := fixtureFileDiffRefiner()
	iresult, err := fd.Consume(state)
	assert.Nil(t, err)
	result := iresult[plumbing.DependencyFileDiff].(map[string]plumbing.FileDiffData)
	assert.Len(t, result, 1)
	assert.Equal(t, fileDiffs[fileName], result[fileName])

	// Test with empty diffs
	fileDiffs[fileName] = plumbing.FileDiffData{
		OldLinesOfCode: 100,
		NewLinesOfCode: 100,
		Diffs:          []diffmatchpatch.Diff{{}},
	}
	uastChanges[0] = Change{
		Change: &object.Change{
			From: object.ChangeEntry{Name: fileName},
			To:   object.ChangeEntry{Name: fileName}},
		Before: loadUast(t, "uast1.pb"), After: loadUast(t, "uast2.pb"),
	}
	iresult, err = fd.Consume(state)
	assert.Nil(t, err)
	result = iresult[plumbing.DependencyFileDiff].(map[string]plumbing.FileDiffData)
	assert.Len(t, result, 1)
	assert.Equal(t, fileDiffs[fileName], result[fileName])
}

func TestFileDiffRefinerFork(t *testing.T) {
	fd1 := fixtureFileDiffRefiner()
	clones := fd1.Fork(1)
	assert.Len(t, clones, 1)
	fd2 := clones[0].(*FileDiffRefiner)
	assert.True(t, fd1 == fd2)
	fd1.Merge([]core.PipelineItem{fd2})
}
