package uast

import (
	"encoding/json"
	"fmt"
)

// Role represents a syntactic/semantic label for a node.
type Role string

// Positions represents the byte and line/col offsets for a node.
type Positions struct {
	StartLine   int `json:"start_line,omitempty"`
	StartCol    int `json:"start_col,omitempty"`
	StartOffset int `json:"start_offset,omitempty"`
	EndLine     int `json:"end_line,omitempty"`
	EndCol      int `json:"end_col,omitempty"`
	EndOffset   int `json:"end_offset,omitempty"`
}

// Node is the canonical UAST node structure.
type Node struct {
	Id       uint64            `json:"id,omitempty"`
	Type     string            `json:"type,omitempty"`
	Token    string            `json:"token,omitempty"`
	Roles    []Role            `json:"roles,omitempty"`
	Pos      *Positions        `json:"pos,omitempty"`
	Props    map[string]string `json:"props,omitempty"`
	Children []*Node           `json:"children,omitempty"`
}

// AddChild appends a child node.
func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
}

// RemoveChild removes the first occurrence of the given child node.
func (n *Node) RemoveChild(child *Node) bool {
	for i, c := range n.Children {
		if c == child {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			return true
		}
	}
	return false
}

// String returns a string representation of the node.
func (n *Node) String() string {
	b, err := json.Marshal(n)
	if err != nil {
		return fmt.Sprintf("Node<error: %v>", err)
	}
	return string(b)
}

// Find returns all nodes in the tree (including root) for which predicate(node) is true.
func (n *Node) Find(predicate func(*Node) bool) []*Node {
	if n == nil {
		return nil
	}
	var result []*Node
	stack := []*Node{n}
	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if predicate(curr) {
			result = append(result, curr)
		}
		for i := len(curr.Children) - 1; i >= 0; i-- {
			stack = append(stack, curr.Children[i])
		}
	}
	return result
}

// PreOrder visits all nodes in pre-order (root, then children left-to-right).
func (n *Node) PreOrder(fn func(*Node)) {
	if n == nil {
		return
	}
	stack := []*Node{n}
	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		fn(curr)
		for i := len(curr.Children) - 1; i >= 0; i-- {
			stack = append(stack, curr.Children[i])
		}
	}
}

// PostOrder visits all nodes in post-order (children left-to-right, then root).
func (n *Node) PostOrder(fn func(*Node)) {
	if n == nil {
		return
	}
	type frame struct {
		node    *Node
		visited bool
	}
	stack := []frame{{n, false}}
	for len(stack) > 0 {
		top := &stack[len(stack)-1]
		if !top.visited {
			top.visited = true
			for i := len(top.node.Children) - 1; i >= 0; i-- {
				stack = append(stack, frame{top.node.Children[i], false})
			}
		} else {
			fn(top.node)
			stack = stack[:len(stack)-1]
		}
	}
}

// Ancestors returns the list of ancestors from root to the parent of target (empty if not found).
func (n *Node) Ancestors(target *Node) []*Node {
	if n == nil || target == nil {
		return nil
	}
	type frame struct {
		node   *Node
		parent []*Node
	}
	stack := []frame{{n, nil}}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if top.node == target {
			return top.parent
		}
		for i := len(top.node.Children) - 1; i >= 0; i-- {
			child := top.node.Children[i]
			anc := append([]*Node{}, top.parent...)
			anc = append(anc, top.node)
			stack = append(stack, frame{child, anc})
		}
	}
	return nil
}

// Transform returns a new tree where each node is replaced by the result of fn(node) (post-order, non-recursive).
func (n *Node) Transform(fn func(*Node) *Node) *Node {
	if n == nil {
		return nil
	}
	type frame struct {
		node     *Node
		parent   *Node
		childIdx int
		newNode  *Node
	}
	var (
		stack   []frame
		results = make(map[*Node]*Node)
	)
	stack = append(stack, frame{n, nil, 0, nil})
	for len(stack) > 0 {
		top := &stack[len(stack)-1]
		if top.childIdx < len(top.node.Children) {
			child := top.node.Children[top.childIdx]
			stack = append(stack, frame{child, top.node, 0, nil})
			top.childIdx++
			continue
		}
		// All children processed
		copy := *top.node
		copy.Children = make([]*Node, len(top.node.Children))
		for i, c := range top.node.Children {
			copy.Children[i] = results[c]
		}
		results[top.node] = fn(&copy)
		stack = stack[:len(stack)-1]
	}
	return results[n]
}

// ReplaceChild replaces the first occurrence of old in Children with new. Returns true if replaced.
func (n *Node) ReplaceChild(old, new *Node) bool {
	for i, c := range n.Children {
		if c == old {
			n.Children[i] = new
			return true
		}
	}
	return false
}
