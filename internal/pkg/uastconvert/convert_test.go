package uastconvert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	uast_nodes "gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

func TestConvertSitterNodeToUAST_Nil(t *testing.T) {
	if ConvertSitterNodeToUAST(nil, nil) != nil {
		t.Error("Expected nil for nil input")
	}
}

func TestHashNode_Basic(t *testing.T) {
	// Create a simple UAST node
	node := uast_nodes.Object{
		"@type":  uast_nodes.String("test_node"),
		"@token": uast_nodes.String("test_content"),
	}

	hash, err := HashNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, hash)
	assert.Len(t, hash, 32) // SHA256 hash length
}

func TestSerializeNode_Basic(t *testing.T) {
	// Create a simple UAST node
	node := uast_nodes.Object{
		"@type":  uast_nodes.String("test_node"),
		"@token": uast_nodes.String("test_content"),
	}

	serialized, err := SerializeNode(node)
	assert.NoError(t, err)
	assert.NotNil(t, serialized)

	// Check that the serialized form has the expected structure
	assert.Equal(t, "test_node", serialized["@type"])
	assert.Equal(t, "test_content", serialized["@token"])
}

func TestStringifyNode_Basic(t *testing.T) {
	// Create a simple UAST node
	node := uast_nodes.Object{
		"@type":  uast_nodes.String("test_node"),
		"@token": uast_nodes.String("test_content"),
	}

	result := StringifyNode(node)
	assert.Contains(t, result, "@type: test_node")
	assert.Contains(t, result, "@token: test_content")
}

func TestStringifyNode_Nil(t *testing.T) {
	result := StringifyNode(nil)
	assert.Equal(t, "nil", result)
}

func TestConvertSitterNodeToUAST_TypeAndAttributeMapping(t *testing.T) {
	// TODO: This test requires a way to construct sitter.Node trees in memory.
	// The go-tree-sitter-bare API does not provide public setters for type/content/children.
	// Integration tests should be added when a suitable API or fixture is available.
	t.Skip("Cannot construct sitter.Node trees in memory with current API")
}

// TODO: Add more tests for non-nil nodes, children, and attributes
