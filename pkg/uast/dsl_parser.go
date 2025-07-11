package uast

import (
	"fmt"
)

// DSLNode is the interface for all nodes in the UAST DSL AST.
type DSLNode any

// MapNode represents a map operation in the DSL.
type MapNode struct{ Expr DSLNode }

// FilterNode represents a filter operation in the DSL.
type FilterNode struct{ Expr DSLNode }

// ReduceNode represents a reduce operation in the DSL.
type ReduceNode struct{ Expr DSLNode }

// FieldNode represents a field/property access in the DSL.
type FieldNode struct{ Name string }

// LiteralNode represents a literal value in the DSL.
type LiteralNode struct{ Value any }

// CallNode represents a function/operator call in the DSL.
type CallNode struct {
	Name string
	Args []DSLNode
}

// PipelineNode represents a pipeline of DSL operations.
type PipelineNode struct{ Stages []DSLNode }

// RMapNode represents a recursive map operation in the DSL.
type RMapNode struct{ Expr DSLNode }

// RFilterNode represents a recursive filter operation in the DSL.
type RFilterNode struct{ Expr DSLNode }

// QueryFunc is a compiled DSL query function that takes a slice of nodes and returns a slice of nodes.
type QueryFunc func([]*Node) []*Node

// LowerDSL compiles a DSL AST to a QueryFunc (Go closure).
// Returns an error if the AST is invalid or unsupported.
func LowerDSL(ast DSLNode) (QueryFunc, error) {
	switch n := ast.(type) {
	case *PipelineNode:
		return lowerPipeline(n)
	case *MapNode:
		return lowerMap(n)
	case *FilterNode:
		return lowerFilter(n)
	case *ReduceNode:
		return lowerReduce(n)
	case *FieldNode:
		return lowerField(n)
	case *LiteralNode:
		return lowerLiteral(n)
	case *CallNode:
		return lowerCall(n)
	case *RMapNode:
		return lowerRMap(n)
	case *RFilterNode:
		return lowerRFilter(n)
	default:
		return nil, fmt.Errorf("unsupported DSL node type: %T", n)
	}
}

func lowerPipeline(n *PipelineNode) (QueryFunc, error) {
	if isEmptyStages(n.Stages) {
		return nil, fmt.Errorf("empty pipeline")
	}
	funcs, err := buildPipelineFuncs(n.Stages)
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runPipelineFuncs(funcs, n.Stages, nodes)
	}, nil
}

func isEmptyStages(stages []DSLNode) bool {
	return len(stages) == 0
}

func buildPipelineFuncs(stages []DSLNode) ([]QueryFunc, error) {
	funcs := make([]QueryFunc, len(stages))
	for i, stage := range stages {
		f, err := LowerDSL(stage)
		if err != nil {
			return nil, err
		}
		funcs[i] = f
	}
	return funcs, nil
}

func runPipelineFuncs(funcs []QueryFunc, stages []DSLNode, nodes []*Node) []*Node {
	results := nodes
	for i, f := range funcs {
		if isLastReduceStage(i, funcs, stages) {
			results = f([]*Node{&Node{Children: results}})
		} else {
			results = f(results)
		}
	}
	return results
}

func isLastReduceStage(i int, funcs []QueryFunc, stages []DSLNode) bool {
	return i == len(funcs)-1 && isReduceStage(stages[i])
}

func isReduceStage(stage DSLNode) bool {
	_, ok := stage.(*ReduceNode)
	return ok
}

func lowerMap(n *MapNode) (QueryFunc, error) {
	exprFunc, err := LowerDSL(n.Expr)
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runMap(exprFunc, nodes)
	}, nil
}

func runMap(exprFunc QueryFunc, nodes []*Node) []*Node {
	// Detect if exprFunc is a field access for 'children'
	isChildrenField := false
	// Hack: create a test node and see if exprFunc returns its children
	testNode := &Node{Type: "test", Children: []*Node{{Type: "c1"}, {Type: "c2"}}}
	res := exprFunc([]*Node{testNode})
	if len(res) == 2 && res[0] == testNode.Children[0] && res[1] == testNode.Children[1] {
		isChildrenField = true
	}

	if isChildrenField {
		totalChildren := calculateTotalChildren(nodes)
		out := make([]*Node, 0, totalChildren)
		for _, node := range nodes {
			out = append(out, node.Children...)
		}
		return out
	}

	if isSingleNodeWithChildren(nodes) {
		childCount := len(nodes[0].Children)
		out := make([]*Node, 0, childCount)
		for _, child := range nodes[0].Children {
			out = append(out, exprFunc([]*Node{child})...)
		}
		return out
	} else {
		out := make([]*Node, 0, len(nodes))
		for _, node := range nodes {
			out = append(out, exprFunc([]*Node{node})...)
		}
		return out
	}
}

