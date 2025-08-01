package research

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/dmytrogajewski/hercules/api/proto/pb"
	"github.com/dmytrogajewski/hercules/internal/app/core"
	items "github.com/dmytrogajewski/hercules/internal/pkg/plumbing"
	uast_items "github.com/dmytrogajewski/hercules/internal/pkg/plumbing/uast"
	"github.com/dmytrogajewski/hercules/pkg/levenshtein"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/sergi/go-diff/diffmatchpatch"
	"google.golang.org/protobuf/proto"
)

// TyposDatasetBuilder collects pairs of typo-fix in source code identifiers.
type TyposDatasetBuilder struct {
	core.NoopMerger

	// MaximumAllowedDistance is the maximum Levenshtein distance between two identifiers
	// to consider them a typo-fix pair.
	MaximumAllowedDistance int

	// typos stores the found typo-fix pairs.
	typos []Typo
	// lcontext is the Context for measuring Levenshtein distance between lines.
	lcontext *levenshtein.Context
	// remote carries the repository remote URL (for debugging)
	remote string

	l core.Logger
}

// TyposResult is returned by TyposDatasetBuilder.Finalize() and carries the found typo-fix
// pairs of identifiers.
type TyposResult struct {
	Typos []Typo
}

// Typo carries the information about a typo-fix pair.
type Typo struct {
	Wrong   string
	Correct string
	Commit  plumbing.Hash
	File    string
	Line    int
}

const (
	// DefaultMaximumAllowedTypoDistance is the default value of the maximum Levenshtein distance
	// between two identifiers to consider them a typo-fix pair.
	DefaultMaximumAllowedTypoDistance = 4
	// ConfigTyposDatasetMaximumAllowedDistance is the name of the configuration option
	// (`TyposDatasetBuilder.Configure()`) which sets the maximum Levenshtein distance between
	// two identifiers to consider them a typo-fix pair.
	ConfigTyposDatasetMaximumAllowedDistance = "TyposDatasetBuilder.MaximumAllowedDistance"
)

// Name of this PipelineItem. Uniquely identifies the type, used for mapping keys, etc.
func (tdb *TyposDatasetBuilder) Name() string {
	return "TyposDataset"
}

// Provides returns the list of names of entities which are produced by this PipelineItem.
// Each produced entity will be inserted into `deps` of dependent Consume()-s according
// to this list. Also used by core.Registry to build the global map of providers.
func (tdb *TyposDatasetBuilder) Provides() []string {
	return []string{}
}

// Requires returns the list of names of entities which are needed by this PipelineItem.
// Each requested entity will be inserted into `deps` of Consume(). In turn, those
// entities are Provides() upstream.
func (tdb *TyposDatasetBuilder) Requires() []string {
	return []string{
		uast_items.DependencyUastChanges, items.DependencyFileDiff, items.DependencyBlobCache}
}

// ListConfigurationOptions returns the list of changeable public properties of this PipelineItem.
func (tdb *TyposDatasetBuilder) ListConfigurationOptions() []core.ConfigurationOption {
	options := [...]core.ConfigurationOption{{
		Name: ConfigTyposDatasetMaximumAllowedDistance,
		Description: "Maximum Levenshtein distance between two identifiers to consider them " +
			"a typo-fix pair.",
		Flag:    "typos-max-distance",
		Type:    core.IntConfigurationOption,
		Default: DefaultMaximumAllowedTypoDistance},
	}
	return options[:]
}

// Configure sets the properties previously published by ListConfigurationOptions().
func (tdb *TyposDatasetBuilder) Configure(facts map[string]interface{}) error {
	if l, exists := facts[core.ConfigLogger].(core.Logger); exists {
		tdb.l = l
	}
	if val, exists := facts[ConfigTyposDatasetMaximumAllowedDistance].(int); exists {
		tdb.MaximumAllowedDistance = val
	}
	return nil
}

