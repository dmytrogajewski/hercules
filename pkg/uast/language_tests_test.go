package uast

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"context"

	sitter "github.com/alexaandru/go-tree-sitter-bare"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/spec"
	"github.com/xeipuuv/gojsonschema"
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
	// Get the language provider for this language using the loader
	ext := "." + getLanguageFileExtension(language)
	provider, exists := tr.parser.loader.LanguageParser(ext)
	if !exists {
		return fmt.Errorf("no provider found for language: %s", language)
	}

	// Try to get raw tree-sitter AST if it's a DSL provider
	if dslProvider, ok := provider.(*DSLParser); ok {
		return tr.DumpRawTreeSitterAST(dslProvider, input, filename)
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
func (tr *TestRunner) DumpRawTreeSitterAST(provider *DSLParser, input string, filename string) error {
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
func convertNodeToRawAST(n *node.Node) RawAST {
	if n == nil {
		return RawAST{Type: "null"}
	}

	rawAST := RawAST{
		Type: string(n.Type),
		Text: n.Token,
	}

	if len(n.Props) > 0 {
		rawAST.Fields = make(map[string]interface{})
		for k, v := range n.Props {
			rawAST.Fields[k] = v
		}
	}

	if len(n.Children) > 0 {
		rawAST.Children = make([]RawAST, len(n.Children))
		for i, child := range n.Children {
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

// ValidateUAST validates a UAST node against the schema
func (tr *TestRunner) ValidateUAST(node *node.Node, testName string) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}

	// Convert node to map for validation
	nodeMap := nodeToMap(node)

	// Load schema from embed.FS
	schemaBytes, err := spec.UASTSchemaFS.ReadFile("uast-schema.json")
	if err != nil {
		return fmt.Errorf("failed to read embedded schema: %v", err)
	}
	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	inputLoader := gojsonschema.NewGoLoader(nodeMap)

	result, err := gojsonschema.Validate(schemaLoader, inputLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %v", err)
	}

	if result.Valid() {
		return nil
	}

	// Calculate compliance
	compliance := calculateCompliance(nodeMap, result.Errors())

	// Build Rust-style error message
	var errorMsgs []string
	errorMsgs = append(errorMsgs, fmt.Sprintf("❌ UAST validation failed for %s (Compliance: %d%%)", testName, compliance))
	errorMsgs = append(errorMsgs, "")
	errorMsgs = append(errorMsgs, "📋 Validation Errors:")
	errorMsgs = append(errorMsgs, "")

	// Group errors by type for better organization
	roleErrors := make([]gojsonschema.ResultError, 0)
	typeErrors := make([]gojsonschema.ResultError, 0)
	otherErrors := make([]gojsonschema.ResultError, 0)

	for _, verr := range result.Errors() {
		if strings.Contains(verr.Description(), "must be one of the following") {
			roleErrors = append(roleErrors, verr)
		} else if strings.Contains(verr.Description(), "type") {
			typeErrors = append(typeErrors, verr)
		} else {
			otherErrors = append(otherErrors, verr)
		}
	}

	// Report role errors
	if len(roleErrors) > 0 {
		errorMsgs = append(errorMsgs, "🔑 Role Validation Errors:")
		for _, verr := range roleErrors {
			actualValue := getActualValue(nodeMap, verr.Field())
			errorPath := strings.Split(verr.Field(), ".")
			treeWithHighlight := printUASTWithHighlight(nodeMap, errorPath)
			suggestion := getRoleSuggestion(verr.Description(), actualValue)
			nodeType := getNodeTypeAtPath(nodeMap, verr.Field())

			errorMsgs = append(errorMsgs, "   • UAST Tree Structure (full):")
			errorMsgs = append(errorMsgs, treeWithHighlight)
			if nodeType != "" {
				errorMsgs = append(errorMsgs, fmt.Sprintf("     Node Type: %s", nodeType))
			}
			errorMsgs = append(errorMsgs, fmt.Sprintf("     Expected: %s", extractExpectedRoles(verr.Description())))
			if actualValue != "" {
				errorMsgs = append(errorMsgs, fmt.Sprintf("     Got: %q", actualValue))
			}
			if suggestion != "" {
				errorMsgs = append(errorMsgs, fmt.Sprintf("     💡 %s", suggestion))
			}
			errorMsgs = append(errorMsgs, "")
		}
	}

	// Report type errors
	if len(typeErrors) > 0 {
		errorMsgs = append(errorMsgs, "🏷️  Type Validation Errors:")
		for _, verr := range typeErrors {
			actualValue := getActualValue(nodeMap, verr.Field())
			fieldPath := formatFieldPath(verr.Field(), nodeMap)

			errorMsgs = append(errorMsgs, fmt.Sprintf("   • %s", fieldPath))
			errorMsgs = append(errorMsgs, fmt.Sprintf("     %s", verr.Description()))
			if actualValue != "" {
				errorMsgs = append(errorMsgs, fmt.Sprintf("     Got: %q", actualValue))
			}
			errorMsgs = append(errorMsgs, "")
		}
	}

	// Report other errors
	if len(otherErrors) > 0 {
		errorMsgs = append(errorMsgs, "⚠️  Other Validation Errors:")
		for _, verr := range otherErrors {
			actualValue := getActualValue(nodeMap, verr.Field())
			fieldPath := formatFieldPath(verr.Field(), nodeMap)

			errorMsgs = append(errorMsgs, fmt.Sprintf("   • %s", fieldPath))
			errorMsgs = append(errorMsgs, fmt.Sprintf("     %s", verr.Description()))
			if actualValue != "" {
				errorMsgs = append(errorMsgs, fmt.Sprintf("     Got: %q", actualValue))
			}
			errorMsgs = append(errorMsgs, "")
		}
	}

	// Add helpful suggestions
	errorMsgs = append(errorMsgs, "💡 Suggestions:")
	errorMsgs = append(errorMsgs, "   • Check that all nodes have valid roles from the UAST schema")
	errorMsgs = append(errorMsgs, "   • Ensure Block nodes have roles: \"Body\" where appropriate")
	errorMsgs = append(errorMsgs, "   • Verify parameter grouping matches provider output structure")
	errorMsgs = append(errorMsgs, "   • Update test YAML to match actual provider output exactly")
	errorMsgs = append(errorMsgs, "")

	return fmt.Errorf("%s", strings.Join(errorMsgs, "\n"))
}

// calculateCompliance calculates compliance percentage
func calculateCompliance(data interface{}, errors []gojsonschema.ResultError) int {
	totalNodes := countNodes(data)
	if totalNodes == 0 {
		return 0
	}

	validNodes := totalNodes - len(errors)
	compliance := int(float64(validNodes) / float64(totalNodes) * 100)

	if compliance < 0 {
		compliance = 0
	} else if compliance > 100 {
		compliance = 100
	}

	return compliance
}

// countNodes counts total nodes in UAST
func countNodes(data interface{}) int {
	count := 1

	switch v := data.(type) {
	case map[string]interface{}:
		if children, ok := v["children"].([]interface{}); ok {
			for _, child := range children {
				count += countNodes(child)
			}
		}
	case []interface{}:
		for _, item := range v {
			count += countNodes(item)
		}
	}

	return count
}

// getActualValue gets the actual value at a field path
func getActualValue(data interface{}, fieldPath string) string {
	parts := strings.Split(fieldPath, ".")

	current := data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, ok := v[part]; ok {
				current = val
			} else {
				return ""
			}
		case []interface{}:
			if idx, err := strconv.Atoi(part); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return ""
			}
		default:
			return ""
		}
	}

	switch v := current.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// getCwd returns current working directory
func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
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

	// Validate UAST against schema
	if err := tr.ValidateUAST(node, fmt.Sprintf("%s/%s", testName, tc.Name)); err != nil {
		t.Errorf("UAST validation failed: %v", err)
		// Continue with comparison even if validation fails
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
		// Dump actual as YAML for debugging
		actualYAML, err := yaml.Marshal(actual)
		if err == nil {
			t.Logf("Actual YAML:\n%s", string(actualYAML))
		}
		t.FailNow()
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

	// Validate UAST against schema
	if err := tr.ValidateUAST(node, fmt.Sprintf("%s/%s", testName, tc.Name)); err != nil {
		t.Errorf("UAST validation failed: %v", err)
		// Continue with query even if validation fails
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

func nodeToMap(n *node.Node) map[string]interface{} {
	if n == nil {
		return nil
	}

	result := map[string]interface{}{
		"type": n.Type,
	}

	if n.Token != "" {
		result["token"] = n.Token
	}

	if len(n.Props) > 0 {
		result["props"] = n.Props
	}

	// Always include roles, even if empty
	roles := make([]string, len(n.Roles))
	for i, role := range n.Roles {
		roles[i] = string(role)
	}
	result["roles"] = roles

	if len(n.Children) > 0 {
		children := make([]map[string]interface{}, len(n.Children))
		for i, child := range n.Children {
			children[i] = nodeToMap(child)
		}
		result["children"] = children
	}

	return result
}

func nodesToMap(nodes []*node.Node) map[string]interface{} {
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

// formatFieldPath creates a Rust-style error format showing the actual UAST tree with highlighting
func formatFieldPath(fieldPath string, nodeMap map[string]interface{}) string {
	parts := strings.Split(fieldPath, ".")

	// Build the complete tree structure
	treeLines := buildTreeStructure(nodeMap, 0, "")

	// Find and highlight the error location
	highlightedLines := highlightErrorLocation(treeLines, parts)

	return strings.Join(highlightedLines, "\n")
}

// buildTreeStructure recursively builds a visual representation of the UAST tree
func buildTreeStructure(data interface{}, depth int, prefix string) []string {
	var lines []string

	switch v := data.(type) {
	case map[string]interface{}:
		// Add node type and properties
		nodeType := "Unknown"
		if t, ok := v["type"].(string); ok {
			nodeType = t
		}

		// Add roles if present
		roles := ""
		if r, ok := v["roles"].([]interface{}); ok && len(r) > 0 {
			roleStrs := make([]string, len(r))
			for i, role := range r {
				roleStrs[i] = fmt.Sprintf("%v", role)
			}
			roles = fmt.Sprintf(" [%s]", strings.Join(roleStrs, ", "))
		}

		// Add token if present
		token := ""
		if t, ok := v["token"].(string); ok && t != "" {
			token = fmt.Sprintf(" '%s'", t)
		}

		// Add props if present
		props := ""
		if p, ok := v["props"].(map[string]interface{}); ok && len(p) > 0 {
			propStrs := make([]string, 0, len(p))
			for k, val := range p {
				propStrs = append(propStrs, fmt.Sprintf("%s:%v", k, val))
			}
			props = fmt.Sprintf(" {%s}", strings.Join(propStrs, ", "))
		}

		line := fmt.Sprintf("%s%s%s%s%s", prefix, nodeType, roles, token, props)
		lines = append(lines, line)

		// Add children
		if children, ok := v["children"].([]interface{}); ok && len(children) > 0 {
			for i, child := range children {
				isLast := i == len(children)-1
				childPrefix := prefix
				if depth > 0 {
					if isLast {
						childPrefix += "   └─ "
					} else {
						childPrefix += "   ├─ "
					}
				}
				childLines := buildTreeStructure(child, depth+1, childPrefix)
				lines = append(lines, childLines...)
			}
		}

	case []interface{}:
		for i, item := range v {
			isLast := i == len(v)-1
			itemPrefix := prefix
			if depth > 0 {
				if isLast {
					itemPrefix += "   └─ "
				} else {
					itemPrefix += "   ├─ "
				}
			}
			itemLines := buildTreeStructure(item, depth+1, itemPrefix)
			lines = append(lines, itemLines...)
		}
	}

	return lines
}

// highlightErrorLocation highlights the problematic part of the tree
func highlightErrorLocation(treeLines []string, errorPath []string) []string {
	if len(errorPath) < 2 {
		return treeLines
	}

	// Find the line that corresponds to the error
	for i, line := range treeLines {
		// Check if this line represents the error location
		if isErrorLocation(line, errorPath) {
			// Highlight the problematic part
			treeLines[i] = fmt.Sprintf(">>> %s <<< ERROR HERE", strings.TrimSpace(line))
			break
		}
	}

	return treeLines
}

// isErrorLocation checks if a line represents the error location
func isErrorLocation(line string, errorPath []string) bool {
	// Look for the last part of the error path (the actual error field)
	if len(errorPath) == 0 {
		return false
	}

	lastPart := errorPath[len(errorPath)-1]
	trimmedLine := strings.TrimSpace(line)

	// Check if this line contains the error field
	if strings.Contains(trimmedLine, lastPart) {
		return true
	}

	// Check for array indices
	if idx, err := strconv.Atoi(lastPart); err == nil {
		// Look for array-like patterns in the line
		if strings.Contains(trimmedLine, fmt.Sprintf("[%d]", idx)) {
			return true
		}
	}

	return false
}

// getNodeTypeAtPath gets the node type at a specific field path
func getNodeTypeAtPath(data interface{}, fieldPath string) string {
	// Remove the last part (which is usually "roles" or "type") to get to the node
	parts := strings.Split(fieldPath, ".")
	if len(parts) < 2 {
		return ""
	}

	// Reconstruct path to the node (without the last part)
	nodePath := strings.Join(parts[:len(parts)-1], ".")

	current := data
	pathParts := strings.Split(nodePath, ".")
	for _, part := range pathParts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, ok := v[part]; ok {
				current = val
			} else {
				return ""
			}
		case []interface{}:
			if idx, err := strconv.Atoi(part); err == nil && idx >= 0 && idx < len(v) {
				current = v[idx]
			} else {
				return ""
			}
		default:
			return ""
		}
	}

	// Try to get the type from the node
	switch v := current.(type) {
	case map[string]interface{}:
		if nodeType, ok := v["type"].(string); ok {
			return nodeType
		}
	}

	return ""
}

// Improved pretty-printer for UAST tree with better tree structure
func printUASTWithHighlight(data interface{}, errorPath []string) string {
	var lines []string
	// If the error is on a role (ends with 'roles' or 'roles.N'), highlight the parent node
	highlightPath := errorPath
	if len(errorPath) >= 2 && (errorPath[len(errorPath)-2] == "roles" || errorPath[len(errorPath)-1] == "roles") {
		highlightPath = errorPath[:len(errorPath)-2]
	}
	printUASTTreeRec(data, highlightPath, []string{}, &lines, "", true)
	return strings.Join(lines, "\n")
}

// printUASTTreeRec prints the tree with improved structure
func printUASTTreeRec(data interface{}, errorPath []string, curPath []string, lines *[]string, prefix string, isLast bool) {
	highlight := len(curPath) == len(errorPath)
	if highlight {
		for i := range curPath {
			if curPath[i] != errorPath[i] {
				highlight = false
				break
			}
		}
	}

	switch v := data.(type) {
	case map[string]interface{}:
		nodeType := "Unknown"
		if t, ok := v["type"].(string); ok {
			nodeType = t
		}
		roles := []string{}
		if r, ok := v["roles"].([]string); ok && len(r) > 0 {
			roles = r
		} else if r, ok := v["roles"].([]interface{}); ok && len(r) > 0 {
			for _, role := range r {
				roles = append(roles, fmt.Sprintf("%v", role))
			}
		}
		rolesStr := ""
		if len(roles) > 0 {
			rolesStr = fmt.Sprintf(" [%s]", strings.Join(roles, ", "))
		} else {
			rolesStr = " []"
		}
		token := ""
		if t, ok := v["token"].(string); ok && t != "" {
			token = fmt.Sprintf(" '%s'", t)
		}
		props := ""
		if p, ok := v["props"].(map[string]interface{}); ok && len(p) > 0 {
			propStrs := make([]string, 0, len(p))
			for k, val := range p {
				propStrs = append(propStrs, fmt.Sprintf("%s:%v", k, val))
			}
			props = fmt.Sprintf(" {%s}", strings.Join(propStrs, ", "))
		}
		branch := "├─"
		if isLast {
			branch = "└─"
		}
		line := prefix + branch + " " + nodeType + rolesStr + token + props
		if highlight {
			line += "   <--- ERROR HERE"
		}
		*lines = append(*lines, line)
		// Children
		childrenCount := 0
		if childrenRaw, ok := v["children"]; ok {
			switch children := childrenRaw.(type) {
			case []interface{}:
				childrenCount = len(children)
			case []map[string]interface{}:
				childrenCount = len(children)
			}
		}
		childIdx := 0
		if childrenRaw, ok := v["children"]; ok {
			switch children := childrenRaw.(type) {
			case []interface{}:
				for i, child := range children {
					isChildLast := i == childrenCount-1
					nextPrefix := prefix
					if isLast {
						nextPrefix += "   "
					} else {
						nextPrefix += "│  "
					}
					printUASTTreeRec(child, errorPath, append(curPath, "children", fmt.Sprintf("%d", childIdx)), lines, nextPrefix, isChildLast)
					childIdx++
				}
			case []map[string]interface{}:
				for i, child := range children {
					isChildLast := i == childrenCount-1
					nextPrefix := prefix
					if isLast {
						nextPrefix += "   "
					} else {
						nextPrefix += "│  "
					}
					printUASTTreeRec(child, errorPath, append(curPath, "children", fmt.Sprintf("%d", childIdx)), lines, nextPrefix, isChildLast)
					childIdx++
				}
			}
		}
	case []interface{}:
		for i, item := range v {
			isLastItem := i == len(v)-1
			printUASTTreeRec(item, errorPath, append(curPath, fmt.Sprintf("%d", i)), lines, prefix, isLastItem)
		}
	case string, float64, int, bool:
		line := prefix + fmt.Sprintf("%v", v)
		if highlight {
			line += "   <--- ERROR HERE"
		}
		*lines = append(*lines, line)
	}
}

// getRoleSuggestion provides helpful suggestions for role validation errors
func getRoleSuggestion(description, actualValue string) string {
	if strings.Contains(description, "roles") {
		if actualValue == "" || actualValue == "[]" {
			return "Node is missing roles. Consider adding appropriate roles like 'Identifier', 'Function', 'Declaration', etc."
		}
		return "Check that the roles match the expected UAST schema roles for this node type."
	}
	return ""
}

// extractExpectedRoles extracts expected roles from validation error description
func extractExpectedRoles(description string) string {
	// Look for patterns like "expected one of [Role1, Role2, Role3]"
	if strings.Contains(description, "expected one of") {
		start := strings.Index(description, "[")
		end := strings.Index(description, "]")
		if start != -1 && end != -1 && end > start {
			return description[start+1 : end]
		}
	}

	// Look for patterns like "must be one of: Role1, Role2, Role3"
	if strings.Contains(description, "must be one of:") {
		parts := strings.Split(description, "must be one of:")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}

	return "valid UAST roles"
}
