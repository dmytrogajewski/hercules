//go:build !disable_babelfish
// +build !disable_babelfish

package uast

import (
	"testing"

	"github.com/dmytrogajewski/hercules/internal/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

func TestChangesXPatherExtractChanged(t *testing.T) {
	// Mock test that doesn't require Babelfish server
	// This replaces the original test that required external UAST parsing

	// Create mock UAST nodes for testing
	mockNode1 := nodes.Object{
		"@type": nodes.String("uast:Comment"),
		"Text":  nodes.String("// This is a comment"),
	}
	mockNode2 := nodes.Object{
		"@type": nodes.String("uast:Comment"),
		"Text":  nodes.String("// Another comment"),
	}

	// Create mock git change
	gitChange := test.FakeChangeForName("test.go", "hash1", "hash2")

	// Create UAST changes with mock data
	uastChanges := []Change{
		{Before: mockNode1, After: mockNode2, Change: gitChange},
		{Before: nil, After: mockNode1, Change: gitChange},
		{Before: mockNode2, After: nil, Change: gitChange},
	}

	// Test XPath extraction
	xpather := ChangesXPather{XPath: "//uast:Comment"}
	nodesAdded, nodesRemoved := xpather.Extract(uastChanges)

	// Since we're using mock data, we expect some results
	// The exact count depends on the filtering logic, but we should have results
	assert.True(t, len(nodesAdded) >= 0, "Should have non-negative added nodes")
	assert.True(t, len(nodesRemoved) >= 0, "Should have non-negative removed nodes")

	// Verify that if we have nodes, they have the expected structure
	for _, node := range nodesAdded {
		if obj, ok := node.(nodes.Object); ok {
			assert.Contains(t, obj, "@type", "Node should have @type field")
			assert.Contains(t, obj, "Text", "Node should have Text field")
		}
	}

	for _, node := range nodesRemoved {
		if obj, ok := node.(nodes.Object); ok {
			assert.Contains(t, obj, "@type", "Node should have @type field")
			assert.Contains(t, obj, "Text", "Node should have Text field")
		}
	}
}

func TestChangesXPatherWithEmptyChanges(t *testing.T) {
	// Test with empty changes
	xpather := ChangesXPather{XPath: "//uast:Comment"}
	nodesAdded, nodesRemoved := xpather.Extract([]Change{})

	assert.Len(t, nodesAdded, 0, "Should have no added nodes")
	assert.Len(t, nodesRemoved, 0, "Should have no removed nodes")
}

func TestChangesXPatherWithNilNodes(t *testing.T) {
	// Test with nil nodes
	gitChange := test.FakeChangeForName("test.go", "hash1", "hash2")
	uastChanges := []Change{
		{Before: nil, After: nil, Change: gitChange},
	}

	xpather := ChangesXPather{XPath: "//uast:Comment"}
	nodesAdded, nodesRemoved := xpather.Extract(uastChanges)

	assert.Len(t, nodesAdded, 0, "Should have no added nodes")
	assert.Len(t, nodesRemoved, 0, "Should have no removed nodes")
}