func isSingleNodeWithChildren(nodes []*Node) bool {
	return len(nodes) == 1 && len(nodes[0].Children) > 0
}

func calculateTotalChildren(nodes []*Node) int {
	total := 0
	for _, node := range nodes {
		total += len(node.Children)
	}
	return total
}

func estimateRMapResultSize(nodes []*Node) int {
	// Estimate based on total nodes in the tree
	total := 0
	for _, node := range nodes {
		total += countNodesInTree(node)
	}
	return total
}

func countNodesInTree(node *Node) int {
	if node == nil {
		return 0
	}
	count := 1
	for _, child := range node.Children {
		count += countNodesInTree(child)
	}
	return count
}

func lowerRMap(n *RMapNode) (QueryFunc, error) {
	exprFunc, err := LowerDSL(n.Expr)
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runRMap(exprFunc, nodes)
	}, nil
}

func runRMap(exprFunc QueryFunc, nodes []*Node) []*Node {
	estimatedSize := estimateRMapResultSize(nodes)
	out := make([]*Node, 0, estimatedSize)
	for _, node := range nodes {
		stack := []*Node{node}
		for hasStack(stack) {
			curr := popStack(&stack)
			out = append(out, exprFunc([]*Node{curr})...)
			pushChildrenToStack(curr, &stack)
		}
	}
	return out
}

func hasStack(stack []*Node) bool {
	return len(stack) > 0
}

func popStack(stack *[]*Node) *Node {
	last := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return last
}

func pushChildrenToStack(node *Node, stack *[]*Node) {
	for i := len(node.Children) - 1; i >= 0; i-- {
		*stack = append(*stack, node.Children[i])
	}
}

func lowerFilter(n *FilterNode) (QueryFunc, error) {
	predFunc, err := LowerDSL(n.Expr)
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runFilter(predFunc, nodes)
	}, nil
}

func runFilter(predFunc QueryFunc, nodes []*Node) []*Node {
	out := make([]*Node, 0, len(nodes))
	for _, node := range nodes {
		if isPredicateTrue(predFunc, node) {
			out = append(out, node)
		}
	}
	return out
}

func isPredicateTrue(predFunc QueryFunc, node *Node) bool {
	res := predFunc([]*Node{node})
	return len(res) > 0 && res[0].Type == "Literal" && res[0].Token == "true"
}

func lowerRFilter(n *RFilterNode) (QueryFunc, error) {
	predFunc, err := LowerDSL(n.Expr)
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runRFilter(predFunc, nodes)
	}, nil
}

func runRFilter(predFunc QueryFunc, nodes []*Node) []*Node {
	estimatedSize := estimateRMapResultSize(nodes)
	out := make([]*Node, 0, estimatedSize)
	for _, node := range nodes {
		stack := []*Node{node}
		for hasStack(stack) {
			curr := popStack(&stack)
			if isPredicateTrue(predFunc, curr) {
				out = append(out, curr)
			}
			pushChildrenToStack(curr, &stack)
		}
	}
	return out
}

func lowerReduce(n *ReduceNode) (QueryFunc, error) {
	if !isReduceCountCall(n.Expr) {
		return nil, fmt.Errorf("only 'reduce(count)' is supported")
	}
	return func(nodes []*Node) []*Node {
		return runReduce(nodes)
	}, nil
}

func isReduceCountCall(expr DSLNode) bool {
	call, ok := expr.(*CallNode)
	return ok && call.Name == "count"
}

func runReduce(nodes []*Node) []*Node {
	if isEmptyNodes(nodes) {
		return []*Node{{Type: "Literal", Token: "0"}}
	}
	return []*Node{{Type: "Literal", Token: fmt.Sprint(len(nodes[0].Children))}}
}

func isEmptyNodes(nodes []*Node) bool {
	return len(nodes) == 0
}

func lowerField(n *FieldNode) (QueryFunc, error) {
	return func(nodes []*Node) []*Node {
		return runField(n, nodes)
	}, nil
}

