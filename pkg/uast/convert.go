package uast

import sitter "github.com/alexaandru/go-tree-sitter-bare"

// FromTree converts a Tree-sitter parse tree to a canonical UAST Node.
func FromTree(tree *sitter.Tree, code []byte, lang string) *Node {
	if tree == nil {
		return nil
	}
	root := tree.RootNode()
	if !root.IsNamed() {
		return nil
	}
	type frame struct {
		tsNode   sitter.Node
		uastNode *Node
		parent   *Node
	}
	var (
		stack   []frame
		nodeMap = make(map[sitter.Node]*Node)
	)
	// Create root node
	rootUAST := &Node{
		Type: lang + ":" + root.Type(),
		Pos: &Positions{
			StartLine:   int(root.StartPoint().Row),
			StartCol:    int(root.StartPoint().Column),
			StartOffset: int(root.StartByte()),
			EndLine:     int(root.EndPoint().Row),
			EndCol:      int(root.EndPoint().Column),
			EndOffset:   int(root.EndByte()),
		},
	}
	stack = append(stack, frame{root, rootUAST, nil})
	for len(stack) > 0 {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		children := make([]*Node, 0, f.tsNode.NamedChildCount())
		for i := uint32(0); i < f.tsNode.NamedChildCount(); i++ {
			child := f.tsNode.NamedChild(i)
			childNode := &Node{
				Type: lang + ":" + child.Type(),
				Pos: &Positions{
					StartLine:   int(child.StartPoint().Row),
					StartCol:    int(child.StartPoint().Column),
					StartOffset: int(child.StartByte()),
					EndLine:     int(child.EndPoint().Row),
					EndCol:      int(child.EndPoint().Column),
					EndOffset:   int(child.EndByte()),
				},
			}
			// If leaf, set Token
			if child.NamedChildCount() == 0 && child.ChildCount() == 0 {
				start := child.StartByte()
				end := child.EndByte()
				if int(end) <= len(code) && int(start) < int(end) {
					childNode.Token = string(code[start:end])
				}
			}
			children = append(children, childNode)
			stack = append(stack, frame{child, childNode, f.uastNode})
		}
		f.uastNode.Children = children
		nodeMap[f.tsNode] = f.uastNode
	}
	return rootUAST
}
