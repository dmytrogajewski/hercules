package uast

import (
	"go/ast"
	"testing"

	sitter "github.com/alexaandru/go-tree-sitter-bare"
)

func TestGoEmbeddedProvider_Parse_ValidGo(t *testing.T) {
	provider := &GoEmbeddedProvider{}
	code := []byte(`package main
func main() {}`)
	node, err := provider.Parse("main.go", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := node.(*ast.File); !ok {
		t.Fatalf("expected *ast.File, got %T", node)
	}
}

func TestGoEmbeddedProvider_Parse_InvalidGo(t *testing.T) {
	provider := &GoEmbeddedProvider{}
	code := []byte(`package main
func main() {`)
	_, err := provider.Parse("main.go", code)
	if err == nil {
		t.Fatal("expected error for invalid Go code, got nil")
	}
}

func TestTreeSitterJavaProvider_Parse_ValidJava(t *testing.T) {
	provider := &TreeSitterJavaProvider{}
	code := []byte(`public class Hello { public static void main(String[] args) { System.out.println("Hello, world!"); } }`)
	node, err := provider.Parse("Hello.java", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	res, ok := node.(*TreeSitterResult)
	if !ok {
		t.Fatalf("expected *TreeSitterResult, got %T", node)
	}
	if (res.Root == sitter.Node{}) {
		t.Fatalf("expected Root to be a valid sitter.Node, got zero value")
	}
}

func TestTreeSitterKotlinProvider_Parse_ValidKotlin(t *testing.T) {
	provider := &TreeSitterKotlinProvider{}
	code := []byte(`class Hello { fun main() { println("Hello, world!") } }`)
	node, err := provider.Parse("Hello.kt", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	res, ok := node.(*TreeSitterResult)
	if !ok {
		t.Fatalf("expected *TreeSitterResult, got %T", node)
	}
	if (res.Root == sitter.Node{}) {
		t.Fatalf("expected Root to be a valid sitter.Node, got zero value")
	}
}

func TestTreeSitterSwiftProvider_Parse_ValidSwift(t *testing.T) {
	provider := &TreeSitterSwiftProvider{}
	code := []byte(`class Hello { func main() { print("Hello, world!") } }`)
	node, err := provider.Parse("Hello.swift", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	res, ok := node.(*TreeSitterResult)
	if !ok {
		t.Fatalf("expected *TreeSitterResult, got %T", node)
	}
	if (res.Root == sitter.Node{}) {
		t.Fatalf("expected Root to be a valid sitter.Node, got zero value")
	}
}

func TestTreeSitterJavaScriptProvider_Parse_ValidJS(t *testing.T) {
	provider := &TreeSitterJavaScriptProvider{}
	code := []byte(`class Hello { main() { console.log("Hello, world!") } }`)
	node, err := provider.Parse("Hello.js", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	res, ok := node.(*TreeSitterResult)
	if !ok {
		t.Fatalf("expected *TreeSitterResult, got %T", node)
	}
	if (res.Root == sitter.Node{}) {
		t.Fatalf("expected Root to be a valid sitter.Node, got zero value")
	}
}

func TestTreeSitterRustProvider_Parse_ValidRust(t *testing.T) {
	provider := &TreeSitterRustProvider{}
	code := []byte(`struct Hello { value: i32 } fn main() { println!("Hello, world!"); }`)
	node, err := provider.Parse("hello.rs", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	res, ok := node.(*TreeSitterResult)
	if !ok {
		t.Fatalf("expected *TreeSitterResult, got %T", node)
	}
	if (res.Root == sitter.Node{}) {
		t.Fatalf("expected Root to be a valid sitter.Node, got zero value")
	}
}

func TestTreeSitterPHPProvider_Parse_ValidPHP(t *testing.T) {
	provider := &TreeSitterPHPProvider{}
	code := []byte(`<?php class Hello { function main() { echo "Hello, world!"; } }`)
	node, err := provider.Parse("hello.php", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	res, ok := node.(*TreeSitterResult)
	if !ok {
		t.Fatalf("expected *TreeSitterResult, got %T", node)
	}
	if (res.Root == sitter.Node{}) {
		t.Fatalf("expected Root to be a valid sitter.Node, got zero value")
	}
}

func TestTreeSitterPythonProvider_Parse_ValidPython(t *testing.T) {
	provider := &TreeSitterPythonProvider{}
	code := []byte(`class Hello:\n    def main(self):\n        print("Hello, world!")`)
	node, err := provider.Parse("hello.py", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	res, ok := node.(*TreeSitterResult)
	if !ok {
		t.Fatalf("expected *TreeSitterResult, got %T", node)
	}
	if (res.Root == sitter.Node{}) {
		t.Fatalf("expected Root to be a valid sitter.Node, got zero value")
	}
}

func TestTreeSitterTypeScriptProvider_Parse_ValidTS(t *testing.T) {
	provider := &TreeSitterTypeScriptProvider{}
	code := []byte(`class Hello { main(): void { console.log("Hello, TypeScript!"); } }`)
	node, err := provider.Parse("Hello.ts", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	res, ok := node.(*TreeSitterResult)
	if !ok {
		t.Fatalf("expected *TreeSitterResult, got %T", node)
	}
	if (res.Root == sitter.Node{}) {
		t.Fatalf("expected Root to be a valid sitter.Node, got zero value")
	}
}

func TestTreeSitterTSXProvider_Parse_ValidTSX(t *testing.T) {
	provider := &TreeSitterTSXProvider{}
	code := []byte(`const Hello = () => (<div>Hello, TSX!</div>);`)
	node, err := provider.Parse("Hello.tsx", code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	res, ok := node.(*TreeSitterResult)
	if !ok {
		t.Fatalf("expected *TreeSitterResult, got %T", node)
	}
	if (res.Root == sitter.Node{}) {
		t.Fatalf("expected Root to be a valid sitter.Node, got zero value")
	}
}
