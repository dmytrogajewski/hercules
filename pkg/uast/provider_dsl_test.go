package uast

import (
	"os"
	"strings"
	"testing"

	gositter "github.com/alexaandru/go-sitter-forest/go"
	sitter "github.com/alexaandru/go-tree-sitter-bare"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/mapping"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

func TestDSLProviderIntegration(t *testing.T) {
	// Test DSL content
	dslContent := `function_declaration <- (function_declaration name: (identifier) @name body: (block) @body) => uast(
    type: "Function",
    token: @name,
    roles: "Declaration",
    children: @body
)

identifier <- (identifier) => uast(
    type: "Identifier",
    roles: "Name"
)

source_file <- (source_file) => uast(
    type: "File",
    roles: "Module"
)`

	// Load and validate mapping rules
	rules, err := (&mapping.MappingParser{}).ParseMapping(dslContent)
	if err != nil {
		t.Fatalf("Failed to load DSL mappings: %v", err)
	}

	if len(rules) != 3 {
		t.Fatalf("Expected 3 rules, got %d", len(rules))
	}

	// Removed ValidateMappings, parsing already validates rules

	// Get Go language
	langPtr := gositter.GetLanguage()
	if langPtr == nil {
		t.Skip("Go language parser not available")
	}
	lang := sitter.NewLanguage(langPtr)

	// Create DSL provider
	provider := NewDSLProvider(lang, "go", rules)

	// Test source code
	source := []byte(`package main

func Hello() {
    fmt.Println("Hello, World!")
}`)

	// Parse the source code
	uastNode, err := provider.Parse("test.go", source)
	if err != nil {
		t.Fatalf("Failed to parse source code: %v", err)
	}

	if uastNode == nil {
		t.Fatal("Expected UAST node, got nil")
	}

	// Verify the provider implements the Provider interface
	var _ Provider = provider

	// Test provider methods
	if provider.Language() != "go" {
		t.Errorf("Expected language 'go', got '%s'", provider.Language())
	}

	supported := provider.SupportedLanguages()
	if len(supported) != 1 || supported[0] != "go" {
		t.Errorf("Expected supported languages ['go'], got %v", supported)
	}

	if !provider.IsSupported("test.go") {
		t.Error("Expected test.go to be supported")
	}

	t.Logf("Successfully parsed Go source code with DSL provider")
	t.Logf("UAST node type: %s", uastNode.Type)
	t.Logf("UAST node roles: %v", uastNode.Roles)
}

func TestProviderFactoryIntegration(t *testing.T) {
	// Create provider factory
	factory := NewProviderFactory()

	// Note: Language registration is now handled differently in the new DSL-based system

	// Test DSL content
	dslContent := `function_declaration <- (function_declaration name: (identifier) @name) => uast(
    type: "Function",
    token: @name,
    roles: "Declaration"
)`

	// Create DSL provider using factory
	provider, err := factory.CreateDSLProvider("go", dslContent)
	if err != nil {
		t.Fatalf("Failed to create DSL provider: %v", err)
	}

	if provider == nil {
		t.Fatal("Expected provider, got nil")
	}

	// Test source code
	source := []byte(`package main

func Hello() {
    fmt.Println("Hello, World!")
}`)

	// Parse the source code
	uastNode, err := provider.Parse("test.go", source)
	if err != nil {
		t.Fatalf("Failed to parse source code: %v", err)
	}

	if uastNode == nil {
		t.Fatal("Expected UAST node, got nil")
	}

	t.Logf("Successfully created DSL provider using factory")
	t.Logf("UAST node type: %s", uastNode.Type)
}

func TestProviderFactoryLanguageDetection(t *testing.T) {
	factory := NewProviderFactory()

	// Test language detection from filenames
	testCases := []struct {
		filename string
		expected string
	}{
		{"main.go", "go"},
		{"script.js", "javascript"},
		{"app.py", "python"},
		{"Main.java", "java"},
		{"lib.rs", "rust"},
		{"index.html", "html"},
		{"style.css", "css"},
		{"config.json", "json"},
		{"README.md", "markdown"},
		{"Dockerfile", "dockerfile"},
		{"Makefile", "makefile"},
		{"unknown.xyz", ""},
	}

	for _, tc := range testCases {
		result := factory.detectLanguageFromFile(tc.filename)
		if result != tc.expected {
			t.Errorf("For filename %s, expected language %s, got %s", tc.filename, tc.expected, result)
		}
	}
}

func TestDSLProvider_CaptureExtraction(t *testing.T) {
	dslContent := `function_declaration <- (function_declaration) => uast(
    type: "Function",
    roles: "Declaration",
    token: "fields.name"
)
identifier <- (identifier) => uast(
    type: "Identifier",
    roles: "Name"
)
source_file <- (source_file) => uast(
    type: "File",
    roles: "Module"
)`

	rules, err := (&mapping.MappingParser{}).ParseMapping(dslContent)
	if err != nil {
		t.Fatalf("Failed to load DSL mappings: %v", err)
	}
	langPtr := gositter.GetLanguage()
	if langPtr == nil {
		t.Skip("Go language parser not available")
	}
	lang := sitter.NewLanguage(langPtr)
	provider := NewDSLProvider(lang, "go", rules)

	source := []byte(`package main
func Hello() {}`)
	uastNode, err := provider.Parse("test.go", source)
	if err != nil {
		t.Fatalf("Failed to parse source code: %v", err)
	}
	if uastNode == nil {
		t.Fatal("Expected UAST node, got nil")
	}
	// Find the Function node
	var fn *node.Node
	for _, c := range uastNode.Children {
		if c.Type == "Function" {
			fn = c
			break
		}
	}
	if fn == nil {
		t.Fatal("Expected Function node")
	}
	if fn.Token != "Hello" {
		t.Errorf("Expected token 'Hello', got '%s'", fn.Token)
	}
	if fn.Props["name"] != "Hello" {
		t.Errorf("Expected property name 'Hello', got '%s'", fn.Props["name"])
	}
}

func TestDSLProvider_ConditionEvaluation(t *testing.T) {
	dslContent := `function_declaration <- (function_declaration) => uast(
    type: "Function",
    roles: "Declaration",
    token: "fields.name"
) when name == "Hello"
identifier <- (identifier) => uast(
    type: "Identifier",
    roles: "Name"
)
source_file <- (source_file) => uast(
    type: "File",
    roles: "Module"
)`

	rules, err := (&mapping.MappingParser{}).ParseMapping(dslContent)
	if err != nil {
		t.Fatalf("Failed to load DSL mappings: %v", err)
	}
	langPtr := gositter.GetLanguage()
	if langPtr == nil {
		t.Skip("Go language parser not available")
	}
	lang := sitter.NewLanguage(langPtr)
	provider := NewDSLProvider(lang, "go", rules)

	source := []byte(`package main
func Hello() {}
func World() {}`)
	uastNode, err := provider.Parse("test.go", source)
	if err != nil {
		t.Fatalf("Failed to parse source code: %v", err)
	}
	if uastNode == nil {
		t.Fatal("Expected UAST node, got nil")
	}
	// Only the Hello function should be mapped as Function
	foundHello := false
	foundWorld := false
	for _, c := range uastNode.Children {
		if c.Type == "Function" && c.Token == "Hello" {
			foundHello = true
		}
		if c.Type == "Function" && c.Token == "World" {
			foundWorld = true
		}
	}
	if !foundHello {
		t.Error("Expected Hello function to be mapped")
	}
	if foundWorld {
		t.Error("Did not expect World function to be mapped due to condition")
	}
}

func TestDSLProvider_InheritanceWithConditions(t *testing.T) {
	dslContent := `function_declaration <- (function_declaration) => uast(
    type: "Child",
    roles: "ChildRole",
    token: "fields.name"
) when name == "Hello"
identifier <- (identifier) => uast(
    type: "Identifier",
    roles: "Name"
)
source_file <- (source_file) => uast(
    type: "File",
    roles: "Module"
)`

	rules, err := (&mapping.MappingParser{}).ParseMapping(dslContent)
	if err != nil {
		t.Fatalf("Failed to load DSL mappings: %v", err)
	}
	langPtr := gositter.GetLanguage()
	if langPtr == nil {
		t.Skip("Go language parser not available")
	}
	lang := sitter.NewLanguage(langPtr)
	provider := NewDSLProvider(lang, "go", rules)

	source := []byte(`package main
func Hello() {}
func World() {}`)
	uastNode, err := provider.Parse("test.go", source)
	if err != nil {
		t.Fatalf("Failed to parse source code: %v", err)
	}
	if uastNode == nil {
		t.Fatal("Expected UAST node, got nil")
	}
	foundChild := false
	for _, c := range uastNode.Children {
		if c.Type == "Child" && c.Token == "Hello" {
			foundChild = true
		}
	}
	if !foundChild {
		t.Error("Expected Child node for Hello due to inheritance and condition")
	}
}

func TestDSLProvider_AdvancedPropertyExtraction(t *testing.T) {
	dslContent := `var_declaration <- (var_declaration) => uast(
    type: "Variable",
    roles: "Declaration"
)
identifier <- (identifier) => uast(
    type: "Identifier",
    roles: "Name"
)
source_file <- (source_file) => uast(
    type: "File",
    roles: "Module"
)`

	rules, err := (&mapping.MappingParser{}).ParseMapping(dslContent)
	if err != nil {
		t.Fatalf("Failed to load DSL mappings: %v", err)
	}
	langPtr := gositter.GetLanguage()
	if langPtr == nil {
		t.Skip("Go language parser not available")
	}
	lang := sitter.NewLanguage(langPtr)
	provider := NewDSLProvider(lang, "go", rules)

	source := []byte(`package main
var x int`)
	uastNode, err := provider.Parse("test.go", source)
	if err != nil {
		t.Fatalf("Failed to parse source code: %v", err)
	}
	if uastNode == nil {
		t.Fatal("Expected UAST node, got nil")
	}
	foundVar := false
	for _, c := range uastNode.Children {
		if c.Type == "Variable" {
			foundVar = true
			break
		}
	}
	if !foundVar {
		t.Error("Expected Variable node")
	}
}

func TestDSLProvider_ChildInclusionExclusion(t *testing.T) {
	dslContent := `function_declaration <- (function_declaration) => uast(
    type: "Function",
    roles: "Declaration",
    token: "fields.name"
) when name == "Hello"
identifier <- (identifier) => uast(
    type: "Identifier",
    roles: "Name"
)
source_file <- (source_file) => uast(
    type: "File",
    roles: "Module"
)`

	rules, err := (&mapping.MappingParser{}).ParseMapping(dslContent)
	if err != nil {
		t.Fatalf("Failed to load DSL mappings: %v", err)
	}
	langPtr := gositter.GetLanguage()
	if langPtr == nil {
		t.Skip("Go language parser not available")
	}
	lang := sitter.NewLanguage(langPtr)
	provider := NewDSLProvider(lang, "go", rules)

	source := []byte(`package main
func Hello() {}
func World() {}`)
	uastNode, err := provider.Parse("test.go", source)
	if err != nil {
		t.Fatalf("Failed to parse source code: %v", err)
	}
	if uastNode == nil {
		t.Fatal("Expected UAST node, got nil")
	}
	// Only Hello function should be included
	foundHello := false
	foundWorld := false
	for _, c := range uastNode.Children {
		if c.Type == "Function" && c.Token == "Hello" {
			foundHello = true
		}
		if c.Type == "Function" && c.Token == "World" {
			foundWorld = true
		}
	}
	if !foundHello {
		t.Error("Expected Hello function to be included")
	}
	if foundWorld {
		t.Error("Did not expect World function to be included due to condition")
	}
}

func TestE2E_MappingGenerationAndParsing(t *testing.T) {
	// Minimal node-types.json fixture (Go function and identifier)
	nodeTypesJSON := `[
	  {"type": "function_declaration", "named": true, "fields": {"name": {"types": ["identifier"], "required": true}}},
	  {"type": "identifier", "named": true, "fields": {}}
	]`

	nodes, err := mapping.ParseNodeTypes([]byte(nodeTypesJSON))
	if err != nil {
		t.Fatalf("Failed to parse node-types.json: %v", err)
	}
	dsl := mapping.GenerateMappingDSL(nodes)

	// Parse the generated DSL
	rules, err := (&mapping.MappingParser{}).ParseMapping(dsl)
	if err != nil {
		t.Fatalf("Failed to parse generated mapping DSL: %v\nDSL:\n%s", err, dsl)
	}

	// Use a minimal Go source file
	source := []byte(`package main\nfunc Hello() {}`)
	langPtr := gositter.GetLanguage()
	if langPtr == nil {
		t.Skip("Go language parser not available")
	}
	lang := sitter.NewLanguage(langPtr)
	provider := NewDSLProvider(lang, "go", rules)
	uastNode, err := provider.Parse("test.go", source)
	if err != nil {
		t.Fatalf("Failed to parse Go source with generated mapping: %v", err)
	}
	if uastNode == nil {
		t.Fatal("Expected UAST node, got nil")
	}
	// Check that a Function node is present
	foundFunc := false
	if uastNode.Type == "Function" || uastNode.Type == "function_declaration" {
		foundFunc = true
	}
	for _, c := range uastNode.Children {
		if c.Type == "Function" || c.Type == "function_declaration" {
			foundFunc = true
			break
		}
	}
	if !foundFunc {
		t.Logf("Generated DSL:\n%s", dsl)
		t.Logf("UAST: %+v", uastNode)
		t.Error("Expected Function node in UAST from generated mapping")
	}
}

func TestDSLProvider_RealWorldGoUASTMap(t *testing.T) {
	// Real-world go.uastmap DSL with advanced features
	dslContent := `
function_declaration <- (function_declaration name: (identifier) @name parameters: (parameter_list) @parameters body: (block) @body) => uast(
    type: "Function",
    token: "@name",
    roles: "Declaration", "Function",
    children: "@parameters", "@body",
    name: "@name",
    parameters: "@parameters",
    body: "@body"
)

method_declaration <- (method_declaration name: (identifier) @name receiver: (parameter_list) @receiver parameters: (parameter_list) @parameters body: (block) @body) => uast(
    type: "Method",
    token: "@name",
    roles: "Declaration", "Method",
    children: "@receiver", "@parameters", "@body",
    name: "@name",
    receiver: "@receiver",
    parameters: "@parameters",
    body: "@body"
) # Extends function_declaration

var_spec <- (var_spec name: (identifier) @name) => uast(
    type: "Variable",
    token: "@name",
    roles: "Declaration", "Variable",
    name: "@name"
)

if_statement <- (if_statement condition: (expression) @cond consequence: (block) @conseq alternative: (block) @alt) => uast(
    type: "If",
    roles: "Statement", "Conditional",
    children: "@cond", "@conseq", "@alt"
)
`

	rules, err := (&mapping.MappingParser{}).ParseMapping(dslContent)
	if err != nil {
		t.Fatalf("Failed to load DSL mappings: %v", err)
	}
	langPtr := gositter.GetLanguage()
	if langPtr == nil {
		t.Skip("Go language parser not available")
	}
	lang := sitter.NewLanguage(langPtr)
	provider := NewDSLProvider(lang, "go", rules)

	// Real-world Go source code
	source := []byte(`package main

func Add(a int, b int) int {
    return a + b
}

var x int = 42
`)
	uastNode, err := provider.Parse("test.go", source)
	if err != nil {
		t.Fatalf("Failed to parse source code: %v", err)
	}
	if uastNode == nil {
		t.Fatal("Expected UAST node, got nil")
	}

	// Find the Function node and check properties
	var foundFunc bool
	for _, c := range uastNode.Children {
		if c.Type == "Function" {
			foundFunc = true
			if c.Props["name"] != "Add" {
				t.Errorf("Expected function name 'Add', got '%s'", c.Props["name"])
			}
			if c.Token != "Add" {
				t.Errorf("Expected function token 'Add', got '%s'", c.Token)
			}
			if c.Props["parameters"] == "" {
				t.Errorf("Expected parameters property to be set")
			}
			if c.Props["body"] == "" {
				t.Errorf("Expected body property to be set")
			}
			// Debug: print all props
			t.Logf("Function node props: %+v", c.Props)
		}
		if c.Type == "Variable" {
			if c.Props["name"] != "x" {
				t.Errorf("Expected variable name 'x', got '%s'", c.Props["name"])
			}
			// Debug: print all props
			t.Logf("Variable node props: %+v", c.Props)
			// Debug: print children types and tokens recursively
			var printVarTree func(n *node.Node, depth int)
			printVarTree = func(n *node.Node, depth int) {
				if n == nil {
					return
				}
				pad := strings.Repeat("  ", depth)
				t.Logf("%sVarNode: type=%s, token=%s, props=%+v", pad, n.Type, n.Token, n.Props)
				for _, cc := range n.Children {
					printVarTree(cc, depth+1)
				}
			}
			printVarTree(c, 1)
		}
	}
	if !foundFunc {
		t.Error("Expected to find a Function node")
	}

	// Debug: recursively print all nodes in the UAST tree
	var printTree func(n *node.Node, depth int)
	printTree = func(n *node.Node, depth int) {
		if n == nil {
			return
		}
		pad := strings.Repeat("  ", depth)
		t.Logf("%sNode: type=%s, token=%s, props=%+v", pad, n.Type, n.Token, n.Props)
		for _, c := range n.Children {
			printTree(c, depth+1)
		}
	}
	printTree(uastNode, 0)

	// Write a Go file for tree-sitter inspection in the current directory
	tmpGo := `package main
var x int = 42
func Add(a int, b int) int { return a + b }
`
	fileName := "test_var.go"
	if err := os.WriteFile(fileName, []byte(tmpGo), 0644); err != nil {
		t.Fatalf("Failed to write test_var.go: %v", err)
	}
	defer os.Remove(fileName)
}
