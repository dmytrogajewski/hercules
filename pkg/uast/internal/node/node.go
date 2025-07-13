// Package uast provides a universal abstract syntax tree (UAST) representation and utilities for parsing, navigating, querying, and mutating code structure in a language-agnostic way.
package node

import (
	"fmt"
	"strings"
	"sync"
)

// Role represents a syntactic/semantic label for a node.
type Role string

// Positions represents the byte and line/col offsets for a node.
// All fields are 1-based except StartOffset/EndOffset, which are byte offsets.
type Positions struct {
	StartLine   int `json:"start_line,omitempty"`
	StartCol    int `json:"start_col,omitempty"`
	StartOffset int `json:"start_offset,omitempty"`
	EndLine     int `json:"end_line,omitempty"`
	EndCol      int `json:"end_col,omitempty"`
	EndOffset   int `json:"end_offset,omitempty"`
}

// Node is the canonical UAST node structure.
//
// Fields:
//
//	Id: unique node identifier (optional)
//	Type: node type (e.g., "Function", "Identifier")
//	Token: string value or token for leaf nodes
//	Roles: semantic/syntactic roles (see Role)
//	Pos: source code position info (optional)
//	Props: additional properties (language-specific)
//	Children: child nodes (ordered)
type Node struct {
	Id       uint64            `json:"id,omitempty"`
	Type     string            `json:"type,omitempty"`
	Token    string            `json:"token,omitempty"`
	Roles    []Role            `json:"roles,omitempty"`
	Pos      *Positions        `json:"pos,omitempty"`
	Props    map[string]string `json:"props,omitempty"`
	Children []*Node           `json:"children,omitempty"`
}

// nodePool is a sync.Pool for Node structs to reduce allocation overhead
var nodePool = sync.Pool{
	New: func() interface{} {
		return &Node{}
	},
}

// New creates a new Node from the pool and initializes it with the given values
func New(id uint64, nodeType, token string, roles []Role, pos *Positions, props map[string]string) *Node {
	node := nodePool.Get().(*Node)
	node.Id = id
	node.Type = nodeType
	node.Token = token
	node.Roles = roles
	node.Pos = pos
	node.Props = props
	node.Children = make([]*Node, 0, 4) // Pre-allocate with reasonable capacity
	return node
}

// NewWithType creates a new Node with just a type
func NewWithType(nodeType string) *Node {
	node := nodePool.Get().(*Node)
	node.Id = 0
	node.Type = nodeType
	node.Token = ""
	node.Roles = nil
	node.Pos = nil
	node.Props = nil
	node.Children = nil
	return node
}

// NewNodeWithToken creates a new Node with type and token
func NewNodeWithToken(nodeType, token string) *Node {
	node := nodePool.Get().(*Node)
	node.Id = 0
	node.Type = nodeType
	node.Token = token
	node.Roles = nil
	node.Pos = nil
	node.Props = nil
	node.Children = nil
	return node
}

// NewLiteralNode creates a new Node for literal values
func NewLiteralNode(token string) *Node {
	return NewNodeWithToken("Literal", token)
}

// ReleaseNode returns a Node to the pool for reuse
func ReleaseNode(node *Node) {
	if node == nil {
		return
	}
	// Clear the node to prevent memory leaks
	node.Id = 0
	node.Type = ""
	node.Token = ""
	node.Roles = nil
	node.Pos = nil
	node.Props = nil
	node.Children = nil
	nodePool.Put(node)
}

// ReleaseNodes returns multiple nodes to the pool
func ReleaseNodes(nodes []*Node) {
	for _, node := range nodes {
		ReleaseNode(node)
	}
}

type nodeFrame struct {
	node    *Node
	visited bool
}

type nodeAncestorFrame struct {
	node   *Node
	parent []*Node
}

type nodeTransformFrame struct {
	node     *Node
	parent   *Node
	childIdx int
	newNode  *Node
}

// AddChild appends a child node to n.
func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
}

// RemoveChild removes the first occurrence of the given child node from n.
// Returns true if the child was found and removed.
func (n *Node) RemoveChild(child *Node) bool {
	for i, c := range n.Children {
		if isChildMatch(c, child) {
			removeChildAtIndex(n, i)
			return true
		}
	}
	return false
}

func isChildMatch(child, target *Node) bool {
	return child == target
}

func removeChildAtIndex(n *Node, index int) {
	n.Children = append(n.Children[:index], n.Children[index+1:]...)
}

// String returns a string representation of the node
func (n *Node) String() string {
	return optimizedNodeString(n)
}

