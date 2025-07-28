package fixtures

import (
	"github.com/sergi/go-diff/diffmatchpatch"
)

// FileDiff represents a file difference for testing
type FileDiff struct {
	OldLinesOfCode int
	NewLinesOfCode int
	Diffs          []diffmatchpatch.Diff
}

// NewFileDiff returns a new FileDiff instance for testing
func NewFileDiff() *FileDiff {
	return &FileDiff{
		OldLinesOfCode: 0,
		NewLinesOfCode: 0,
		Diffs:          []diffmatchpatch.Diff{},
	}
}
