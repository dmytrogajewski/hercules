package uast

import (
	"context"
	"errors"
	"fmt"

	sitter "github.com/alexaandru/go-tree-sitter-bare"
)

// TreeSitterProvider implements the UAST provider interface using Tree-sitter.
type TreeSitterProvider struct {
	language        *sitter.Language
	langName        string
	mapping         map[string]Mapping // kind -> Mapping
	IncludeUnmapped bool
}

// Parse parses the given file content and returns the root UAST node.
// Returns an error if parsing fails.
func (p *TreeSitterProvider) Parse(filename string, content []byte) (*Node, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(p.language)
	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("tree-sitter: failed to parse: %w", err)
	}
	root := tree.RootNode()
	if root.IsNull() {
		return nil, errors.New("tree-sitter: no root node")
	}

	tsNode := &TreeSitterNode{
		Root:            root,
		Tree:            tree,
		Language:        p.langName,
		Source:          content,
		Mapping:         p.mapping,
		IncludeUnmapped: p.IncludeUnmapped,
	}
	canonical := tsNode.ToCanonicalNode()
	if canonical == nil {
		return nil, nil
	}
	return canonical, nil
}

// Language returns the language name for this provider.
func (p *TreeSitterProvider) Language() string {
	return p.langName
}

// TreeSitterNode wraps a Tree-sitter node for conversion to UAST.
type TreeSitterNode struct {
	Root            sitter.Node
	Tree            *sitter.Tree
	Language        string
	Source          []byte
	Mapping         map[string]Mapping // kind -> Mapping
	IncludeUnmapped bool
}

// ToCanonicalNode converts the TreeSitterNode to a canonical UAST Node.
func (n *TreeSitterNode) ToCanonicalNode() *Node {
	kind := n.Root.Type()
	mapping, hasMapping := n.Mapping[kind]
	childCount := n.Root.NamedChildCount()
	children := make([]*Node, 0, childCount)
	for i := uint32(0); i < childCount; i++ {
		child := n.Root.NamedChild(i)
		childNode := &TreeSitterNode{
			Root:            child,
			Tree:            n.Tree,
			Language:        n.Language,
			Source:          n.Source,
			Mapping:         n.Mapping,
			IncludeUnmapped: n.IncludeUnmapped,
		}
		c := childNode.ToCanonicalNode()
		if c != nil {
			children = append(children, c)
		}
	}
	props := map[string]string{}
	var roles []Role
	// Relaxed: only skip if file is truly empty (no content)
	if kind == "source_file" && hasMapping && mapping.SkipIfEmpty && len(children) == 0 && len(n.Source) == 0 {
		return nil
	}
	typeStr := n.Language + ":" + kind
	if kind == "source_file" {
		typeStr = n.Language + ":file"
	}
	if hasMapping {
		typeStr = mapping.Type
		for _, r := range mapping.Roles {
			roles = append(roles, Role(r))
		}
		for propKey, propVal := range mapping.Props {
			if propStr, ok := propVal.(string); ok {
				for i := uint32(0); i < n.Root.NamedChildCount(); i++ {
					c := n.Root.NamedChild(i)
					childKind := c.Type()
					childMapping, ok := n.Mapping[childKind]
					if ok && (childKind == propStr || containsString(childMapping.Roles, propStr)) {
						start := c.StartByte()
						end := c.EndByte()
						if int(end) <= len(n.Source) {
							props[propKey] = string(n.Source[start:end])
							break
						}
					}
				}
			}
		}
		node := NewNode(0, typeStr, n.Token(), roles, n.Positions(), props)
		node.Children = children
		return node
	}
	if n.IncludeUnmapped || kind == "source_file" {
		node := NewNode(0, typeStr, n.Token(), roles, n.Positions(), props)
		node.Children = children
		return node
	}
	return nil
}

// containsString checks if a slice contains a string
func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// Token returns the string token for this node, if any.
func (n *TreeSitterNode) Token() string {
	if n.Root.ChildCount() == 0 {
		start := n.Root.StartByte()
		end := n.Root.EndByte()
		if int(end) <= len(n.Source) {
			return string(n.Source[start:end])
		}
	}
	return ""
}

// Positions returns the source code positions for this node.
func (n *TreeSitterNode) Positions() *Positions {
	return &Positions{
		StartLine:   int(n.Root.StartPoint().Row),
		StartCol:    int(n.Root.StartPoint().Column),
		StartOffset: int(n.Root.StartByte()),
		EndLine:     int(n.Root.EndPoint().Row),
		EndCol:      int(n.Root.EndPoint().Column),
		EndOffset:   int(n.Root.EndByte()),
	}
}
