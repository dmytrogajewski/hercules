package uast

import (
	"testing"
)

// Remove placeholder AST node types and ParseDSL from this file; use real ones from dsl_parser.go

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
		{"filter(.type == \"FunctionDecl\") |> map(.foo)", testUAST, []string{"qux"}, false},
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
	dsl := "filter(.type == \"go:function\") |> map(.type)"
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
	want := []string{"go:function", "go:function"}
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
	// Print Props and Token for all go:function and go:method nodes
	for _, n := range nodes {
		if n.Type == "go:function" || n.Type == "go:method" {
			t.Logf("%s Props: %v, Token: %q", n.Type, n.Props, n.Token)
		}
	}
	// Query: get all function/method names
	dsl := "filter(.type == \"go:function\" || .type == \"go:method\") |> map(.name)"
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
