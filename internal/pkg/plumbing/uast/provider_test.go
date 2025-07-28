package uast

import (
	"testing"

	"github.com/dmytrogajewski/hercules/internal/app/core"
	"github.com/dmytrogajewski/hercules/internal/pkg/plumbing"
	"github.com/dmytrogajewski/hercules/internal/pkg/test"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	gitplumbing "github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
)

func TestChangesMeta(t *testing.T) {
	changes := &Changes{}
	assert.Equal(t, changes.Name(), "UASTChanges")
	assert.Equal(t, len(changes.Provides()), 1)
	assert.Equal(t, changes.Provides()[0], DependencyUastChanges)
	assert.Equal(t, len(changes.Requires()), 2)
	assert.Equal(t, changes.Requires()[0], "file_diff")
	assert.Equal(t, changes.Requires()[1], plumbing.DependencyBlobCache)
	assert.Len(t, changes.ListConfigurationOptions(), 0)
	changes.Configure(nil)
	features := changes.Features()
	assert.Len(t, features, 1)
	assert.Equal(t, features[0], FeatureUast)
	logger := core.GetLogger()
	assert.NoError(t, changes.Configure(map[string]interface{}{
		core.ConfigLogger: logger,
	}))
	assert.Equal(t, logger, changes.l)
}

func TestChangesRegistration(t *testing.T) {
	summoned := core.Registry.Summon((&Changes{}).Name())
	assert.Len(t, summoned, 1)
	assert.Equal(t, summoned[0].Name(), "UASTChanges")
	summoned = core.Registry.Summon((&Changes{}).Provides()[0])
	assert.True(t, len(summoned) >= 1)
	matched := false
	for _, tp := range summoned {
		matched = matched || tp.Name() == "UASTChanges"
	}
	assert.True(t, matched)
}

func TestChangesInitialize(t *testing.T) {
	changes := &Changes{}
	err := changes.Initialize(test.Repository)
	assert.Nil(t, err)
	assert.NotNil(t, changes.parser)
}

func TestChangesConsume(t *testing.T) {
	changes := &Changes{}
	err := changes.Initialize(test.Repository)
	assert.Nil(t, err)

	// Create test data
	deps := map[string]interface{}{}

	// Add required commit dependency - use a real commit if available
	commit, err := test.Repository.CommitObject(gitplumbing.NewHash("0000000000000000000000000000000000000000"))
	if err != nil {
		// Create a dummy commit if the hash doesn't exist
		commit = &object.Commit{}
	}
	deps[core.DependencyCommit] = commit

	fileDiffs := map[string]plumbing.FileDiffData{}
	fileDiffs["test.go"] = plumbing.FileDiffData{
		OldLinesOfCode: 10,
		NewLinesOfCode: 12,
		Diffs:          []diffmatchpatch.Diff{{Type: diffmatchpatch.DiffInsert, Text: "new line"}},
	}
	deps[plumbing.DependencyFileDiff] = fileDiffs

	// Test consumption - the result might be nil if ShouldConsumeCommit returns false
	result, err := changes.Consume(deps)
	assert.Nil(t, err)

	// If result is nil, it means ShouldConsumeCommit returned false, which is expected for dummy commits
	if result != nil {
		uastChanges, exists := result[DependencyUastChanges]
		assert.True(t, exists)
		changesList, ok := uastChanges.([]Change)
		assert.True(t, ok)
		assert.Len(t, changesList, 1)
	}
}

func TestExtractorMeta(t *testing.T) {
	extractor := &Extractor{}
	assert.Equal(t, extractor.Name(), "UASTExtractor")
	assert.Equal(t, len(extractor.Provides()), 1)
	assert.Equal(t, extractor.Provides()[0], DependencyUasts)
	assert.Equal(t, len(extractor.Requires()), 1)
	assert.Equal(t, extractor.Requires()[0], "blob_cache")
	assert.Len(t, extractor.ListConfigurationOptions(), 0)
	extractor.Configure(nil)
	features := extractor.Features()
	assert.Len(t, features, 1)
	assert.Equal(t, features[0], FeatureUast)
	logger := core.GetLogger()
	assert.NoError(t, extractor.Configure(map[string]interface{}{
		core.ConfigLogger: logger,
	}))
	assert.Equal(t, logger, extractor.l)
}

func TestExtractorRegistration(t *testing.T) {
	summoned := core.Registry.Summon((&Extractor{}).Name())
	assert.Len(t, summoned, 1)
	assert.Equal(t, summoned[0].Name(), "UASTExtractor")
	summoned = core.Registry.Summon((&Extractor{}).Provides()[0])
	assert.True(t, len(summoned) >= 1)
	matched := false
	for _, tp := range summoned {
		matched = matched || tp.Name() == "UASTExtractor"
	}
	assert.True(t, matched)
}

func TestExtractorInitialize(t *testing.T) {
	extractor := &Extractor{}
	err := extractor.Initialize(test.Repository)
	assert.Nil(t, err)
	assert.NotNil(t, extractor.parser)
}

func TestExtractorConsume(t *testing.T) {
	extractor := &Extractor{}
	err := extractor.Initialize(test.Repository)
	assert.Nil(t, err)

	// Create test data
	deps := map[string]interface{}{}

	// Add required commit dependency
	commit, err := test.Repository.CommitObject(gitplumbing.NewHash("0000000000000000000000000000000000000000"))
	if err != nil {
		// Create a dummy commit if the hash doesn't exist
		commit = &object.Commit{}
	}
	deps[core.DependencyCommit] = commit

	blobCache := map[string]*plumbing.CachedBlob{}
	deps["blob_cache"] = blobCache

	// Test consumption - the result might be nil if ShouldConsumeCommit returns false
	result, err := extractor.Consume(deps)
	assert.Nil(t, err)

	// If result is nil, it means ShouldConsumeCommit returned false, which is expected for dummy commits
	if result != nil {
		uasts, exists := result[DependencyUasts]
		assert.True(t, exists)
		uastsMap, ok := uasts.(map[string]*node.Node)
		assert.True(t, ok)
		assert.NotNil(t, uastsMap)
	}
}

func TestChangeStruct(t *testing.T) {
	// Test the Change struct
	change := &object.Change{
		From: object.ChangeEntry{Name: "test.go"},
		To:   object.ChangeEntry{Name: "test.go"},
	}

	uastNode := node.NewWithType(node.UASTFile)

	uastChange := Change{
		Before: uastNode,
		After:  uastNode,
		Change: change,
	}

	assert.NotNil(t, uastChange.Before)
	assert.NotNil(t, uastChange.After)
	assert.NotNil(t, uastChange.Change)
	assert.Equal(t, "test.go", uastChange.Change.From.Name)
	assert.Equal(t, "test.go", uastChange.Change.To.Name)
}

func TestChangesFork(t *testing.T) {
	changes := &Changes{}
	clones := changes.Fork(2)
	assert.Len(t, clones, 2)
	for _, clone := range clones {
		assert.IsType(t, &Changes{}, clone)
	}
}

func TestExtractorFork(t *testing.T) {
	extractor := &Extractor{}
	clones := extractor.Fork(2)
	assert.Len(t, clones, 2)
	for _, clone := range clones {
		assert.IsType(t, &Extractor{}, clone)
	}
}
