package uast

import (
	"testing"
)

func TestIntegration_GoFunctionUAST_SPEC(t *testing.T) {
	src := []byte(`package main
func add(a, b int) int { return a + b }`)
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("failed to create parser: %v", err)
	}
	node, err := parser.Parse("main.go", src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if node == nil {
		t.Fatalf("Parse returned nil node")
	}
	// Find the function node
	var fn *Node
	for _, child := range node.Children {
		if child.Type == "go:function" || child.Type == "Function" || child.Type == "FunctionDecl" {
			fn = child
			break
		}
	}
	if fn == nil {
		t.Fatalf("No function node found; got children: %+v", node.Children)
	}
	// Check canonical type
	if fn.Type != "go:function" && fn.Type != "Function" && fn.Type != "FunctionDecl" {
		t.Errorf("Function node has wrong type: got %q", fn.Type)
	}
	// Check roles
	wantRoles := map[string]bool{"Function": true, "Declaration": true}
	for _, r := range fn.Roles {
		delete(wantRoles, string(r))
	}
	for missing := range wantRoles {
		t.Errorf("Function node missing role: %s", missing)
	}
	// Check props
	if fn.Props["name"] != "add" {
		t.Errorf("Function node has wrong name prop: got %q, want 'add'", fn.Props["name"])
	}
	// Check children are present
	if len(fn.Children) == 0 {
		t.Errorf("Function node has no children")
	}
}
