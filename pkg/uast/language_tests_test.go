package uast

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"context"

	sitter "github.com/alexaandru/go-tree-sitter-bare"
	"gopkg.in/yaml.v3"
)

//go:embed language_tests/**/*.yaml
var LanguageTestsFS embed.FS

// TestSuite represents a test suite loaded from YAML
type TestSuite struct {
	Suite      string      `yaml:"suite"`
	Name       string      `yaml:"name"`
	ParseCases []ParseCase `yaml:"parse_cases"`
	QueryCases []QueryCase `yaml:"query_cases"`
}

// ParseCase represents a single parse test case
type ParseCase struct {
	Name   string         `yaml:"name"`
	Input  string         `yaml:"input"`
	Output map[string]any `yaml:"output"`
}

// QueryCase represents a single query test case
type QueryCase struct {
	Name   string         `yaml:"name"`
	Input  string         `yaml:"input"`
	Query  string         `yaml:"query"`
	Output map[string]any `yaml:"output"`
}

// RawAST represents the raw Tree-sitter AST structure
type RawAST struct {
	Type     string         `json:"type"`
	Text     string         `json:"text,omitempty"`
	Children []RawAST       `json:"children,omitempty"`
	Fields   map[string]any `json:"fields,omitempty"`
}

// LanguageTestSuite represents a discovered test suite
type LanguageTestSuite struct {
	Language string
	Suite    string
	Test     string
	Path     string // Path in embed.FS
}

// TestRunner handles running language-specific tests
type TestRunner struct {
	parser *Parser
}

// NewTestRunner creates a new test runner
func NewTestRunner() (*TestRunner, error) {
	parser, err := NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %v", err)
	}
	return &TestRunner{parser: parser}, nil
}

// LoadTestSuite loads a test suite from a YAML file
func (tr *TestRunner) LoadTestSuite(filepath string) (*TestSuite, error) {
	data, err := LanguageTestsFS.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test file: %v", err)
	}

	var suite TestSuite
	if err := yaml.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	return &suite, nil
}

