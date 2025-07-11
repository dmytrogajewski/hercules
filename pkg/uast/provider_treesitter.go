package uast

import (
	"context"
	"fmt"
	"strings"

	sitter "github.com/alexaandru/go-tree-sitter-bare"
)

type TreeSitterProvider struct {
	language *sitter.Language
	langName string
	mapping  map[string]Mapping // kind -> Mapping
}

func (p *TreeSitterProvider) Parse(filename string, content []byte) (*Node, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(p.language)

	if len(content) == 0 || len(strings.TrimSpace(string(content))) == 0 {
		return &Node{
			Type:     fmt.Sprintf("%s:file", p.langName),
			Token:    "",
			Pos:      &Positions{StartLine: 0, StartCol: 0, StartOffset: 0, EndLine: 0, EndCol: 0, EndOffset: 0},
			Children: []*Node{},
		}, nil
	}

	tree, err := parser.ParseString(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	root := tree.RootNode()
	if !root.IsNamed() || root.ChildCount() == 0 {
		return &Node{
			Type:     fmt.Sprintf("%s:file", p.langName),
			Token:    filename,
			Pos:      &Positions{StartLine: 0, StartCol: 0, StartOffset: 0, EndLine: 0, EndCol: 0, EndOffset: 0},
			Children: []*Node{},
		}, nil
	}

	// --- DEBUG: Print AST node types ---
	fmt.Printf("[DEBUG] Tree-sitter root node type: %s\n", root.Type())
	for i := uint32(0); i < root.NamedChildCount(); i++ {
		child := root.NamedChild(i)
		fmt.Printf("[DEBUG] Child %d: type=%s, start=%d, end=%d\n", i, child.Type(), child.StartByte(), child.EndByte())
	}
	// --- END DEBUG ---

	tsNode := &TreeSitterNode{
		Root:     root,
		Tree:     tree,
		Language: p.langName,
		Source:   content,
		Mapping:  p.mapping,
	}

	return tsNode.ToCanonicalNode(), nil
}

func (p *TreeSitterProvider) Language() string {
	return p.langName
}

type TreeSitterNode struct {
	Root     sitter.Node
	Tree     *sitter.Tree
	Language string
	Source   []byte
	Mapping  map[string]Mapping // kind -> Mapping
}

func (n *TreeSitterNode) ToCanonicalNode() *Node {
	kind := n.Root.Type()
	mapping, hasMapping := n.Mapping[kind]
	children := make([]*Node, 0)
	for i := uint32(0); i < n.Root.NamedChildCount(); i++ {
		child := n.Root.NamedChild(i)
		childNode := &TreeSitterNode{
			Root:     child,
			Tree:     n.Tree,
			Language: n.Language,
			Source:   n.Source,
			Mapping:  n.Mapping,
		}
		children = append(children, childNode.ToCanonicalNode())
	}
	props := map[string]string{}
	var roles []Role

	// SPEC: For the root node (source_file), always use 'lang:file' as the canonical type
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
	}
	return &Node{
		Type:     typeStr,
		Token:    n.Token(),
		Pos:      n.Positions(),
		Props:    props,
		Roles:    roles,
		Children: children,
	}
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
