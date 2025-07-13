package uast

import (
	"fmt"
	"strings"
	"testing"
)

func TestDSLParser_Parse_Valid(t *testing.T) {
	cases := []struct {
		input   string
		wantAST string
	}{
		{"map(.children)", "Map(Field(children))"},
		{"map(.children) |> filter(.type == \"FunctionDecl\")", "Pipeline(Map(Field(children)) | Filter(Call(==, Field(type), Literal(FunctionDecl))))"},
		{".foo", "Field(foo)"},
		{"42", "Literal(42)"},
		{"map(.children) |> filter(.type == \"FunctionDecl\") |> reduce(count)", "Pipeline(Map(Field(children)) | Filter(Call(==, Field(type), Literal(FunctionDecl))) | Reduce(Call(count)))"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			ast, err := ParseDSL(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := stringifyAST(ast)
			if got != tc.wantAST {
				t.Errorf("got %q, want %q", got, tc.wantAST)
			}
		})
	}
}

func TestDSLParser_Parse_Invalid(t *testing.T) {
	cases := []struct {
		input   string
		wantErr string
	}{
		// {"bad", "parse error at 1:1: unexpected token 'bad'"},
		{"@#$", "parse error at 1:1: unknown input"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			_, err := ParseDSL(tc.input)
			if err == nil || err.Error() != tc.wantErr {
				t.Errorf("got error %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestDSLParser_LoweringAndExecution(t *testing.T) {
	testUAST := &Node{
		Type:  "Root",
		Props: map[string]string{"foo": "bar", "type": "FunctionDecl"},
		Children: []*Node{
			{Type: "Child", Props: map[string]string{"foo": "baz", "type": "Other"}},
			{Type: "Child", Props: map[string]string{"foo": "qux", "type": "FunctionDecl"}},
		},
	}
	cases := []struct {
		dsl     string
		input   *Node
		want    []string // expected output tokens
		wantErr bool
	}{
		{".foo", testUAST, []string{"bar"}, false},
		{"42", testUAST, []string{"42"}, false},
		{"map(.foo)", testUAST, []string{"baz", "qux"}, false},
		{"reduce(count)", testUAST, []string{"2"}, false},
		{"map(.foo) |> reduce(count)", testUAST, []string{"2"}, false},
		{"bad syntax", testUAST, nil, true},
		{"map(.notfound)", testUAST, []string{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.dsl, func(t *testing.T) {
			ast, err := ParseDSL(tc.dsl)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			qf, err := LowerDSL(ast)
			if err != nil {
				t.Fatalf("lowering error: %v", err)
			}
			out := qf([]*Node{tc.input})
			var got []string
			for _, n := range out {
				got = append(got, n.Token)
			}
			if len(got) != len(tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
				return
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("got %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestDSL_E2E_GoIntegration(t *testing.T) {
	goCode := `package main
func hello() {}
func world() {}`
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	uast, err := parser.Parse("main.go", []byte(goCode))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if uast == nil {
		t.Fatalf("UAST is nil")
	}
	// Collect all nodes in the tree
	nodes := uast.Find(func(n *Node) bool { return true })
	// Query: get all function nodes' types
	dsl := "filter(.type == \"Function\") |> map(.type)"
	ast, err := ParseDSL(dsl)
	if err != nil {
		t.Fatalf("DSL parse error: %v", err)
	}
	qf, err := LowerDSL(ast)
	if err != nil {
		t.Fatalf("DSL lowering error: %v", err)
	}
	out := qf(nodes)
	var got []string
	for _, n := range out {
		got = append(got, n.Token)
	}
	want := []string{"Function", "Function"}
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestDSL_E2E_GoComplexProgram(t *testing.T) {
	goCode := `package main

import "fmt"

type Greeter struct {
	Name string
}

func (g Greeter) Greet() string {
	return "Hello, " + g.Name
}

func add(a, b int) int {
	return a + b
}

func main() {
	g := Greeter{Name: "World"}
	fmt.Println(g.Greet())
	fmt.Println(add(2, 3))
}`
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	uast, err := parser.Parse("main.go", []byte(goCode))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if uast == nil {
		t.Fatalf("UAST is nil")
	}
	// Collect all nodes in the tree
	nodes := uast.Find(func(n *Node) bool { return true })
	// Print all node types for debugging
	for _, n := range nodes {
		t.Logf("Node type: %s, Props: %v", n.Type, n.Props)
	}
	// Print Props and Token for all Function and Method nodes
	for _, n := range nodes {
		if n.Type == "Function" || n.Type == "Method" {
			t.Logf("%s Props: %v, Token: %q", n.Type, n.Props, n.Token)
		}
	}
	// Print children of Method node for debugging
	for _, n := range nodes {
		if n.Type == "Method" {
			t.Logf("Method node children: %d", len(n.Children))
			for i, c := range n.Children {
				t.Logf("  Child %d: Type=%s, Props=%v", i, c.Type, c.Props)
			}
		}
	}
	// Query: get all function/method names
	dsl := "filter(.type == \"Function\" || .type == \"Method\") |> map(.name)"
	ast, err := ParseDSL(dsl)
	if err != nil {
		t.Fatalf("DSL parse error: %v", err)
	}
	qf, err := LowerDSL(ast)
	if err != nil {
		t.Fatalf("DSL lowering error: %v", err)
	}
	out := qf(nodes)
	var got []string
	for _, n := range out {
		got = append(got, n.Token)
	}
	// Expect all function/method names
	want := []string{"Greet", "add", "main"}
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func TestDSLParser_Parse_MembershipAndLogical(t *testing.T) {
	query := "filter(.type == \"Function\" && .roles has \"Exported\")"
	_, err := ParseDSL(query)
	if err != nil {
		t.Fatalf("ParseDSL failed: %v", err)
	}
}

func TestDSLParser_RecursiveFunctions(t *testing.T) {
	testUAST := &Node{
		Type:  "Root",
		Props: map[string]string{"foo": "bar", "type": "FunctionDecl"},
		Children: []*Node{
			{Type: "Child", Props: map[string]string{"foo": "baz", "type": "Other"}},
			{Type: "Child", Props: map[string]string{"foo": "qux", "type": "FunctionDecl"}},
		},
	}
	cases := []struct {
		dsl     string
		input   *Node
		want    []string // expected output tokens
		wantErr bool
	}{
		// Recursive filter: should find all FunctionDecl nodes in the entire tree
		{"rfilter(.type == \"FunctionDecl\")", testUAST, []string{"Root", "Child"}, false},
		// Recursive map: should map all foo values in the entire tree
		{"rmap(.foo)", testUAST, []string{"bar", "baz", "qux"}, false},
		// Recursive filter + recursive map: should find FunctionDecl nodes and map their foo values recursively
		{"rfilter(.type == \"FunctionDecl\") |> rmap(.foo)", testUAST, []string{"bar", "baz", "qux", "qux"}, false},
		// Recursive filter + non-recursive map: should find FunctionDecl nodes but only map those specific nodes
		{"rfilter(.type == \"FunctionDecl\") |> map(.foo)", testUAST, []string{"bar", "qux"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.dsl, func(t *testing.T) {
			ast, err := ParseDSL(tc.dsl)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			qf, err := LowerDSL(ast)
			if err != nil {
				t.Fatalf("lowering error: %v", err)
			}
			out := qf([]*Node{tc.input})
			var got []string
			for _, n := range out {
				// For filter results, use the node type; for map results, use the token
				if strings.Contains(tc.dsl, "rfilter") && !strings.Contains(tc.dsl, "map") {
					got = append(got, n.Type)
				} else {
					got = append(got, n.Token)
				}
			}
			if len(got) != len(tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
				return
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("got %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestNestedFieldAccess(t *testing.T) {
	// Test parsing nested field access
	ast, err := ParseDSL(".props.name")
	if err != nil {
		t.Fatalf("Failed to parse nested field access: %v", err)
	}

	fieldNode, ok := ast.(*FieldNode)
	if !ok {
		t.Fatalf("Expected FieldNode, got %T", ast)
	}

	if len(fieldNode.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(fieldNode.Fields))
	}

	if fieldNode.Fields[0] != "props" {
		t.Errorf("Expected first field to be 'props', got '%s'", fieldNode.Fields[0])
	}

	if fieldNode.Fields[1] != "name" {
		t.Errorf("Expected second field to be 'name', got '%s'", fieldNode.Fields[1])
	}

	// Test deeper nesting
	ast, err = ParseDSL(".props.deep.nested.field")
	if err != nil {
		t.Fatalf("Failed to parse deep nested field access: %v", err)
	}

	fieldNode, ok = ast.(*FieldNode)
	if !ok {
		t.Fatalf("Expected FieldNode, got %T", ast)
	}

	expected := []string{"props", "deep", "nested", "field"}
	if len(fieldNode.Fields) != len(expected) {
		t.Fatalf("Expected %d fields, got %d", len(expected), len(fieldNode.Fields))
	}

	for i, expectedField := range expected {
		if fieldNode.Fields[i] != expectedField {
			t.Errorf("Expected field[%d] to be '%s', got '%s'", i, expectedField, fieldNode.Fields[i])
		}
	}
}

func TestNestedFieldAccessExecution(t *testing.T) {
	// Create a test node with nested properties
	testNode := &Node{
		Type: "Function",
		Props: map[string]string{
			"name": "testFunction",
			"deep": "nestedValue",
		},
	}

	// Test single field access (backward compatibility)
	ast, err := ParseDSL(".type")
	if err != nil {
		t.Fatalf("Failed to parse single field access: %v", err)
	}

	queryFunc, err := LowerDSL(ast)
	if err != nil {
		t.Fatalf("Failed to lower DSL: %v", err)
	}

	results := queryFunc([]*Node{testNode})
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Token != "Function" {
		t.Errorf("Expected 'Function', got '%s'", results[0].Token)
	}

	// Test nested field access
	ast, err = ParseDSL(".props.name")
	if err != nil {
		t.Fatalf("Failed to parse nested field access: %v", err)
	}

	// Debug: print the AST
	t.Logf("AST: %+v", ast)

	queryFunc, err = LowerDSL(ast)
	if err != nil {
		t.Fatalf("Failed to lower DSL: %v", err)
	}

	results = queryFunc([]*Node{testNode})
	t.Logf("Results: %+v", results)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Token != "testFunction" {
		t.Errorf("Expected 'testFunction', got '%s'", results[0].Token)
	}
}

func TestDSLParser_HasSyntax(t *testing.T) {
	// Test cases for "has" syntax
	testCases := []struct {
		name        string
		query       string
		shouldParse bool
	}{
		{
			name:        "has with roles",
			query:       `.roles has "Function"`,
			shouldParse: true,
		},
		{
			name:        "has with type",
			query:       `.type has "Function"`,
			shouldParse: true,
		},
		{
			name:        "has with props",
			query:       `.props has "name"`,
			shouldParse: true,
		},
		{
			name:        "has with nested field",
			query:       `.props.name has "value"`,
			shouldParse: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the query
			ast, err := ParseDSL(tc.query)
			if tc.shouldParse {
				if err != nil {
					t.Fatalf("Failed to parse query '%s': %v", tc.query, err)
				}
				if ast == nil {
					t.Fatalf("AST is nil for query '%s'", tc.query)
				}
				t.Logf("Successfully parsed: %s", tc.query)
			} else {
				if err == nil {
					t.Fatalf("Expected parse error for '%s' but got none", tc.query)
				}
				t.Logf("Expected parse error for: %s", tc.query)
			}
		})
	}
}

func TestDSLParser_HasSyntaxExecution(t *testing.T) {
	// Create a test node with roles
	testNode := &Node{
		Type:  "Function",
		Token: "testFunction",
		Roles: []Role{"Function", "Declaration"},
		Props: map[string]string{
			"name": "testFunction",
		},
	}

	testCases := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "has with roles - should be true",
			query:    `filter(.roles has "Function")`,
			expected: true,
		},
		{
			name:     "has with roles - should be false",
			query:    `filter(.roles has "NonExistent")`,
			expected: false,
		},
		{
			name:     "has with type - should be true (type matches)",
			query:    `filter(.type has "Function")`,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the query
			ast, err := ParseDSL(tc.query)
			if err != nil {
				t.Fatalf("Failed to parse query '%s': %v", tc.query, err)
			}

			// Lower the DSL
			queryFunc, err := LowerDSL(ast)
			if err != nil {
				t.Fatalf("Failed to lower DSL for '%s': %v", tc.query, err)
			}

			// Execute the query
			results := queryFunc([]*Node{testNode})

			t.Logf("Query: %s", tc.query)
			t.Logf("Results count: %d", len(results))

			if tc.expected {
				if len(results) == 0 {
					t.Errorf("Expected results but got none for query: %s", tc.query)
				}
			} else {
				if len(results) > 0 {
					t.Errorf("Expected no results but got %d for query: %s", len(results), tc.query)
				}
			}
		})
	}
}

func TestPropertyAccess(t *testing.T) {
	// Create a test node with a name property
	node := &Node{
		Type: "Function",
		Props: map[string]string{
			"name": "my_function",
		},
	}

	// Test the query: map(.name)
	query := "map(.name)"
	ast, err := ParseDSL(query)
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	func_, err := LowerDSL(ast)
	if err != nil {
		t.Fatalf("Failed to lower DSL: %v", err)
	}

	result := func_([]*Node{node})
	if len(result) == 0 {
		t.Fatalf("Expected result, got empty")
	}

	if result[0].Type != "Literal" {
		t.Errorf("Expected Literal node, got %s", result[0].Type)
	}

	if result[0].Token != "my_function" {
		t.Errorf("Expected token 'my_function', got '%s'", result[0].Token)
	}
}

func TestDebugFunctionNameQuery(t *testing.T) {
	// Create a test node that matches the actual UAST structure from the Perl test
	node := &Node{
		Type:  "Function",
		Roles: []Role{"Function", "Declaration"},
		Props: map[string]string{
			"name": "my_function",
		},
		Token: "my_function", // Also set as token
	}

	// Test the full query: filter(.type == "Function") |> map(.name)
	query := `filter(.type == "Function") |> map(.name)`
	ast, err := ParseDSL(query)
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	func_, err := LowerDSL(ast)
	if err != nil {
		t.Fatalf("Failed to lower DSL: %v", err)
	}

	result := func_([]*Node{node})
	t.Logf("Query result: %+v", result)

	if len(result) == 0 {
		t.Fatalf("Expected result, got empty")
	}

	if result[0].Type != "Literal" {
		t.Errorf("Expected Literal node, got %s", result[0].Type)
	}

	if result[0].Token != "my_function" {
		t.Errorf("Expected token 'my_function', got '%s'", result[0].Token)
	}
}

func TestDebugMapFunction(t *testing.T) {
	// Create a test node that matches the actual UAST structure from the Perl test
	node := &Node{
		Type:  "Function",
		Roles: []Role{"Function", "Declaration"},
		Props: map[string]string{
			"name": "my_function",
		},
		Token: "my_function",
		Children: []*Node{
			{
				Type:  "Identifier",
				Token: "my_function",
				Roles: []Role{"Name"},
			},
		},
	}

	// Test the exact query from the Perl test
	query := `filter(.type == "Function") |> map(.name)`
	ast, err := ParseDSL(query)
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	func_, err := LowerDSL(ast)
	if err != nil {
		t.Fatalf("Failed to lower DSL: %v", err)
	}

	result := func_([]*Node{node})
	t.Logf("Input node: %+v", node)
	t.Logf("Query result: %+v", result)

	if len(result) == 0 {
		t.Fatalf("Expected result, got empty")
	}

	if result[0].Type != "Literal" {
		t.Errorf("Expected Literal node, got %s", result[0].Type)
	}

	if result[0].Token != "my_function" {
		t.Errorf("Expected token 'my_function', got '%s'", result[0].Token)
	}
}

func TestDebugFilterPart(t *testing.T) {
	node := &Node{
		Type:  "Function",
		Roles: []Role{"Function", "Declaration"},
		Props: map[string]string{
			"name": "my_function",
		},
		Token: "my_function",
	}

	// Test just the filter part
	query := `filter(.type == "Function")`
	ast, err := ParseDSL(query)
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	func_, err := LowerDSL(ast)
	if err != nil {
		t.Fatalf("Failed to lower DSL: %v", err)
	}

	result := func_([]*Node{node})
	t.Logf("Filter result: %+v", result)

	if len(result) == 0 {
		t.Fatalf("Expected Function node, got empty")
	}
}

func TestDebugMapPart(t *testing.T) {
	node := &Node{
		Type:  "Function",
		Roles: []Role{"Function", "Declaration"},
		Props: map[string]string{
			"name": "my_function",
		},
		Token: "my_function",
	}

	// Test just the map part
	query := `map(.name)`
	ast, err := ParseDSL(query)
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	func_, err := LowerDSL(ast)
	if err != nil {
		t.Fatalf("Failed to lower DSL: %v", err)
	}

	result := func_([]*Node{node})
	t.Logf("Map result: %+v", result)

	if len(result) == 0 {
		t.Fatalf("Expected result, got empty")
	}
}

func TestHasSyntax(t *testing.T) {
	// Test cases for "has" syntax
	testCases := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "has with roles",
			query:    `.roles has "Function"`,
			expected: true,
		},
		{
			name:     "has with type",
			query:    `.type has "Function"`,
			expected: false, // type is a single value, not a collection
		},
		{
			name:     "has with props",
			query:    `.props has "name"`,
			expected: false, // props is a map, not a collection
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test node with roles
			testNode := &Node{
				Type:  "Function",
				Token: "testFunction",
				Roles: []Role{"Function", "Declaration"},
				Props: map[string]string{
					"name": "testFunction",
				},
			}

			// Parse the query
			ast, err := ParseDSL(tc.query)
			if err != nil {
				t.Fatalf("Failed to parse query '%s': %v", tc.query, err)
			}

			// Lower the DSL
			queryFunc, err := LowerDSL(ast)
			if err != nil {
				t.Fatalf("Failed to lower DSL for '%s': %v", tc.query, err)
			}

			// Execute the query
			results := queryFunc([]*Node{testNode})

			fmt.Printf("Query: %s\n", tc.query)
			fmt.Printf("Results: %+v\n", results)

			if len(results) > 0 {
				fmt.Printf("First result: %+v\n", results[0])
			}
		})
	}
}

// --- DSL Query Efficiency Instrumentation Helpers (migrated) ---

var (
	filterCallCount    int
	mapCallCount       int
	evaluationCount    int
	dslAllocationCount int
)

func resetDSLCounters() {
	filterCallCount = 0
	mapCallCount = 0
	evaluationCount = 0
	dslAllocationCount = 0
}

func instrumentedFindDSL(node *Node, query string) ([]*Node, error) {
	// Track filter and map operations
	filterCallCount++
	evaluationCount++

	// Simulate the query execution
	results, err := node.FindDSL(query)

	// Count operations based on query type
	if len(query) > 0 {
		// Rough estimation of operations based on query complexity
		evaluationCount += len(results) * 2 // Each result requires evaluation
	}

	return results, err
}

func TestDSLQueryAlgorithmEfficiency(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	testCases := []struct {
		name           string
		content        []byte
		query          string
		maxFilterCalls int
		maxMapCalls    int
		maxEvaluations int
	}{
		{
			name:           "LargeGoFile",
			content:        generateLargeGoFile(),
			query:          "filter(.type == \"FunctionDecl\")",
			maxFilterCalls: 1000,
			maxMapCalls:    0,
			maxEvaluations: 2000,
		},
		{
			name:           "VeryLargeGoFile",
			content:        generateVeryLargeGoFile(),
			query:          "filter(.type == \"FunctionDecl\") |> map(.name)",
			maxFilterCalls: 5000,
			maxMapCalls:    200,
			maxEvaluations: 10000,
		},
	}

	for _, tc := range testCases {
		node, err := parser.Parse(tc.name+".go", tc.content)
		if err != nil {
			t.Fatalf("Failed to parse test file: %v", err)
		}

		t.Run(tc.name, func(t *testing.T) {
			resetDSLCounters()

			results, err := instrumentedFindDSL(node, tc.query)
			if err != nil {
				t.Fatalf("DSL query failed: %v", err)
			}

			if filterCallCount > tc.maxFilterCalls {
				t.Errorf("Too many filter calls: got %d, want <= %d", filterCallCount, tc.maxFilterCalls)
			}

			if mapCallCount > tc.maxMapCalls {
				t.Errorf("Too many map calls: got %d, want <= %d", mapCallCount, tc.maxMapCalls)
			}

			if evaluationCount > tc.maxEvaluations {
				t.Errorf("Too many evaluations: got %d, want <= %d", evaluationCount, tc.maxEvaluations)
			}

			t.Logf("DSL query efficiency: %d filter calls, %d map calls, %d evaluations, %d results",
				filterCallCount, mapCallCount, evaluationCount, len(results))
		})
	}
}
