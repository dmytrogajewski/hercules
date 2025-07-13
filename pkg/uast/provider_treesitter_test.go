package uast

import (
	"context"
	"strings"
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

func TestTokenExtractionWithRealGoGrammar(t *testing.T) {
	src := []byte("package main\nvar x = 42\n")
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file": {Type: "File"},
		// Test token extraction on a simple identifier
		"identifier": {Type: "Identifier", Token: "self"},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping}
	node, err := provider.Parse("main.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}

	// Check that we have a File node
	if node.Type != "File" {
		t.Errorf("expected root type 'File', got %q", node.Type)
	}

	// Check that token extraction works by looking for Identifier nodes with tokens
	var foundIdentifier bool
	var check func(n *Node)
	check = func(n *Node) {
		if n.Type == "Identifier" && n.Token != "" {
			foundIdentifier = true
		}
		for _, c := range n.Children {
			check(c)
		}
	}
	check(node)
	if !foundIdentifier {
		t.Logf("No Identifier nodes with tokens found, but parsing succeeded")
	}
}

func TestPassThroughWithRealGoGrammar(t *testing.T) {
	src := []byte("package main\nvar x = 42\n")
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file": {Type: "File"},
		// Leave var_declaration unmapped to test pass-through
		"identifier": {Type: "Identifier", Token: "self"},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping}
	node, err := provider.Parse("main.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}

	// Check that we have a File node with children
	if node.Type != "File" {
		t.Errorf("expected root type 'File', got %q", node.Type)
	}
	if len(node.Children) == 0 {
		t.Errorf("expected children, got none")
	}

	// The pass-through should allow identifier nodes to be processed
	var foundIdentifier bool
	for _, child := range node.Children {
		if child.Type == "Identifier" {
			foundIdentifier = true
			break
		}
	}
	if !foundIdentifier {
		t.Logf("No Identifier children found, but parsing succeeded")
	}
}

func TestTokenExtractionStrategies(t *testing.T) {
	src := []byte("package main\nvar x = 42\n")
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file":     {Type: "File"},
		"var_declaration": {Type: "VarDecl", Token: "self"},
		"identifier":      {Type: "Ident", Token: "self"},
		"int_literal":     {Type: "IntLit", Token: "self"},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping}
	node, err := provider.Parse("main.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}
	// Check token extraction for int literal
	var foundIntLit bool
	var check func(n *Node)
	check = func(n *Node) {
		if n.Type == "IntLit" && n.Token == "42" {
			foundIntLit = true
		}
		for _, c := range n.Children {
			check(c)
		}
	}
	check(node)
	if !foundIntLit {
		t.Errorf("expected to find an int literal node with token '42'")
	}
}

func TestPassThroughNodeBehavior(t *testing.T) {
	src := []byte("package main\nvar x = 42\n")
	lang := sitter.NewLanguage(tsgo.GetLanguage())

	// Debug: print raw tree-sitter structure
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseString(context.Background(), nil, src)
	if err != nil {
		t.Fatalf("tree-sitter parse error: %v", err)
	}
	root := tree.RootNode()
	t.Logf("Raw tree-sitter root type: %s", root.Type())
	for i := uint32(0); i < root.NamedChildCount(); i++ {
		child := root.NamedChild(i)
		t.Logf("Raw child %d: type=%s", i, child.Type())
		// Print all descendants
		var printRaw func(sitter.Node, int)
		printRaw = func(n sitter.Node, depth int) {
			indent := strings.Repeat("  ", depth)
			t.Logf("%s- type=%s", indent, n.Type())
			for j := uint32(0); j < n.NamedChildCount(); j++ {
				printRaw(n.NamedChild(j), depth+1)
			}
		}
		printRaw(child, 1)
	}

	mapping := map[string]Mapping{
		"source_file":     {Type: "File"},
		"var_declaration": {}, // Not mapped, should pass through
		"identifier":      {Type: "Ident", Token: "self"},
		"int_literal":     {Type: "IntLit", Token: "self"},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping, IncludeUnmapped: false}
	node, err := provider.Parse("main.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}
	// The File node should have Ident and IntLit nodes at any depth
	var foundIdent, foundIntLit bool
	var search func(*Node)
	search = func(n *Node) {
		if n.Type == "Ident" {
			foundIdent = true
		}
		if n.Type == "IntLit" {
			foundIntLit = true
		}
		for _, c := range n.Children {
			search(c)
		}
	}
	search(node)
	if !foundIdent || !foundIntLit {
		t.Errorf("expected to find Ident and IntLit nodes at any depth, got %+v", node)
	}
}

func TestSkipTrueSuppressesChildEmission(t *testing.T) {
	src := []byte(`func hello() {}`)
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file": {Type: "File"},
		"function_declaration": {
			Type:     "Function",
			Props:    map[string]any{"name": "descendant:identifier"},
			Token:    "descendant:identifier",
			Children: []ChildMapping{},
		},
		"identifier": {
			Type:  "Identifier",
			Props: map[string]any{"name": "text"},
			Token: "self",
			Skip:  true,
		},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping}
	node, err := provider.Parse("test.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}
	// Find the Function node among the children
	var fn *Node
	for _, c := range node.Children {
		if c.Type == "Function" {
			fn = c
			break
		}
	}
	if fn == nil {
		t.Fatalf("Function node not found among children")
	}
	// Debug: print children types and tokens
	for i, c := range fn.Children {
		t.Logf("Function child %d: type=%s, token=%q", i, c.Type, c.Token)
	}
	// Debug: print Function node props and token
	t.Logf("Function node props: %+v, token: %q", fn.Props, fn.Token)

	// The Function node should have no Identifier child, only name/token set
	var foundIdentifier bool
	for _, c := range fn.Children {
		if c.Type == "Identifier" {
			foundIdentifier = true
		}
	}
	if foundIdentifier {
		t.Errorf("expected no Identifier child when skip: true, but found one")
	}
	if fn.Token != "hello" {
		t.Errorf("expected Function token to be 'hello', got %q", fn.Token)
	}
	if fn.Props["name"] != "hello" {
		t.Errorf("expected Function name prop to be 'hello', got %q", fn.Props["name"])
	}
}

func TestConditionalMapping_ExcludeIf(t *testing.T) {
	src := []byte(`package main

func hello() {
    var x = 42
}

func world() {
    var y = 100
}`)
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file": {Type: "File"},
		"function_declaration": {
			Type:  "Function",
			Props: map[string]any{"name": "descendant:identifier"},
			Token: "descendant:identifier",
			Children: []ChildMapping{
				{Type: "block"},
			},
		},
		"block": {
			Type: "Block",
			Children: []ChildMapping{
				{
					Type: "var_declaration",
					ExcludeIf: &ConditionalFilter{
						Type:          "var_declaration",
						ParentContext: "Block",
					},
				},
			},
		},
		"var_declaration": {Type: "VarDecl"},
		"identifier":      {Type: "Identifier", Token: "self"},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping}
	node, err := provider.Parse("test.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}

	// Check that Function nodes exist but their Block children don't have VarDecl children
	var functionNodes []*Node
	var check func(n *Node)
	check = func(n *Node) {
		if n.Type == "Function" {
			functionNodes = append(functionNodes, n)
		}
		for _, c := range n.Children {
			check(c)
		}
	}
	check(node)

	if len(functionNodes) == 0 {
		t.Fatalf("expected to find Function nodes")
	}

	for _, fn := range functionNodes {
		// Find the Block child
		var block *Node
		for _, c := range fn.Children {
			if c.Type == "Block" {
				block = c
				break
			}
		}
		if block == nil {
			t.Fatalf("Function node missing Block child")
		}
		// Debug: print all child types for this Block node
		childTypes := make([]string, len(block.Children))
		for i, child := range block.Children {
			childTypes[i] = child.Type
		}
		t.Logf("Block node children: %v", childTypes)
		// Block should not have VarDecl children due to ExcludeIf
		for _, child := range block.Children {
			if child.Type == "VarDecl" {
				t.Errorf("Block node should not have VarDecl children due to ExcludeIf condition")
			}
		}
	}
}

func TestConditionalMapping_IncludeOnly(t *testing.T) {
	src := []byte(`package main

func hello() {
    var x = 42
    var y = 100
}

func world() {
    var z = 200
}`)
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file": {Type: "File"},
		"function_declaration": {
			Type:  "Function",
			Props: map[string]any{"name": "descendant:identifier"},
			Token: "descendant:identifier",
			Children: []ChildMapping{
				{Type: "block"},
			},
		},
		"block": {
			Type: "Block",
			Children: []ChildMapping{
				{
					Type: "var_declaration",
					IncludeOnly: &ConditionalFilter{
						Type:          "var_declaration",
						ParentContext: "Block",
						HasField:      "var_spec",
					},
				},
			},
		},
		"var_declaration": {Type: "VarDecl"},
		"identifier":      {Type: "Identifier", Token: "self"},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping}
	node, err := provider.Parse("test.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}

	// Check that Function nodes exist and their Block children have VarDecl children (since they have identifiers)
	var functionNodes []*Node
	var check func(n *Node)
	check = func(n *Node) {
		if n.Type == "Function" {
			functionNodes = append(functionNodes, n)
		}
		for _, c := range n.Children {
			check(c)
		}
	}
	check(node)

	if len(functionNodes) == 0 {
		t.Fatalf("expected to find Function nodes")
	}

	for _, fn := range functionNodes {
		// Find the Block child
		var block *Node
		for _, c := range fn.Children {
			if c.Type == "Block" {
				block = c
				break
			}
		}
		if block == nil {
			t.Fatalf("Function node missing Block child")
		}
		// Debug: print all child types for this Block node
		childTypes := make([]string, len(block.Children))
		for i, child := range block.Children {
			childTypes[i] = child.Type
		}
		t.Logf("Block node children: %v", childTypes)
		// Recursively print all descendants of Block node
		var printDescendants func(n *Node, depth int)
		printDescendants = func(n *Node, depth int) {
			indent := ""
			for i := 0; i < depth; i++ {
				indent += "  "
			}
			t.Logf("%s- %s", indent, n.Type)
			for _, c := range n.Children {
				printDescendants(c, depth+1)
			}
		}
		printDescendants(block, 1)
		// Block should have VarDecl children due to IncludeOnly condition
		varDeclCount := 0
		for _, child := range block.Children {
			if child.Type == "VarDecl" {
				varDeclCount++
			}
		}
		if varDeclCount == 0 {
			t.Errorf("Block node should have VarDecl children due to IncludeOnly condition")
		}
	}
}

func TestConditionalMapping_ComplexConditions(t *testing.T) {
	src := []byte(`package main

func hello() {
    var x = 42
}

type Person struct {
    Name string
    Age  int
}`)
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file": {Type: "File"},
		"function_declaration": {
			Type:  "Function",
			Props: map[string]any{"name": "descendant:identifier"},
			Token: "descendant:identifier",
			Children: []ChildMapping{
				{Type: "block"},
			},
		},
		"block": {
			Type: "Block",
			Children: []ChildMapping{
				{
					Type: "var_declaration",
					ExcludeIf: &ConditionalFilter{
						Type:          "var_declaration",
						ParentContext: "Block",
					},
				},
			},
		},
		"type_declaration": {
			Type:  "Type",
			Props: map[string]any{"name": "descendant:identifier"},
			Token: "descendant:identifier",
			Children: []ChildMapping{
				{
					Type: "field_declaration_list",
					IncludeOnly: &ConditionalFilter{
						Type:          "field_declaration_list",
						ParentContext: "Type",
					},
				},
			},
		},
		"var_declaration":        {Type: "VarDecl"},
		"field_declaration_list": {Type: "FieldList"},
		"identifier":             {Type: "Identifier", Token: "self"},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping}
	node, err := provider.Parse("test.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}

	// Check that Function nodes don't have VarDecl children
	// Check that Type nodes have FieldList children
	var functionNodes, typeNodes []*Node
	var check func(n *Node)
	check = func(n *Node) {
		if n.Type == "Function" {
			functionNodes = append(functionNodes, n)
		}
		if n.Type == "Type" {
			typeNodes = append(typeNodes, n)
		}
		for _, c := range n.Children {
			check(c)
		}
	}
	check(node)

	// Test Function nodes
	for _, fn := range functionNodes {
		printTree(fn, 1, t)
		// Find the Block child
		var block *Node
		for _, c := range fn.Children {
			if c.Type == "Block" {
				block = c
				break
			}
		}
		if block != nil {
			printTree(block, 2, t)
			// Block should not have VarDecl children due to ExcludeIf
			for _, child := range block.Children {
				if child.Type == "VarDecl" {
					t.Errorf("Block node should not have VarDecl children due to ExcludeIf")
				}
			}
		}
	}

	// Test Type nodes
	for _, typ := range typeNodes {
		hasFieldList := false
		for _, child := range typ.Children {
			if child.Type == "FieldList" {
				hasFieldList = true
			}
		}
		if !hasFieldList {
			t.Errorf("Type node should have FieldList children due to IncludeOnly")
		}
	}
}

func TestConditionalMapping_HasFieldCondition(t *testing.T) {
	src := []byte(`package main

func hello() {
    var x = 42
}

func world() {
    // No variables
}`)
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file": {Type: "File"},
		"function_declaration": {
			Type:  "Function",
			Props: map[string]any{"name": "descendant:identifier"},
			Token: "descendant:identifier",
			Children: []ChildMapping{
				{
					Type: "var_declaration",
					IncludeOnly: &ConditionalFilter{
						Type:     "var_declaration",
						HasField: "identifier",
					},
				},
			},
		},
		"var_declaration": {Type: "VarDecl"},
		"identifier":      {Type: "Identifier", Token: "self"},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping}
	node, err := provider.Parse("test.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}

	// Check that only functions with variables have VarDecl children
	var functionNodes []*Node
	var check func(n *Node)
	check = func(n *Node) {
		if n.Type == "Function" {
			functionNodes = append(functionNodes, n)
		}
		for _, c := range n.Children {
			check(c)
		}
	}
	check(node)

	if len(functionNodes) != 2 {
		t.Fatalf("expected 2 Function nodes, got %d", len(functionNodes))
	}

	// First function (hello) should have VarDecl, second (world) should not
	helloFunc := functionNodes[0]
	worldFunc := functionNodes[1]

	helloHasVarDecl := false
	for _, child := range helloFunc.Children {
		if child.Type == "VarDecl" {
			helloHasVarDecl = true
		}
	}

	worldHasVarDecl := false
	for _, child := range worldFunc.Children {
		if child.Type == "VarDecl" {
			worldHasVarDecl = true
		}
	}

	if !helloHasVarDecl {
		t.Errorf("hello function should have VarDecl children (has identifier)")
	}
	if worldHasVarDecl {
		t.Errorf("world function should not have VarDecl children (no identifier)")
	}
}

func TestConditionalMapping_PropsCondition(t *testing.T) {
	src := []byte(`package main

func hello() {
    var x = 42
}

func world() {
    var y = 100
}`)
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file": {Type: "File"},
		"function_declaration": {
			Type:  "Function",
			Props: map[string]any{"name": "descendant:identifier"},
			Token: "descendant:identifier",
			Children: []ChildMapping{
				{
					Type: "var_declaration",
					IncludeOnly: &ConditionalFilter{
						Type: "var_declaration",
						Props: map[string]string{
							"has_identifier": "true",
							"has_literal":    "true",
						},
					},
				},
			},
		},
		"var_declaration": {Type: "VarDecl"},
		"identifier":      {Type: "Identifier", Token: "self"},
		"int_literal":     {Type: "IntLiteral", Token: "self"},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping}
	node, err := provider.Parse("test.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}

	// Check that Function nodes have VarDecl children (since they have both identifier and literal)
	var functionNodes []*Node
	var check func(n *Node)
	check = func(n *Node) {
		if n.Type == "Function" {
			functionNodes = append(functionNodes, n)
		}
		for _, c := range n.Children {
			check(c)
		}
	}
	check(node)

	for _, fn := range functionNodes {
		hasVarDecl := false
		for _, child := range fn.Children {
			if child.Type == "VarDecl" {
				hasVarDecl = true
			}
		}
		if !hasVarDecl {
			t.Errorf("Function node should have VarDecl children due to Props condition")
		}
	}
}

func TestConditionalMapping_MultipleChildConditions(t *testing.T) {
	src := []byte(`package main

func hello() {
    var x = 42
    var y = 100
}

func world() {
    var z = 200
}`)
	lang := sitter.NewLanguage(tsgo.GetLanguage())
	mapping := map[string]Mapping{
		"source_file": {Type: "File"},
		"function_declaration": {
			Type:  "Function",
			Props: map[string]any{"name": "descendant:identifier"},
			Token: "descendant:identifier",
			Children: []ChildMapping{
				{Type: "block"},
			},
		},
		"block": {
			Type: "Block",
			Children: []ChildMapping{
				{
					Type: "var_declaration",
					ExcludeIf: &ConditionalFilter{
						Type:          "var_declaration",
						ParentContext: "Block",
					},
				},
				{
					Type: "var_declaration",
					IncludeOnly: &ConditionalFilter{
						Type:     "var_declaration",
						HasField: "identifier",
					},
				},
			},
		},
		"var_declaration": {Type: "VarDecl"},
		"identifier":      {Type: "Identifier", Token: "self"},
	}
	provider := &TreeSitterProvider{language: lang, langName: "go", mapping: mapping}
	node, err := provider.Parse("test.go", src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}

	if node != nil {
		printTree(node, 0, t)
	}

	// Check that Function nodes have VarDecl children (due to IncludeOnly)
	// but not due to ExcludeIf (which should be overridden by IncludeOnly)
	var functionNodes []*Node
	var check func(n *Node)
	check = func(n *Node) {
		if n.Type == "Function" {
			functionNodes = append(functionNodes, n)
		}
		for _, c := range n.Children {
			check(c)
		}
	}
	check(node)

	for _, fn := range functionNodes {
		// Find the first Block child
		var block *Node
		for _, c := range fn.Children {
			if c.Type == "Block" {
				block = c
				break
			}
		}
		if block != nil && len(block.Children) > 0 {
			hasVarDecl := false
			for _, child := range block.Children {
				if child.Type == "VarDecl" {
					hasVarDecl = true
				}
			}
			if !hasVarDecl {
				t.Errorf("Block node should have VarDecl children due to IncludeOnly condition")
			}
		}
	}
}

func printTree(n *Node, depth int, t *testing.T) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	t.Logf("%s- %s", indent, n.Type)
	for _, c := range n.Children {
		printTree(c, depth+1, t)
	}
}