// Flag for the command line switch which enables this analysis.
func (tdb *TyposDatasetBuilder) Flag() string {
	return "typos-dataset"
}

// Description returns the text which explains what the analysis is doing.
func (tdb *TyposDatasetBuilder) Description() string {
	return "Extracts typo-fix identifier pairs from source code in commit diffs."
}

// Initialize resets the temporary caches and prepares this PipelineItem for a series of Consume()
// calls. The repository which is going to be analysed is supplied as an argument.
func (tdb *TyposDatasetBuilder) Initialize(repository *git.Repository) error {
	tdb.l = core.GetLogger()
	if tdb.MaximumAllowedDistance <= 0 {
		tdb.MaximumAllowedDistance = DefaultMaximumAllowedTypoDistance
	}
	tdb.lcontext = &levenshtein.Context{}
	tdb.remote = core.GetSensibleRemote(repository)
	return nil
}

type candidate struct {
	Before int
	After  int
}

// Consume runs this PipelineItem on the next commit data.
// `deps` contain all the results from upstream PipelineItem-s as requested by Requires().
// Additionally, DependencyCommit is always present there and represents the analysed *object.Commit.
// This function returns the mapping with analysis results. The keys must be the same as
// in Provides(). If there was an error, nil is returned.
func (tdb *TyposDatasetBuilder) Consume(deps map[string]interface{}) (map[string]interface{}, error) {
	if deps[core.DependencyIsMerge].(bool) {
		return nil, nil
	}
	commit := deps[core.DependencyCommit].(*object.Commit).Hash
	cache := deps[items.DependencyBlobCache].(map[plumbing.Hash]*items.CachedBlob)
	diffs := deps[items.DependencyFileDiff].(map[string]items.FileDiffData)
	changes := deps[uast_items.DependencyUastChanges].([]uast_items.Change)
	for _, change := range changes {
		if change.Before == nil || change.After == nil {
			continue
		}
		linesBefore := bytes.Split(cache[change.Change.From.TreeEntry.Hash].Data, []byte{'\n'})
		linesAfter := bytes.Split(cache[change.Change.To.TreeEntry.Hash].Data, []byte{'\n'})
		diff := diffs[change.Change.To.Name]
		var lineNumBefore, lineNumAfter int
		var candidates []candidate
		focusedLinesBefore := map[int]bool{}
		focusedLinesAfter := map[int]bool{}
		removedSize := 0
		for _, edit := range diff.Diffs {
			size := utf8.RuneCountInString(edit.Text)
			switch edit.Type {
			case diffmatchpatch.DiffDelete:
				lineNumBefore += size
				removedSize = size
			case diffmatchpatch.DiffInsert:
				if size == removedSize {
					for i := 0; i < size; i++ {
						lb := lineNumBefore - size + i
						la := lineNumAfter + i
						dist := tdb.lcontext.Distance(string(linesBefore[lb]), string(linesAfter[la]))
						if dist <= tdb.MaximumAllowedDistance {
							candidates = append(candidates, candidate{lb, la})
							focusedLinesBefore[lb] = true
							focusedLinesAfter[la] = true
						}
					}
				}
				lineNumAfter += size
				removedSize = 0
			case diffmatchpatch.DiffEqual:
				lineNumBefore += size
				lineNumAfter += size
				removedSize = 0
			}
		}
		if len(candidates) == 0 {
			continue
		}
		// at this point we have pairs of very similar lines
		// we need to build the line mappings of the identifiers before/after the change
		// we should keep only those which are present on those focused lines
		addedIdentifiers := map[int][]*node.Node{}
		removedIdentifiers := map[int][]*node.Node{}

		// Extract identifiers from before and after UASTs
		if change.Before != nil {
			identifiers := tdb.extractIdentifiers(change.Before)
			for _, id := range identifiers {
				if id.Pos != nil && focusedLinesBefore[int(id.Pos.StartLine)-1] {
					line := int(id.Pos.StartLine) - 1
					removedIdentifiers[line] = append(removedIdentifiers[line], id)
				}
			}
		}

		if change.After != nil {
			identifiers := tdb.extractIdentifiers(change.After)
			for _, id := range identifiers {
				if id.Pos != nil && focusedLinesAfter[int(id.Pos.StartLine)-1] {
					line := int(id.Pos.StartLine) - 1
					addedIdentifiers[line] = append(addedIdentifiers[line], id)
				}
			}
		}

		for _, c := range candidates {
			nodesBefore := removedIdentifiers[c.Before]
			nodesAfter := addedIdentifiers[c.After]
			if len(nodesBefore) == 1 && len(nodesAfter) == 1 {
				idBefore := nodesBefore[0].Token
				idAfter := nodesAfter[0].Token
				tdb.typos = append(tdb.typos, Typo{
					Wrong:   idBefore,
					Correct: idAfter,
					Commit:  commit,
					File:    change.Change.To.Name,
					Line:    c.After,
				})
			}
		}
	}
	return nil, nil
}

