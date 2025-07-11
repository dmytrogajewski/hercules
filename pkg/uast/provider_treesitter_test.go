package uast

import (
	"testing"

	tsgo "github.com/alexaandru/go-sitter-forest/go"
	sitter "github.com/alexaandru/go-tree-sitter-bare"
)

func TestTreeSitterProvider_Parse_Go(t *testing.T) {
	src := []byte("package main\nfunc hello() {}")
	provider := &TreeSitterProvider{
		language: sitter.NewLanguage(tsgo.GetLanguage()),
		langName: "go",
		mapping: map[string]Mapping{
			"source_file": {
				Type: "File",
			},
			"function_declaration": {
				Type:  "Function",
				Roles: []string{"Function", "Declaration"},
				Props: map[string]any{"name": "Name"},
			},
		},
	}
	node, err := provider.Parse("main.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}
	if node.Type != "File" {
		t.Errorf("expected root type 'File', got %q", node.Type)
	}
	if len(node.Children) == 0 {
		t.Errorf("expected children, got none")
	}
	foundFunc := false
	for _, child := range node.Children {
		if child.Type == "Function" {
			foundFunc = true
			break
		}
	}
	if !foundFunc {
		t.Errorf("expected at least one child of type 'Function'")
	}
}

func TestEmbeddedProvider_LanguageDetectionAndParsing(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	tests := []struct {
		name         string
		filename     string
		content      string
		wantLang     string
		wantType     string
		wantToken    string
		wantChildren bool
		wantErr      bool
	}{
		{"Go file", "main.go", "package main\nfunc main() {}", "go", "File", "", true, false},
		{"Empty file", "empty.go", "", "go", "", "", false, false},
		{"Unsupported ext", "file.xxx", "hello", "", "", "", false, true},
		// For invalid code, expect a root node but allow zero children
		{"Invalid code", "broken.go", "func {", "go", "File", "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parser.Parse(tt.filename, []byte(tt.content))
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if tt.wantType == "" {
				if node != nil {
					t.Errorf("Expected nil node for empty file, got non-nil")
				}
				return
			}
			if node == nil {
				t.Errorf("Expected non-nil node")
				return
			}
			if node.Type != tt.wantType {
				t.Errorf("Type: got %q, want %q", node.Type, tt.wantType)
			}
			if node.Token != tt.wantToken {
				t.Errorf("Token: got %q, want %q", node.Token, tt.wantToken)
			}
			hasChildren := len(node.Children) > 0
			if hasChildren != tt.wantChildren {
				t.Errorf("Children: got %v, want %v", hasChildren, tt.wantChildren)
			}
		})
	}
}

func TestTreeSitterProvider_HybridMapping(t *testing.T) {
	src := []byte("package main\nvar x int\nfunc hello() {}\n")
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file":          {Type: "File"},
		"function_declaration": {Type: "Function"},
	}
	// Canonical mode: only mapped nodes
	provider := &TreeSitterProvider{
		language:        lang,
		langName:        "go",
		mapping:         mapping,
		IncludeUnmapped: false,
	}
	node, err := provider.Parse("main.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}
	for _, child := range node.Children {
		if child.Type == "go:var_declaration" {
			t.Errorf("unmapped node 'var_declaration' should not appear when IncludeUnmapped=false")
		}
	}
	// Hybrid mode: unmapped nodes included
	provider.IncludeUnmapped = true
	node, err = provider.Parse("main.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node (hybrid mode)")
	}
	foundUnmapped := false
	for _, child := range node.Children {
		if child.Type == "go:var_declaration" {
			foundUnmapped = true
		}
	}
	if !foundUnmapped {
		t.Errorf("expected unmapped node 'go:var_declaration' to appear when IncludeUnmapped=true")
	}
}
