package uast

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"

	tsjava "github.com/alexaandru/go-sitter-forest/java"
	tsjs "github.com/alexaandru/go-sitter-forest/javascript"
	tskt "github.com/alexaandru/go-sitter-forest/kotlin"
	tsphp "github.com/alexaandru/go-sitter-forest/php"
	tspy "github.com/alexaandru/go-sitter-forest/python"
	tsrs "github.com/alexaandru/go-sitter-forest/rust"
	tsswift "github.com/alexaandru/go-sitter-forest/swift"
	tstsx "github.com/alexaandru/go-sitter-forest/tsx"
	tsts "github.com/alexaandru/go-sitter-forest/typescript"
	sitter "github.com/alexaandru/go-tree-sitter-bare"
)

// UASTNode is a generic interface for UAST nodes (compatible with Hercules expectations)
type UASTNode interface{}

// UASTProvider abstracts UAST extraction from source code
// (Babelfish, embedded, or other backends)
type UASTProvider interface {
	Parse(filename string, content []byte) (UASTNode, error)
}

// GoEmbeddedProvider implements UASTProvider for Go code using go/parser
type GoEmbeddedProvider struct{}

func (p *GoEmbeddedProvider) Parse(filename string, content []byte) (UASTNode, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("go/parser error: %w", err)
	}
	return file, nil // *ast.File is the root node
}

// TreeSitterResult holds both the tree and the root node to ensure the tree stays alive.
type TreeSitterResult struct {
	Tree *sitter.Tree
	Root sitter.Node
}

// TreeSitterJavaProvider implements UASTProvider for Java using Tree-sitter
// Use for .java files
// Returns the root *sitter.Node
type TreeSitterJavaProvider struct{}

func (p *TreeSitterJavaProvider) Parse(filename string, content []byte) (UASTNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tsjava.GetLanguage()))
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	return &TreeSitterResult{Tree: tree, Root: tree.RootNode()}, nil
}

// TreeSitterKotlinProvider implements UASTProvider for Kotlin using Tree-sitter
// Use for .kt files
// Returns the root *sitter.Node
type TreeSitterKotlinProvider struct{}

func (p *TreeSitterKotlinProvider) Parse(filename string, content []byte) (UASTNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tskt.GetLanguage()))
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	return &TreeSitterResult{Tree: tree, Root: tree.RootNode()}, nil
}

// TreeSitterSwiftProvider implements UASTProvider for Swift using Tree-sitter
// Use for .swift files
// Returns the root *sitter.Node
type TreeSitterSwiftProvider struct{}

func (p *TreeSitterSwiftProvider) Parse(filename string, content []byte) (UASTNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tsswift.GetLanguage()))
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	return &TreeSitterResult{Tree: tree, Root: tree.RootNode()}, nil
}

// TreeSitterJavaScriptProvider implements UASTProvider for JavaScript/JSX using Tree-sitter
// Use for .js/.jsx files
// Returns the root *sitter.Node
type TreeSitterJavaScriptProvider struct{}

func (p *TreeSitterJavaScriptProvider) Parse(filename string, content []byte) (UASTNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tsjs.GetLanguage()))
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	return &TreeSitterResult{Tree: tree, Root: tree.RootNode()}, nil
}

// TreeSitterRustProvider implements UASTProvider for Rust using Tree-sitter
// Use for .rs files
// Returns the root *sitter.Node
type TreeSitterRustProvider struct{}

func (p *TreeSitterRustProvider) Parse(filename string, content []byte) (UASTNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tsrs.GetLanguage()))
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	return &TreeSitterResult{Tree: tree, Root: tree.RootNode()}, nil
}

// TreeSitterPHPProvider implements UASTProvider for PHP using Tree-sitter
// Use for .php files
// Returns the root *sitter.Node
type TreeSitterPHPProvider struct{}

func (p *TreeSitterPHPProvider) Parse(filename string, content []byte) (UASTNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tsphp.GetLanguage()))
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	return &TreeSitterResult{Tree: tree, Root: tree.RootNode()}, nil
}

// TreeSitterPythonProvider implements UASTProvider for Python using Tree-sitter
// Use for .py files
// Returns the root *sitter.Node
type TreeSitterPythonProvider struct{}

func (p *TreeSitterPythonProvider) Parse(filename string, content []byte) (UASTNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tspy.GetLanguage()))
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	return &TreeSitterResult{Tree: tree, Root: tree.RootNode()}, nil
}

// TreeSitterTypeScriptProvider implements UASTProvider for TypeScript using Tree-sitter
// Use for .ts files
// Returns the root sitter.Node
type TreeSitterTypeScriptProvider struct{}

func (p *TreeSitterTypeScriptProvider) Parse(filename string, content []byte) (UASTNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tsts.GetLanguage()))
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	return &TreeSitterResult{Tree: tree, Root: tree.RootNode()}, nil
}

// TreeSitterTSXProvider implements UASTProvider for TSX using Tree-sitter
// Use for .tsx files
// Returns the root sitter.Node
type TreeSitterTSXProvider struct{}

func (p *TreeSitterTSXProvider) Parse(filename string, content []byte) (UASTNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(sitter.NewLanguage(tstsx.GetLanguage()))
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, err
	}
	return &TreeSitterResult{Tree: tree, Root: tree.RootNode()}, nil
}

// BabelfishProvider is a stub for CLI logic; actual Babelfish logic is handled in Extractor when Provider is nil.
type BabelfishProvider struct{}

func (p *BabelfishProvider) Parse(filename string, content []byte) (UASTNode, error) {
	return nil, fmt.Errorf("BabelfishProvider is a stub; Babelfish is handled natively in Extractor")
}
