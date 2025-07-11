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
				Type: "go:file",
			},
			"function_declaration": {
				Type:  "go:function",
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
	if node.Type != "go:file" {
		t.Errorf("expected root type 'go:file', got %q", node.Type)
	}
	if len(node.Children) == 0 {
		t.Errorf("expected children, got none")
	}
	foundFunc := false
	for _, child := range node.Children {
		if child.Type == "go:function" {
			foundFunc = true
			break
		}
	}
	if !foundFunc {
		t.Errorf("expected at least one child of type 'go:function'")
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
		{"Go file", "main.go", "package main\nfunc main() {}", "go", "go:file", "", true, false},
		{"Empty file", "empty.go", "", "go", "go:file", "", false, false},
		{"Unsupported ext", "file.txt", "hello", "", "", "", false, true},
		{"Invalid code", "broken.go", "func {", "go", "go:file", "", true, false},
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