// Optimized string representation without JSON marshaling
func optimizedNodeString(node *Node) string {
	if node == nil {
		return "nil"
	}

	var buf strings.Builder
	buf.WriteString("Node{")
	buf.WriteString("Type:")
	buf.WriteString(node.Type)

	if node.Token != "" {
		buf.WriteString(",Token:")
		buf.WriteString(node.Token)
	}

	if len(node.Roles) > 0 {
		buf.WriteString(",Roles:[")
		for i, role := range node.Roles {
			if i > 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(string(role))
		}
		buf.WriteString("]")
	}

	if len(node.Props) > 0 {
		buf.WriteString(",Props:")
		buf.WriteString(fmt.Sprintf("%v", node.Props))
	}

	if len(node.Children) > 0 {
		buf.WriteString(",Children:")
		buf.WriteString(fmt.Sprintf("%d", len(node.Children)))
	}

	buf.WriteString("}")
	return buf.String()
}

func isJsonError(err error) bool {
	return err != nil
}

func createErrorString(err error) string {
	return fmt.Sprintf("Node<error: %v>", err)
}

// Find returns all nodes in the tree (including root) for which predicate(node) is true.
// Traversal is pre-order. Returns nil if n is nil.
func (n *Node) Find(predicate func(*Node) bool) []*Node {
	if isNodeNil(n) {
		return nil
	}
	return findNodesWithPredicate(n, predicate)
}

func isNodeNil(n *Node) bool {
	return n == nil
}

func findNodesWithPredicate(n *Node, predicate func(*Node) bool) []*Node {
	var result []*Node
	stack := []*Node{n}
	for nodeHasStack(stack) {
		curr := nodePopStack(&stack)
		if predicate(curr) {
			result = append(result, curr)
		}
		nodePushChildrenToStack(curr, &stack)
	}
	return result
}

func nodeHasStack(stack []*Node) bool {
	return len(stack) > 0
}

func nodePopStack(stack *[]*Node) *Node {
	last := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return last
}

func nodePushChildrenToStack(node *Node, stack *[]*Node) {
	for i := len(node.Children) - 1; i >= 0; i-- {
		*stack = append(*stack, node.Children[i])
	}
}

func estimateTreeSize(node *Node) int {
	if node == nil {
		return 0
	}
	// Estimate based on number of children and their potential children
	size := 1
	for _, child := range node.Children {
		size += estimateTreeSize(child)
	}
	return size
}

// VisitPreOrder visits all nodes in pre-order (root, then children left-to-right).
// Now uses the final optimized implementation with strict depth limiting.
func (n *Node) VisitPreOrder(fn func(*Node)) {
	if n == nil {
		return
	}
	// Use the channel-based optimized version and consume it
	for node := range preOrder(n) {
		fn(node)
	}
}

// PreOrder returns a channel streaming nodes in pre-order traversal.
// Now uses the final optimized implementation with strict depth limiting.
func (n *Node) PreOrder() <-chan *Node {
	return preOrder(n)
}

// VisitPostOrder visits all nodes in post-order (children left-to-right, then root).
// Now uses the final optimized implementation with strict depth limiting.
func (n *Node) VisitPostOrder(fn func(*Node)) {
	postOrder(n, fn)
}

// Ancestors returns the list of ancestors from root to the parent of target (empty if not found).
// Returns nil if n or target is nil.
func (n *Node) Ancestors(target *Node) []*Node {
	if isNodeNil(n) || isNodeNil(target) {
		return nil
	}
	return findAncestors(n, target)
}

func findAncestors(n, target *Node) []*Node {
	stack := []nodeAncestorFrame{{n, nil}}
	for nodeHasAncestorStack(stack) {
		top := nodePopAncestorStack(&stack)
		if isTargetFound(top.node, target) {
			return top.parent
		}
		nodePushChildAncestors(top, &stack)
	}
	return nil
}

func nodeHasAncestorStack(stack []nodeAncestorFrame) bool {
	return len(stack) > 0
}

func nodePopAncestorStack(stack *[]nodeAncestorFrame) nodeAncestorFrame {
	last := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return last
}

func isTargetFound(node, target *Node) bool {
	return node == target
}

func nodePushChildAncestors(top nodeAncestorFrame, stack *[]nodeAncestorFrame) {
	for i := len(top.node.Children) - 1; i >= 0; i-- {
		child := top.node.Children[i]
		anc := append([]*Node{}, top.parent...)
		anc = append(anc, top.node)
		*stack = append(*stack, nodeAncestorFrame{child, anc})
	}
}

