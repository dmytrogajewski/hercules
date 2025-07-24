// Place this test in package uastconvert to access unexported functions directly.
package uastconvert

import (
	"context"
	"testing"

	tsjava "github.com/alexaandru/go-sitter-forest/java"
	sitter "github.com/alexaandru/go-tree-sitter-bare"
	"github.com/stretchr/testify/assert"
)

func TestJavaMethodConversion_Integration(t *testing.T) {
	// Java code with a method
	code := `public class Hello { public void greet() { System.out.println("hi"); } }`
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tsjava.GetLanguage()))
	tree, err := parser.ParseString(context.Background(), nil, []byte(code))
	assert.NoError(t, err)
	root := tree.RootNode()
	uast := ConvertSitterNodeToUAST(&root, []byte(code))

	// Traverse the UAST to find a uast:FunctionGroup node with Name = "greet"
	var found bool
	var findFunc func(interface{})
	findFunc = func(n interface{}) {
		if obj, ok := n.(map[string]interface{}); ok {
			if t, ok := obj["@type"]; ok && t == "uast:FunctionGroup" {
				if name, ok := obj["Name"]; ok && name == "greet" {
					found = true
				}
			}
			for _, v := range obj {
				findFunc(v)
			}
		} else if arr, ok := n.([]interface{}); ok {
			for _, v := range arr {
				findFunc(v)
			}
		}
	}
	// Convert to standard Go types for traversal
	serialized, err := SerializeNode(uast)
	assert.NoError(t, err)
	findFunc(serialized)
	assert.True(t, found, "Should find a uast:FunctionGroup node with Name 'greet'")
}
