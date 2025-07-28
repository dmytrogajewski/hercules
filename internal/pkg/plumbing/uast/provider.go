package uast

import (
	"fmt"

	"github.com/dmytrogajewski/hercules/internal/app/core"
	items "github.com/dmytrogajewski/hercules/internal/pkg/plumbing"
	"github.com/dmytrogajewski/hercules/pkg/uast"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

// DependencyUastChanges is the name of the dependency provided by Changes.
const DependencyUastChanges = "uast_changes"

// DependencyUasts is the name of the dependency provided by Extractor.
const DependencyUasts = "uasts"

// FeatureUast is the name of the feature which enables UAST-based analyses.
const FeatureUast = "uast"

// ConfigUASTProvider is the configuration key for UAST provider selection.
const ConfigUASTProvider = "uast_provider"

// Change represents a structural change between two versions of code
type Change struct {
	Before *node.Node
	After  *node.Node
	Change *object.Change
}

// Changes extracts UAST changes from file changes in commits.
type Changes struct {
	core.NoopMerger
	core.OneShotMergeProcessor

	parser *uast.Parser
	l      core.Logger
}

// Name returns the name of this PipelineItem.
func (changes *Changes) Name() string {
	return "UASTChanges"
}

// Provides returns the list of names of entities which are produced by this PipelineItem.
func (changes *Changes) Provides() []string {
	return []string{DependencyUastChanges}
}

// Requires returns the list of names of entities which are needed by this PipelineItem.
func (changes *Changes) Requires() []string {
	return []string{"file_diff", items.DependencyBlobCache}
}

// Features which must be enabled for this PipelineItem to be automatically inserted into the DAG.
func (changes *Changes) Features() []string {
	return []string{FeatureUast}
}

// ListConfigurationOptions returns the list of changeable public properties of this PipelineItem.
func (changes *Changes) ListConfigurationOptions() []core.ConfigurationOption {
	return []core.ConfigurationOption{}
}

// Configure applies the parameters specified in the command line.
func (changes *Changes) Configure(facts map[string]interface{}) error {
	if l, exists := facts[core.ConfigLogger].(core.Logger); exists {
		changes.l = l
	}
	return nil
}

// Initialize prepares and resets the item.
func (changes *Changes) Initialize(repository *git.Repository) error {
	changes.l = core.GetLogger()
	changes.OneShotMergeProcessor.Initialize()

	// Initialize UAST parser
	parser, err := uast.NewParser()
	if err != nil {
		return fmt.Errorf("failed to initialize UAST parser: %w", err)
	}
	changes.parser = parser

	return nil
}

// Consume processes the next commit.
func (changes *Changes) Consume(deps map[string]interface{}) (map[string]interface{}, error) {
	if !changes.ShouldConsumeCommit(deps) {
		return nil, nil
	}

	fileDiffs := deps[items.DependencyFileDiff].(map[string]items.FileDiffData)
	var result []Change

	// For now, we'll create a simple implementation that doesn't rely on object.Change
	// We'll need to implement proper UAST change detection based on file diffs
	// This is a placeholder implementation
	for filename, fileDiff := range fileDiffs {
		// Skip if no changes
		if len(fileDiff.Diffs) == 0 {
			continue
		}

		// For now, we'll create a simple change object
		// In a full implementation, we'd need to parse the before/after files
		// and detect structural changes
		change := &object.Change{
			From: object.ChangeEntry{
				Name: filename,
			},
			To: object.ChangeEntry{
				Name: filename,
			},
		}

		result = append(result, Change{
			Before: nil, // TODO: Parse before file if it exists
			After:  nil, // TODO: Parse after file if it exists
			Change: change,
		})
	}

	return map[string]interface{}{DependencyUastChanges: result}, nil
}

// parseFile parses a single file and returns its UAST.
func (changes *Changes) parseFile(hash plumbing.Hash, filename string, blobCache map[plumbing.Hash]*items.CachedBlob) (*node.Node, error) {
	// Check if the file is supported by our UAST parser
	if !changes.parser.IsSupported(filename) {
		return nil, fmt.Errorf("unsupported file type: %s", filename)
	}

	// Get file content from blob cache
	cachedBlob, exists := blobCache[hash]
	if !exists {
		return nil, fmt.Errorf("blob not found in cache: %s", hash.String())
	}

	// Get the file content - Data is a field, not a method
	content := cachedBlob.Data

	// Parse with UAST parser
	return changes.parser.Parse(filename, content)
}

// Fork clones the item the requested number of times.
func (changes *Changes) Fork(n int) []core.PipelineItem {
	return core.ForkCopyPipelineItem(changes, n)
}

// Merge combines several branches together.
func (changes *Changes) Merge(branches []core.PipelineItem) {
	// Implementation for merging branches
}

// Extractor extracts UASTs from files in commits.
type Extractor struct {
	core.NoopMerger
	core.OneShotMergeProcessor

	parser *uast.Parser
	l      core.Logger
}

// Name returns the name of this PipelineItem.
func (extractor *Extractor) Name() string {
	return "UASTExtractor"
}

// Provides returns the list of names of entities which are produced by this PipelineItem.
func (extractor *Extractor) Provides() []string {
	return []string{DependencyUasts}
}

// Requires returns the list of names of entities which are needed by this PipelineItem.
func (extractor *Extractor) Requires() []string {
	return []string{"blob_cache"}
}

// Features which must be enabled for this PipelineItem to be automatically inserted into the DAG.
func (extractor *Extractor) Features() []string {
	return []string{FeatureUast}
}

// ListConfigurationOptions returns the list of changeable public properties of this PipelineItem.
func (extractor *Extractor) ListConfigurationOptions() []core.ConfigurationOption {
	return []core.ConfigurationOption{}
}

// Configure applies the parameters specified in the command line.
func (extractor *Extractor) Configure(facts map[string]interface{}) error {
	if l, exists := facts[core.ConfigLogger].(core.Logger); exists {
		extractor.l = l
	}
	return nil
}

// Initialize prepares and resets the item.
func (extractor *Extractor) Initialize(repository *git.Repository) error {
	extractor.l = core.GetLogger()
	extractor.OneShotMergeProcessor.Initialize()

	// Initialize UAST parser
	parser, err := uast.NewParser()
	if err != nil {
		return fmt.Errorf("failed to initialize UAST parser: %w", err)
	}
	extractor.parser = parser

	return nil
}

// Consume processes the next commit.
func (extractor *Extractor) Consume(deps map[string]interface{}) (map[string]interface{}, error) {
	if !extractor.ShouldConsumeCommit(deps) {
		return nil, nil
	}

	// Implementation would extract UASTs from blob cache
	// For now, return empty result
	return map[string]interface{}{DependencyUasts: make(map[string]*node.Node)}, nil
}

// Fork clones the item the requested number of times.
func (extractor *Extractor) Fork(n int) []core.PipelineItem {
	return core.ForkCopyPipelineItem(extractor, n)
}

// Merge combines several branches together.
func (extractor *Extractor) Merge(branches []core.PipelineItem) {
	// Implementation for merging branches
}

func init() {
	core.Registry.Register(&Changes{})
	core.Registry.Register(&Extractor{})
}