func runField(n *FieldNode, nodes []*Node) []*Node {
	var out []*Node
	for _, node := range nodes {
		if n.Name == "children" {
			out = append(out, node.Children...)
		} else if n.Name == "token" {
			out = append(out, NewLiteralNode(node.Token))
		} else if n.Name == "id" {
			out = append(out, NewLiteralNode(fmt.Sprint(node.Id)))
		} else if isRolesField(n.Name) {
			for _, r := range node.Roles {
				out = append(out, NewLiteralNode(string(r)))
			}
		} else if hasProp(node, n.Name) {
			out = append(out, NewLiteralNode(node.Props[n.Name]))
		} else if isTypeField(n.Name, node) {
			out = append(out, NewLiteralNode(node.Type))
		}
	}
	return out
}

func isRolesField(name string) bool {
	return name == "roles"
}

func hasProp(node *Node, name string) bool {
	_, ok := node.Props[name]
	return ok
}

func isTypeField(name string, node *Node) bool {
	return name == "type" && node.Type != ""
}

func lowerLiteral(n *LiteralNode) (QueryFunc, error) {
	return func(nodes []*Node) []*Node {
		return []*Node{NewLiteralNode(fmt.Sprint(n.Value))}
	}, nil
}

func lowerCall(n *CallNode) (QueryFunc, error) {
	if isLogicalOr(n) {
		return lowerLogicalOr(n)
	}
	if isLogicalAnd(n) {
		return lowerLogicalAnd(n)
	}
	if isEquality(n) {
		return lowerEquality(n)
	}
	if isMembership(n) {
		return lowerMembership(n)
	}
	return nil, fmt.Errorf("unsupported call: %s", n.Name)
}

func isLogicalOr(n *CallNode) bool {
	return n.Name == "||" && len(n.Args) == 2
}

func lowerLogicalOr(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if err != nil {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runLogicalOr(leftFunc, rightFunc, nodes)
	}, nil
}

func runLogicalOr(leftFunc, rightFunc QueryFunc, nodes []*Node) []*Node {
	var out []*Node
	for _, node := range nodes {
		l := leftFunc([]*Node{node})
		r := rightFunc([]*Node{node})
		if isTrue(l) || isTrue(r) {
			out = append(out, NewLiteralNode("true"))
		} else {
			out = append(out, NewLiteralNode("false"))
		}
	}
	return out
}

func isLogicalAnd(n *CallNode) bool {
	return n.Name == "&&" && len(n.Args) == 2
}

func lowerLogicalAnd(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if err != nil {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runLogicalAnd(leftFunc, rightFunc, nodes)
	}, nil
}

func runLogicalAnd(leftFunc, rightFunc QueryFunc, nodes []*Node) []*Node {
	var out []*Node
	for _, node := range nodes {
		l := leftFunc([]*Node{node})
		r := rightFunc([]*Node{node})
		if isTrue(l) && isTrue(r) {
			out = append(out, NewLiteralNode("true"))
		} else {
			out = append(out, NewLiteralNode("false"))
		}
	}
	return out
}

func isTrue(nodes []*Node) bool {
	return len(nodes) > 0 && nodes[0].Type == "Literal" && nodes[0].Token == "true"
}

func isEquality(n *CallNode) bool {
	return n.Name == "==" && len(n.Args) == 2
}

func lowerEquality(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if err != nil {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runEquality(leftFunc, rightFunc, nodes)
	}, nil
}

func runEquality(leftFunc, rightFunc QueryFunc, nodes []*Node) []*Node {
	var out []*Node
	for _, node := range nodes {
		left := leftFunc([]*Node{node})
		right := rightFunc([]*Node{node})
		if isEqual(left, right) {
			out = append(out, NewLiteralNode("true"))
		} else {
			out = append(out, NewLiteralNode("false"))
		}
	}
	return out
}

func isEqual(left, right []*Node) bool {
	if len(left) > 0 && len(right) > 0 && left[0].Token == right[0].Token {
		return true
	}
	return false
}

func isMembership(n *CallNode) bool {
	return n.Name == "has" && len(n.Args) == 2
}

func lowerMembership(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if err != nil {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runMembership(leftFunc, rightFunc, nodes)
	}, nil
}

func runMembership(leftFunc, rightFunc QueryFunc, nodes []*Node) []*Node {
	var out []*Node
	for _, node := range nodes {
		leftVals := leftFunc([]*Node{node})
		rightVals := rightFunc(nil)
		if isEmptyNodes(leftVals) || isEmptyNodes(rightVals) {
			out = append(out, NewLiteralNode("false"))
			continue
		}
		rval := rightVals[0].Token
		if containsToken(leftVals, rval) {
			out = append(out, NewLiteralNode("true"))
		} else {
			out = append(out, NewLiteralNode("false"))
		}
	}
	return out
}

func containsToken(nodes []*Node, token string) bool {
	for _, n := range nodes {
		if n.Token == token {
			return true
		}
	}
	return false
}