// extractIdentifiers extracts all identifier nodes from a UAST
func (tdb *TyposDatasetBuilder) extractIdentifiers(root *node.Node) []*node.Node {
	var identifiers []*node.Node
	root.VisitPreOrder(func(n *node.Node) {
		if n.Type == node.UASTIdentifier {
			identifiers = append(identifiers, n)
		}
	})
	return identifiers
}

// Finalize returns the result of the analysis. Further Consume() calls are not expected.
func (tdb *TyposDatasetBuilder) Finalize() interface{} {
	// deduplicate
	typos := make([]Typo, 0, len(tdb.typos))
	pairs := map[string]bool{}
	for _, t := range tdb.typos {
		id := t.Wrong + "|" + t.Correct
		if _, exists := pairs[id]; !exists {
			pairs[id] = true
			typos = append(typos, t)
		}
	}
	return TyposResult{Typos: typos}
}

// Fork clones this pipeline item.
func (tdb *TyposDatasetBuilder) Fork(n int) []core.PipelineItem {
	return core.ForkSamePipelineItem(tdb, n)
}

// Serialize converts the analysis result as returned by Finalize() to text or bytes.
// The text format is YAML and the bytes format is Protocol Buffers.
func (tdb *TyposDatasetBuilder) Serialize(result interface{}, binary bool, writer io.Writer) error {
	commitsResult := result.(TyposResult)
	if binary {
		return tdb.serializeBinary(&commitsResult, writer)
	}
	tdb.serializeText(&commitsResult, writer)
	return nil
}

func (tdb *TyposDatasetBuilder) serializeText(result *TyposResult, writer io.Writer) {
	for _, t := range result.Typos {
		fmt.Fprintf(writer, "  - wrong: %s\n", t.Wrong)
		fmt.Fprintf(writer, "    correct: %s\n", t.Correct)
		fmt.Fprintf(writer, "    commit: %s\n", t.Commit.String())
		fmt.Fprintf(writer, "    file: %s\n", t.File)
		fmt.Fprintf(writer, "    line: %d\n", t.Line)
	}
}

func (tdb *TyposDatasetBuilder) serializeBinary(result *TyposResult, writer io.Writer) error {
	message := pb.TyposDataset{}
	message.Typos = make([]*pb.Typo, len(result.Typos))
	for i, t := range result.Typos {
		message.Typos[i] = &pb.Typo{
			Wrong:   t.Wrong,
			Correct: t.Correct,
			Commit:  t.Commit.String(),
			File:    t.File,
			Line:    int32(t.Line),
		}
	}
	serialized, err := proto.Marshal(&message)
	if err != nil {
		return err
	}
	_, err = writer.Write(serialized)
	return err
}

func init() {
	core.Registry.Register(&TyposDatasetBuilder{})
}