// Transform returns a new tree where each node is replaced by the result of fn(node) (post-order, non-recursive).
// The returned tree is a deep copy with transformations applied. Returns nil if n is nil.
func (n *Node) Transform(fn func(*Node) *Node) *Node {
	if isNodeNil(n) {
		return nil
	}
	return transformNode(n, fn)
}

func transformNode(n *Node, fn func(*Node) *Node) *Node {
	var (
		stack   []nodeTransformFrame
		results = make(map[*Node]*Node)
	)
	stack = append(stack, nodeTransformFrame{n, nil, 0, nil})
	for nodeHasTransformStack(stack) {
		top := &stack[len(stack)-1]
		if nodeHasMoreChildren(top) {
			nodePushChildTransform(top, &stack)
			nodeIncrementChildIndex(top)
			continue
		}
		nodeProcessTransformedNode(top, results, fn)
		nodePopTransformStack(&stack)
	}
	return results[n]
}

func nodeHasTransformStack(stack []nodeTransformFrame) bool {
	return len(stack) > 0
}

func nodeHasMoreChildren(top *nodeTransformFrame) bool {
	return top.childIdx < len(top.node.Children)
}

func nodePushChildTransform(top *nodeTransformFrame, stack *[]nodeTransformFrame) {
	child := top.node.Children[top.childIdx]
	*stack = append(*stack, nodeTransformFrame{child, top.node, 0, nil})
}

func nodeIncrementChildIndex(top *nodeTransformFrame) {
	top.childIdx++
}

func nodeProcessTransformedNode(top *nodeTransformFrame, results map[*Node]*Node, fn func(*Node) *Node) {
	copy := *top.node
	copy.Children = make([]*Node, len(top.node.Children))
	for i, c := range top.node.Children {
		copy.Children[i] = results[c]
	}
	results[top.node] = fn(&copy)
}

func nodePopTransformStack(stack *[]nodeTransformFrame) {
	*stack = (*stack)[:len(*stack)-1]
}

// ReplaceChild replaces the first occurrence of old in Children with new. Returns true if replaced.
func (n *Node) ReplaceChild(old, new *Node) bool {
	for i, c := range n.Children {
		if isChildMatch(c, old) {
			replaceChildAtIndex(n, i, new)
			return true
		}
	}
	return false
}

func replaceChildAtIndex(n *Node, index int, new *Node) {
	n.Children[index] = new
}

// FindDSL queries nodes using a DSL string.
// Example:
//
//	nodes, err := node.FindDSL("type == 'Function' | map(.children)")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, n := range nodes {
//	    fmt.Println(n.Type)
//	}
func (n *Node) FindDSL(query string) ([]*Node, error) {
	if len(query) == 0 {
		return nil, fmt.Errorf("query string is empty")
	}
	ast, err := ParseDSL(query)
	if err != nil {
		return nil, fmt.Errorf("DSL parse error: %w", err)
	}
	// If the top-level AST is a FilterNode, run over n.Children
	if filter, ok := ast.(*FilterNode); ok {
		runtime, err := LowerDSL(filter)
		if err != nil {
			return nil, fmt.Errorf("DSL lowering error: %w", err)
		}
		result := runtime(n.Children)
		if result == nil {
			return []*Node{}, nil
		}
		return result, nil
	}
	// If the top-level AST is a PipelineNode, use n.Children as input
	if _, ok := ast.(*PipelineNode); ok {
		runtime, err := LowerDSL(ast)
		if err != nil {
			return nil, fmt.Errorf("DSL lowering error: %w", err)
		}
		result := runtime(n.Children)
		if result == nil {
			return []*Node{}, nil
		}
		return result, nil
	}
	runtime, err := LowerDSL(ast)
	if err != nil {
		return nil, fmt.Errorf("DSL lowering error: %w", err)
	}
	result := runtime([]*Node{n})
	if result == nil {
		return []*Node{}, nil
	}
	return result, nil
}

// HasRole checks if the node has the given role.
// Example:
//
//	if uast.HasRole(node, uast.RoleFunction) {
//	    fmt.Println("Node is a function")
//	}
func HasRole(node *Node, role Role) bool {
	if isNodeNil(node) || hasNoRoles(node) {
		return false
	}
	return hasRoleInList(node.Roles, role)
}

func hasNoRoles(node *Node) bool {
	return len(node.Roles) == 0
}

func hasRoleInList(roles []Role, target Role) bool {
	for _, r := range roles {
		if isRoleMatch(r, target) {
			return true
		}
	}
	return false
}

func isRoleMatch(role, target Role) bool {
	return role == target
}

