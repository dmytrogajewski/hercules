package uast

import (
	"context"
	"testing"

	forest "github.com/alexaandru/go-sitter-forest"
	sitter "github.com/alexaandru/go-tree-sitter-bare"
)

func TestFromTree_Golden(t *testing.T) {
	tests := []struct {
		name           string
		lang           string
		code           string
		wantType       string
		wantChildTypes []string
	}{
		{
			name:           "Go simple function",
			lang:           "go",
			code:           "package main\nfunc main() {}",
			wantType:       "go:source_file",
			wantChildTypes: []string{"go:package_clause", "go:function_declaration"},
		},
		{
			name:           "Python def",
			lang:           "python",
			code:           "def foo():\n    pass",
			wantType:       "python:module",
			wantChildTypes: []string{"python:function_definition"},
		},
		{
			name:           "Java class",
			lang:           "java",
			code:           "public class Main { void foo() {} }",
			wantType:       "java:program",
			wantChildTypes: []string{"java:class_declaration"},
		},
		{
			name:           "Empty file",
			lang:           "go",
			code:           "",
			wantType:       "go:source_file",
			wantChildTypes: []string{},
		},
		{
			name:           "Syntax error",
			lang:           "go",
			code:           "func {",
			wantType:       "go:source_file",
			wantChildTypes: []string{"go:ERROR"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := sitter.NewParser()
			parser.SetLanguage(getSitterLanguage(tt.lang))
			tree, err := parser.ParseString(context.Background(), nil, []byte(tt.code))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			node := FromTree(tree, []byte(tt.code), tt.lang)
			if node == nil {
				t.Fatalf("FromTree returned nil")
			}
			if node.Type != tt.wantType {
				t.Errorf("Root type: got %q, want %q", node.Type, tt.wantType)
			}
			var gotChildTypes []string
			for _, c := range node.Children {
				gotChildTypes = append(gotChildTypes, c.Type)
			}
			if len(gotChildTypes) != len(tt.wantChildTypes) {
				t.Errorf("Child count: got %d, want %d", len(gotChildTypes), len(tt.wantChildTypes))
			}
			for i := range gotChildTypes {
				if i >= len(tt.wantChildTypes) {
					break
				}
				if gotChildTypes[i] != tt.wantChildTypes[i] {
					t.Errorf("Child %d type: got %q, want %q", i, gotChildTypes[i], tt.wantChildTypes[i])
				}
			}
		})
	}
}

func getSitterLanguage(lang string) *sitter.Language {
	return forest.GetLanguage(lang)
}
