package uast

import (
	"fmt"
	"unicode/utf8"

	"github.com/dmytrogajewski/hercules/internal/app/core"
	"github.com/dmytrogajewski/hercules/internal/pkg/plumbing"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/go-git/go-git/v6"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var _ core.PipelineItem = (*FileDiffRefiner)(nil)

// FileDiffRefiner uses UASTs to improve the human interpretability of diffs.
// It is a PipelineItem.
// The idea behind this algorithm is simple: in case of multiple choices which are equally
// optimal, choose the one which touches less AST nodes.
type FileDiffRefiner struct {
	core.NoopMerger

	l core.Logger
}

// Name of this PipelineItem. Uniquely identifies the type, used for mapping keys, etc.
func (ref *FileDiffRefiner) Name() string {
	return "FileDiffRefiner"
}

// Provides returns the list of names of entities which are produced by this PipelineItem.
// Each produced entity will be inserted into `deps` of dependent Consume()-s according
// to this list. Also used by core.Registry to build the global map of providers.
func (ref *FileDiffRefiner) Provides() []string {
	return []string{plumbing.DependencyFileDiff}
}

// Requires returns the list of names of entities which are needed by this PipelineItem.
// Each requested entity will be inserted into `deps` of Consume(). In turn, those
// entities are Provides() upstream.
func (ref *FileDiffRefiner) Requires() []string {
	return []string{plumbing.DependencyFileDiff, DependencyUastChanges}
}

// Features which must be enabled for this PipelineItem to be automatically inserted into the DAG.
func (ref *FileDiffRefiner) Features() []string {
	return []string{FeatureUast}
}

// ListConfigurationOptions returns the list of changeable public properties of this PipelineItem.
func (ref *FileDiffRefiner) ListConfigurationOptions() []core.ConfigurationOption {
	return []core.ConfigurationOption{}
}

// Configure sets the properties previously published by ListConfigurationOptions().
func (ref *FileDiffRefiner) Configure(facts map[string]interface{}) error {
	if l, exists := facts[core.ConfigLogger].(core.Logger); exists {
		ref.l = l
	}
	return nil
}

// Initialize resets the temporary caches and prepares this PipelineItem for a series of Consume()
// calls. The repository which is going to be analysed is supplied as an argument.
func (ref *FileDiffRefiner) Initialize(repository *git.Repository) error {
	ref.l = core.GetLogger()
	return nil
}

// Consume runs this PipelineItem on the next commit data.
// `deps` contain all the results from upstream PipelineItem-s as requested by Requires().
// Additionally, DependencyCommit is always present there and represents the analysed *object.Commit.
// This function returns the mapping with analysis results. The keys must be the same as
// in Provides(). If there was an error, nil is returned.
func (ref *FileDiffRefiner) Consume(deps map[string]interface{}) (map[string]interface{}, error) {
	changesList := deps[DependencyUastChanges].([]Change)
	changes := map[string]Change{}
	for _, change := range changesList {
		if change.Before != nil && change.After != nil {
			changes[change.Change.To.Name] = change
		}
	}
	diffs := deps[plumbing.DependencyFileDiff].(map[string]plumbing.FileDiffData)
	result := map[string]plumbing.FileDiffData{}
	for fileName, oldDiff := range diffs {
		uastChange, exists := changes[fileName]
		if !exists {
			result[fileName] = oldDiff
			continue
		}
		suspicious := map[int][2]int{}
		line := 0
		for i, diff := range oldDiff.Diffs {
			if i == len(oldDiff.Diffs)-1 {
				break
			}
			if diff.Type == diffmatchpatch.DiffInsert &&
				oldDiff.Diffs[i+1].Type == diffmatchpatch.DiffEqual {
				matched := 0
				runesAdded := []rune(diff.Text)
				runesEqual := []rune(oldDiff.Diffs[i+1].Text)
				for ; matched < len(runesAdded) && matched < len(runesEqual) &&
					runesAdded[matched] == runesEqual[matched]; matched++ {
				}
				if matched > 0 {
					suspicious[i] = [2]int{line, matched}
				}
			}
			if diff.Type != diffmatchpatch.DiffDelete {
				line += utf8.RuneCountInString(diff.Text)
			}
		}
		if len(suspicious) == 0 {
			result[fileName] = oldDiff
			continue
		}
		line2node := make([][]*node.Node, oldDiff.NewLinesOfCode)
		VisitEachNode(uastChange.After, func(node *node.Node) {
			if node.Pos != nil {
				for l := node.Pos.StartLine; l <= node.Pos.EndLine; l++ {
					if l > 0 && int(l) <= len(line2node) {
						line2node[l-1] = append(line2node[l-1], node) // line starts with 1
					}
				}
			}
		})
		newDiff := plumbing.FileDiffData{
			OldLinesOfCode: oldDiff.OldLinesOfCode,
			NewLinesOfCode: oldDiff.NewLinesOfCode,
			Diffs:          []diffmatchpatch.Diff{},
		}
		skipNext := false
		for i, diff := range oldDiff.Diffs {
			if skipNext {
				skipNext = false
				continue
			}
			info, exists := suspicious[i]
			if !exists {
				newDiff.Diffs = append(newDiff.Diffs, diff)
				continue
			}
			line := info[0]
			matched := info[1]
			size := utf8.RuneCountInString(diff.Text)
			n1 := countNodesInInterval(line2node, line, line+size)
			n2 := countNodesInInterval(line2node, line+matched, line+size+matched)
			if n1 <= n2 {
				newDiff.Diffs = append(newDiff.Diffs, diff)
				continue
			}
			skipNext = true
			runes := []rune(diff.Text)
			newDiff.Diffs = append(newDiff.Diffs, diffmatchpatch.Diff{
				Type: diffmatchpatch.DiffEqual, Text: string(runes[:matched]),
			})
			newDiff.Diffs = append(newDiff.Diffs, diffmatchpatch.Diff{
				Type: diffmatchpatch.DiffInsert, Text: string(runes[matched:]) + string(runes[:matched]),
			})
			runes = []rune(oldDiff.Diffs[i+1].Text)
			if len(runes) > matched {
				newDiff.Diffs = append(newDiff.Diffs, diffmatchpatch.Diff{
					Type: diffmatchpatch.DiffEqual, Text: string(runes[matched:]),
				})
			}
		}
		result[fileName] = newDiff
	}
	return map[string]interface{}{plumbing.DependencyFileDiff: result}, nil
}

// Fork clones this PipelineItem.
func (ref *FileDiffRefiner) Fork(n int) []core.PipelineItem {
	return core.ForkSamePipelineItem(ref, n)
}

// VisitEachNode is a handy routine to execute a callback on every node in the subtree,
// including the root itself. Depth first tree traversal.
func VisitEachNode(root *node.Node, payload func(*node.Node)) {
	root.VisitPreOrder(payload)
}

func countNodesInInterval(occupiedMap [][]*node.Node, start, end int) int {
	inodes := map[string]bool{}
	for i := start; i < end && i < len(occupiedMap); i++ {
		for _, node := range occupiedMap[i] {
			// Create a unique key for the node based on its ID and position
			key := nodeUniqueKey(node)
			inodes[key] = true
		}
	}
	return len(inodes)
}

// nodeUniqueKey creates a unique key for a node based on its ID and position
func nodeUniqueKey(n *node.Node) string {
	if n == nil {
		return ""
	}

	key := n.Id
	if n.Pos != nil {
		key += fmt.Sprintf(":%d:%d:%d:%d",
			n.Pos.StartLine, n.Pos.StartCol,
			n.Pos.EndLine, n.Pos.EndCol)
	}
	return key
}

func init() {
	core.Registry.Register(&FileDiffRefiner{})
}