// Transform mutates the tree in place using the provided function.
// Example:
//
//	uast.Transform(root, func(n *uast.Node) bool {
//	    if n.Type == "Comment" {
//	        n.Token = ""
//	    }
//	    return true // continue traversal
//	})
func Transform(root *Node, fn func(*Node) bool) {
	if isNodeNil(root) {
		return
	}
	transformInPlace(root, fn)
}

func transformInPlace(root *Node, fn func(*Node) bool) {
	stack := []*Node{root}
	for nodeHasStack(stack) {
		n := nodePopStack(&stack)
		if shouldContinueTransform(n, fn) {
			nodePushChildrenToStack(n, &stack)
		}
	}
}

func shouldContinueTransform(n *Node, fn func(*Node) bool) bool {
	return fn(n)
}

// Final optimized tree traversal with strict depth limiting
func preOrder(node *Node) <-chan *Node {
	ch := make(chan *Node)
	go func() {
		defer close(ch)
		if node == nil {
			return
		}

		maxAllowedDepth := 25
		stack := make([]*Node, 0, 64)
		stack = append(stack, node)

		for len(stack) > 0 {
			n := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if n == nil {
				continue
			}

			ch <- n
			childCount := len(n.Children)

			if childCount > 0 {
				if len(stack) >= maxAllowedDepth {
					processRemainingNodesDepthLimited(n, ch, 0, 10)
					continue
				}

				if cap(stack) < len(stack)+childCount {
					newStack := make([]*Node, len(stack), len(stack)+childCount+32)
					copy(newStack, stack)
					stack = newStack
				}

				for i := childCount - 1; i >= 0; i-- {
					stack = append(stack, n.Children[i])
				}
			}
		}
	}()
	return ch
}

// Process remaining nodes with depth-limited recursion
func processRemainingNodesDepthLimited(node *Node, ch chan<- *Node, depth, maxDepth int) {
	if depth >= maxDepth {
		processRemainingNodesIterative(node, ch)
		return
	}

	ch <- node
	for _, child := range node.Children {
		processRemainingNodesDepthLimited(child, ch, depth+1, maxDepth)
	}
}

// Process remaining nodes iteratively
func processRemainingNodesIterative(node *Node, ch chan<- *Node) {
	queue := make([]*Node, 0, 32)
	queue = append(queue, node)

	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]

		if n == nil {
			continue
		}

		ch <- n

		for _, child := range n.Children {
			queue = append(queue, child)
		}
	}
}

// Final optimized post-order traversal with strict depth limiting
func postOrder(node *Node, fn func(*Node)) {
	if node == nil {
		return
	}

	maxAllowedDepth := 25
	stack := make([]struct {
		node  *Node
		index int
	}, 0, 64)

	stack = append(stack, struct {
		node  *Node
		index int
	}{node, 0})

	for len(stack) > 0 {
		if len(stack) >= maxAllowedDepth {
			processRemainingNodesPostOrderDepthLimited(node, fn, 0, 10)
			break
		}

		top := &stack[len(stack)-1]

		if top.index == 0 {
			childCount := len(top.node.Children)
			if childCount > 0 {
				if cap(stack) < len(stack)+childCount {
					newStack := make([]struct {
						node  *Node
						index int
					}, len(stack), len(stack)+childCount+32)
					copy(newStack, stack)
					stack = newStack
				}

				for i := childCount - 1; i >= 0; i-- {
					stack = append(stack, struct {
						node  *Node
						index int
					}{top.node.Children[i], 0})
				}
				top.index = 1
			} else {
				fn(top.node)
				stack = stack[:len(stack)-1]
			}
		} else {
			fn(top.node)
			stack = stack[:len(stack)-1]
		}
	}
}

// Process remaining nodes for post-order with depth limiting
func processRemainingNodesPostOrderDepthLimited(node *Node, fn func(*Node), depth, maxDepth int) {
	if depth >= maxDepth {
		// Switch to iterative approach
		processRemainingNodesPostOrderIterative(node, fn)
		return
	}

	for _, child := range node.Children {
		processRemainingNodesPostOrderDepthLimited(child, fn, depth+1, maxDepth)
	}
	fn(node)
}

// Process remaining nodes for post-order iteratively
func processRemainingNodesPostOrderIterative(node *Node, fn func(*Node)) {
	stack := make([]*Node, 0, 32)
	visited := make(map[*Node]bool)
	stack = append(stack, node)

	for len(stack) > 0 {
		n := stack[len(stack)-1]

		if visited[n] {
			fn(n)
			stack = stack[:len(stack)-1]
		} else {
			visited[n] = true

			for i := len(n.Children) - 1; i >= 0; i-- {
				stack = append(stack, n.Children[i])
			}
		}
	}
}