// DumpRawAST dumps the raw AST for debugging purposes
func (tr *TestRunner) DumpRawAST(language, input string, filename string) error {
	// Get the language provider for this language
	provider, exists := tr.parser.providers[language]
	if !exists {
		return fmt.Errorf("no provider found for language: %s", language)
	}

	// Try to get raw tree-sitter AST if it's a tree-sitter provider
	if tsProvider, ok := provider.(*TreeSitterProvider); ok {
		return tr.DumpRawTreeSitterAST(tsProvider, input, filename)
	}

	// Fallback to UAST conversion for non-tree-sitter providers
	node, err := provider.Parse("test."+getLanguageFileExtension(language), []byte(input))
	if err != nil {
		return fmt.Errorf("failed to parse input: %v", err)
	}

	// Convert the UAST node back to raw AST structure
	rawAST := convertNodeToRawAST(node)

	data, err := json.MarshalIndent(rawAST, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal raw AST: %v", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// DumpRawTreeSitterAST dumps the actual tree-sitter AST before mapping
func (tr *TestRunner) DumpRawTreeSitterAST(provider *TreeSitterProvider, input string, filename string) error {
	parser := sitter.NewParser()
	parser.SetLanguage(provider.language)
	tree, err := parser.ParseString(context.Background(), nil, []byte(input))
	if err != nil {
		return fmt.Errorf("failed to parse with tree-sitter: %v", err)
	}

	root := tree.RootNode()
	if root.IsNull() {
		return fmt.Errorf("tree-sitter: no root node")
	}

	// Convert tree-sitter node to raw AST structure
	rawAST := convertTreeSitterNodeToRawAST(root, []byte(input))

	data, err := json.MarshalIndent(rawAST, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal raw AST: %v", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// convertNodeToRawAST converts a UAST node to raw AST structure
func convertNodeToRawAST(node *Node) RawAST {
	if node == nil {
		return RawAST{Type: "null"}
	}

	rawAST := RawAST{
		Type: node.Type,
		Text: node.Token,
	}

	if len(node.Props) > 0 {
		rawAST.Fields = make(map[string]interface{})
		for k, v := range node.Props {
			rawAST.Fields[k] = v
		}
	}

	if len(node.Children) > 0 {
		rawAST.Children = make([]RawAST, len(node.Children))
		for i, child := range node.Children {
			rawAST.Children[i] = convertNodeToRawAST(child)
		}
	}

	return rawAST
}

// convertTreeSitterNodeToRawAST converts a tree-sitter node to raw AST structure
func convertTreeSitterNodeToRawAST(node sitter.Node, source []byte) RawAST {
	rawAST := RawAST{
		Type: node.Type(),
		Text: node.String(),
	}

	// Add position information
	rawAST.Fields = map[string]interface{}{
		"start_line":   node.StartPoint().Row,
		"start_column": node.StartPoint().Column,
		"end_line":     node.EndPoint().Row,
		"end_column":   node.EndPoint().Column,
		"start_byte":   node.StartByte(),
		"end_byte":     node.EndByte(),
	}

	// Add children
	childCount := node.NamedChildCount()
	if childCount > 0 {
		rawAST.Children = make([]RawAST, childCount)
		for i := uint32(0); i < childCount; i++ {
			child := node.NamedChild(i)
			rawAST.Children[i] = convertTreeSitterNodeToRawAST(child, source)
		}
	}

	return rawAST
}

// RunParseTest runs a single parse test case
func (tr *TestRunner) RunParseTest(t *testing.T, language, testName string, tc ParseCase, suitePath string) {
	// Dump raw AST before parsing
	rawASTFile := filepath.Join(filepath.Dir(suitePath), fmt.Sprintf("%s-%s-%s.rawast.json", language, testName, tc.Name))
	if err := tr.DumpRawAST(language, tc.Input, rawASTFile); err != nil {
		t.Logf("Warning: failed to dump raw AST: %v", err)
	}

	// Parse the input
	node, err := tr.parser.Parse(fmt.Sprintf("test.%s", getLanguageFileExtension(language)), []byte(tc.Input))
	if err != nil {
		t.Errorf("Parse failed: %v", err)
		return
	}

	// Convert node to comparable format
	actual := nodeToMap(node)
	expected := tc.Output

	// Compare results
	if !compareUAST(actual, expected) {
		t.Errorf("Parse result mismatch for %s/%s", testName, tc.Name)
		t.Logf("Expected: %+v", expected)
		t.Logf("Actual: %+v", actual)
		t.Logf("Raw node: %+v", node)
	}
}

// RunQueryTest runs a single query test case
func (tr *TestRunner) RunQueryTest(t *testing.T, language, testName string, tc QueryCase, suitePath string) {
	// Dump raw AST before parsing
	rawASTFile := filepath.Join(filepath.Dir(suitePath), fmt.Sprintf("%s-%s-%s.rawast.json", language, testName, tc.Name))
	if err := tr.DumpRawAST(language, tc.Input, rawASTFile); err != nil {
		t.Logf("Warning: failed to dump raw AST: %v", err)
	}

	// Parse the input
	node, err := tr.parser.Parse(fmt.Sprintf("test.%s", getLanguageFileExtension(language)), []byte(tc.Input))
	if err != nil {
		t.Errorf("Parse failed: %v", err)
		return
	}

	// Check if node is nil before querying
	if node == nil {
		t.Errorf("Parsed node is nil for %s/%s", testName, tc.Name)
		return
	}

	// Execute query
	results, err := node.FindDSL(tc.Query)
	if err != nil {
		t.Errorf("Query failed: %v", err)
		return
	}

	// Convert results to comparable format
	actual := nodesToMap(results)
	expected := tc.Output

	// Compare results
	if !compareUAST(actual, expected) {
		t.Errorf("Query result mismatch for %s/%s", testName, tc.Name)
		t.Logf("Expected: %+v", expected)
		t.Logf("Actual: %+v", actual)
	}
}

// RunTestSuite runs all tests in a test suite
func (tr *TestRunner) RunTestSuite(t *testing.T, language, suitePath string) {
	suite, err := tr.LoadTestSuite(suitePath)
	if err != nil {
		t.Fatalf("Failed to load test suite: %v", err)
	}

	testName := strings.TrimSuffix(filepath.Base(suitePath), ".yaml")

	// Run parse tests
	for _, tc := range suite.ParseCases {
		t.Run(fmt.Sprintf("Parse_%s_%s", testName, tc.Name), func(t *testing.T) {
			tr.RunParseTest(t, language, testName, tc, suitePath)
		})
	}

	// Run query tests
	for _, tc := range suite.QueryCases {
		t.Run(fmt.Sprintf("Query_%s_%s", testName, tc.Name), func(t *testing.T) {
			tr.RunQueryTest(t, language, testName, tc, suitePath)
		})
	}
}

// DiscoverTestSuites finds all test suite files in the embedded language_tests directory
func DiscoverTestSuites() ([]LanguageTestSuite, error) {
	var suites []LanguageTestSuite

	err := fs.WalkDir(LanguageTestsFS, "language_tests", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		// path: language_tests/{language}/{suite}-{test}.yaml
		relPath := strings.TrimPrefix(path, "language_tests/")
		pathParts := strings.Split(relPath, string(filepath.Separator))
		if len(pathParts) != 2 {
			return fmt.Errorf("invalid test file format: %s", relPath)
		}
		language := pathParts[0]
		filename := pathParts[1]
		parts := strings.Split(strings.TrimSuffix(filename, ".yaml"), "-")
		if len(parts) < 2 {
			return fmt.Errorf("invalid test file format: %s", relPath)
		}
		suite := parts[0]
		test := strings.Join(parts[1:], "-")
		suites = append(suites, LanguageTestSuite{
			Language: language,
			Suite:    suite,
			Test:     test,
			Path:     path,
		})
		return nil
	})
	return suites, err
}

// Helper functions

func getLanguageFileExtension(language string) string {
	extensions := map[string]string{
		"go":         "go",
		"python":     "py",
		"javascript": "js",
		"typescript": "ts",
		"java":       "java",
		"cpp":        "cpp",
		"c":          "c",
		"rust":       "rs",
		"ruby":       "rb",
		"php":        "php",
		"csharp":     "cs",
		"kotlin":     "kt",
		"swift":      "swift",
		"scala":      "scala",
		"dart":       "dart",
		"elixir":     "ex",
		"clojure":    "clj",
		"haskell":    "hs",
		"ocaml":      "ml",
		"fsharp":     "fs",
		"erlang":     "erl",
		"lua":        "lua",
		"perl":       "pl",
		"bash":       "sh",
		"powershell": "ps1",
		"yaml":       "yaml",
		"json":       "json",
		"xml":        "xml",
		"html":       "html",
		"css":        "css",
		"sql":        "sql",
		"markdown":   "md",
		"toml":       "toml",
		"ini":        "ini",
		"dockerfile": "Dockerfile",
		"makefile":   "Makefile",
	}

	if ext, ok := extensions[language]; ok {
		return ext
	}
	return "txt"
}

func nodeToMap(node *Node) map[string]interface{} {
	if node == nil {
		return nil
	}

	result := map[string]interface{}{
		"type": node.Type,
	}

	if node.Token != "" {
		result["token"] = node.Token
	}

	if len(node.Props) > 0 {
		result["props"] = node.Props
	}

	if len(node.Roles) > 0 {
		roles := make([]string, len(node.Roles))
		for i, role := range node.Roles {
			roles[i] = string(role)
		}
		result["roles"] = roles
	}

	if len(node.Children) > 0 {
		children := make([]map[string]interface{}, len(node.Children))
		for i, child := range node.Children {
			children[i] = nodeToMap(child)
		}
		result["children"] = children
	}

	return result
}

func nodesToMap(nodes []*Node) map[string]interface{} {
	if len(nodes) == 0 {
		return map[string]interface{}{"results": []interface{}{}}
	}

	// Check if all nodes are literal nodes (type "Literal")
	allLiterals := true
	for _, node := range nodes {
		if node.Type != "Literal" {
			allLiterals = false
			break
		}
	}

	// If all nodes are literals, return just the token values
	if allLiterals {
		results := make([]interface{}, len(nodes))
		for i, node := range nodes {
			results[i] = node.Token
		}
		return map[string]interface{}{"results": results}
	}

	// Otherwise, convert to node maps as before
	if len(nodes) == 1 {
		return map[string]interface{}{"results": []interface{}{nodeToMap(nodes[0])}}
	}

	results := make([]interface{}, len(nodes))
	for i, node := range nodes {
		results[i] = nodeToMap(node)
	}

	return map[string]interface{}{"results": results}
}

func compareUAST(actual, expected map[string]interface{}) bool {
	// Simple comparison for now - can be enhanced with more sophisticated matching
	actualJSON, _ := json.Marshal(actual)
	expectedJSON, _ := json.Marshal(expected)
	return string(actualJSON) == string(expectedJSON)
}

// Test functions

func TestAllLanguageTests(t *testing.T) {
	suites, err := DiscoverTestSuites()
	if err != nil {
		t.Fatalf("Failed to discover test suites: %v", err)
	}
	if len(suites) == 0 {
		t.Skip("No test suites found")
	}
	runner, err := NewTestRunner()
	if err != nil {
		t.Fatalf("Failed to create test runner: %v", err)
	}
	for _, suite := range suites {
		t.Run(fmt.Sprintf("%s_%s_%s", suite.Language, suite.Suite, suite.Test), func(t *testing.T) {
			runner.RunTestSuite(t, suite.Language, suite.Path)
		})
	}
}
