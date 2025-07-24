package node

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
type FieldNode struct {
	Fields []string // Support nested field access like .props.name
}

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

func LowerDSL(ast DSLNode) (QueryFunc, error) {
	switch {
	case isPipelineNode(ast):
		return lowerPipeline(ast.(*PipelineNode))
	case isMapNode(ast):
		return lowerMap(ast.(*MapNode))
	case isFilterNode(ast):
		return lowerFilter(ast.(*FilterNode))
	case isReduceNode(ast):
		return lowerReduce(ast.(*ReduceNode))
	case isFieldNode(ast):
		return lowerField(ast.(*FieldNode))
	case isLiteralNode(ast):
		return lowerLiteral(ast.(*LiteralNode))
	case isCallNode(ast):
		return lowerCall(ast.(*CallNode))
	case isRMapNode(ast):
		return lowerRMap(ast.(*RMapNode))
	case isRFilterNode(ast):
		return lowerRFilter(ast.(*RFilterNode))
	default:
		return nil, fmt.Errorf("unsupported DSL node type: %T", ast)
	}
}

func isPipelineNode(n DSLNode) bool { return isType[*PipelineNode](n) }
func isMapNode(n DSLNode) bool      { return isType[*MapNode](n) }
func isFilterNode(n DSLNode) bool   { return isType[*FilterNode](n) }
func isReduceNode(n DSLNode) bool   { return isType[*ReduceNode](n) }
func isFieldNode(n DSLNode) bool    { return isType[*FieldNode](n) }
func isLiteralNode(n DSLNode) bool  { return isType[*LiteralNode](n) }
func isCallNode(n DSLNode) bool     { return isType[*CallNode](n) }
func isRMapNode(n DSLNode) bool     { return isType[*RMapNode](n) }
func isRFilterNode(n DSLNode) bool  { return isType[*RFilterNode](n) }

func isType[T any](n DSLNode) bool {
	_, ok := n.(T)
	return ok
}

func lowerPipeline(n *PipelineNode) (QueryFunc, error) {
	if isEmptyStages(n.Stages) {
		return nil, fmt.Errorf("empty pipeline")
	}
	funcs, err := buildPipelineFuncs(n.Stages)
	if hasError(err) {
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
		if hasError(err) {
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
			results = f(results)
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
	if hasError(err) {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runMap(exprFunc, nodes)
	}, nil
}

func runMap(exprFunc QueryFunc, nodes []*Node) []*Node {
	if isExactChildrenField(exprFunc) {
		return flattenChildren(nodes)
	}
	return mapOverNodes(exprFunc, nodes)
}

// Only returns true if exprFunc is exactly a .children field access
func isExactChildrenField(exprFunc QueryFunc) bool {
	testNode := &Node{Type: "test", Children: []*Node{{Type: "c1"}, {Type: "c2"}}}
	res := exprFunc([]*Node{testNode})
	if len(res) != len(testNode.Children) {
		return false
	}
	for i := range res {
		if res[i] != testNode.Children[i] {
			return false
		}
	}
	return true
}

func flattenChildren(nodes []*Node) []*Node {
	totalChildren := calculateTotalChildren(nodes)
	out := make([]*Node, 0, totalChildren)
	for _, node := range nodes {
		out = append(out, node.Children...)
	}
	return out
}

func mapOverChildren(exprFunc QueryFunc, node *Node) []*Node {
	childCount := len(node.Children)
	out := make([]*Node, 0, childCount)
	for _, child := range node.Children {
		out = append(out, exprFunc([]*Node{child})...)
	}
	return out
}

func mapOverNodes(exprFunc QueryFunc, nodes []*Node) []*Node {
	out := make([]*Node, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, exprFunc([]*Node{node})...)
	}
	return out
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

func lowerRMap(n *RMapNode) (QueryFunc, error) {
	exprFunc, err := LowerDSL(n.Expr)
	if hasError(err) {
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

func estimateRMapResultSize(nodes []*Node) int {
	total := 0
	for _, node := range nodes {
		total += countNodesInTree(node)
	}
	return total
}

func countNodesInTree(node *Node) int {
	if isNilNode(node) {
		return 0
	}
	count := 1
	for _, child := range node.Children {
		count += countNodesInTree(child)
	}
	return count
}

func isNilNode(node *Node) bool {
	return node == nil
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
	if hasError(err) {
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
	if hasError(err) {
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
	return []*Node{{Type: "Literal", Token: fmt.Sprint(len(nodes))}}
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
		if isEmptyFields(n.Fields) {
			continue
		}
		if isSingleField(n.Fields) {
			fieldName := n.Fields[0]
			if isChildrenFieldName(fieldName) {
				out = append(out, node.Children...)
			} else if isTokenFieldName(fieldName) {
				out = append(out, NewLiteralNode(node.Token))
			} else if isIdFieldName(fieldName) {
				out = append(out, NewLiteralNode(fmt.Sprint(node.Id)))
			} else if isRolesField(fieldName) {
				for _, r := range node.Roles {
					out = append(out, NewLiteralNode(string(r)))
				}
			} else if isTypeField(fieldName, node) {
				out = append(out, NewLiteralNode(node.Type))
			} else if hasProp(node, fieldName) {
				out = append(out, NewLiteralNode(node.Props[fieldName]))
			}
		} else {
			value := getNestedFieldValue(node, n.Fields)
			if value != nil {
				out = append(out, NewLiteralNode(fmt.Sprint(value)))
			}
		}
	}
	return out
}

func isEmptyFields(fields []string) bool {
	return len(fields) == 0
}

func isSingleField(fields []string) bool {
	return len(fields) == 1
}

func isChildrenFieldName(fieldName string) bool {
	return fieldName == "children"
}

func isTokenFieldName(fieldName string) bool {
	return fieldName == "token"
}

func isIdFieldName(fieldName string) bool {
	return fieldName == "id"
}

func getNestedFieldValue(node *Node, fields []string) interface{} {
	if isEmptyFields(fields) {
		return nil
	}
	current := node
	for i, field := range fields {
		if isFirstField(i) {
			if isTypeField(field, current) {
				return current.Type
			} else if isTokenFieldName(field) {
				return current.Token
			} else if isIdFieldName(field) {
				return current.Id
			} else if isRolesField(field) {
				if !isLastField(i, fields) {
					return nil
				}
				return current.Roles
			} else if isChildrenFieldName(field) {
				if !isLastField(i, fields) {
					return nil
				}
				return current.Children
			} else if field == "props" {
				// Special handling for props - return the props map for nested access
				if isLastField(i, fields) {
					return current.Props
				}
				// Continue with props map for nested access
				current = &Node{Props: current.Props}
			} else if hasProp(current, field) {
				propValue := current.Props[field]
				if isLastField(i, fields) {
					return propValue
				}
				// For nested access, we need to continue with the prop value
				// But since props are strings, we can't continue nesting
				return nil
			} else {
				return nil
			}
		} else {
			// For non-first fields, we're trying to access properties of the previous field's value
			if hasProp(current, field) {
				propValue := current.Props[field]
				if isLastField(i, fields) {
					return propValue
				}
				// Continue with the prop value for further nesting
				current = &Node{Props: map[string]string{field: propValue}}
			} else {
				return nil
			}
		}
	}
	return nil
}

func isFirstField(i int) bool {
	return i == 0
}

func isLastField(i int, fields []string) bool {
	return i == len(fields)-1
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
	if isNotEqual(n) {
		return lowerNotEqual(n)
	}
	if isNot(n) {
		return lowerNot(n)
	}
	if isGreaterThan(n) {
		return lowerGreaterThan(n)
	}
	if isGreaterThanOrEqual(n) {
		return lowerGreaterThanOrEqual(n)
	}
	if isLessThan(n) {
		return lowerLessThan(n)
	}
	if isLessThanOrEqual(n) {
		return lowerLessThanOrEqual(n)
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
	if hasError(err) {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if hasError(err) {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runLogicalOr(leftFunc, rightFunc, nodes)
	}, nil
}

func runLogicalOr(leftFunc, rightFunc QueryFunc, nodes []*Node) []*Node {
	var out []*Node
	for _, node := range nodes {
		if isLogicalOrTrue(leftFunc, rightFunc, node) {
			out = append(out, NewLiteralNode("true"))
		} else {
			out = append(out, NewLiteralNode("false"))
		}
	}
	return out
}

func isLogicalOrTrue(leftFunc, rightFunc QueryFunc, node *Node) bool {
	l := leftFunc([]*Node{node})
	r := rightFunc([]*Node{node})
	return isTrue(l) || isTrue(r)
}

func isLogicalAnd(n *CallNode) bool {
	return n.Name == "&&" && len(n.Args) == 2
}

func lowerLogicalAnd(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if hasError(err) {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if hasError(err) {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runLogicalAnd(leftFunc, rightFunc, nodes)
	}, nil
}

func runLogicalAnd(leftFunc, rightFunc QueryFunc, nodes []*Node) []*Node {
	var out []*Node
	for _, node := range nodes {
		if isLogicalAndTrue(leftFunc, rightFunc, node) {
			out = append(out, NewLiteralNode("true"))
		} else {
			out = append(out, NewLiteralNode("false"))
		}
	}
	return out
}

func isLogicalAndTrue(leftFunc, rightFunc QueryFunc, node *Node) bool {
	l := leftFunc([]*Node{node})
	r := rightFunc([]*Node{node})
	return isTrue(l) && isTrue(r)
}

func isTrue(nodes []*Node) bool {
	return len(nodes) > 0 && nodes[0].Type == "Literal" && nodes[0].Token == "true"
}

func isEquality(n *CallNode) bool {
	return n.Name == "==" && len(n.Args) == 2
}

func lowerEquality(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if hasError(err) {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if hasError(err) {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		return runEquality(leftFunc, rightFunc, nodes)
	}, nil
}

func runEquality(leftFunc, rightFunc QueryFunc, nodes []*Node) []*Node {
	var out []*Node
	for _, node := range nodes {
		if isEqual(leftFunc([]*Node{node}), rightFunc([]*Node{node})) {
			out = append(out, NewLiteralNode("true"))
		} else {
			out = append(out, NewLiteralNode("false"))
		}
	}
	return out
}

func isEqual(left, right []*Node) bool {
	return len(left) > 0 && len(right) > 0 && left[0].Token == right[0].Token
}

func isMembership(n *CallNode) bool {
	return n.Name == "has" && len(n.Args) == 2
}

func lowerMembership(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if hasError(err) {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if hasError(err) {
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

func hasError(err error) bool {
	return err != nil
}

// --- Operator helpers and lowering implementations ---

func isNotEqual(n *CallNode) bool {
	return n.Name == "!=" && len(n.Args) == 2
}

func lowerNotEqual(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if err != nil {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		var out []*Node
		for _, node := range nodes {
			l := leftFunc([]*Node{node})
			r := rightFunc([]*Node{node})
			if !tokensEqual(l, r) {
				out = append(out, NewLiteralNode("true"))
			} else {
				out = append(out, NewLiteralNode("false"))
			}
		}
		return out
	}, nil
}

func isNot(n *CallNode) bool {
	return n.Name == "!" && len(n.Args) == 1
}

func lowerNot(n *CallNode) (QueryFunc, error) {
	argFunc, err := LowerDSL(n.Args[0])
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		var out []*Node
		for _, node := range nodes {
			res := argFunc([]*Node{node})
			if isTrue(res) {
				out = append(out, NewLiteralNode("false"))
			} else {
				out = append(out, NewLiteralNode("true"))
			}
		}
		return out
	}, nil
}

func isGreaterThan(n *CallNode) bool {
	return n.Name == ">" && len(n.Args) == 2
}

func lowerGreaterThan(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if err != nil {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		var out []*Node
		for _, node := range nodes {
			l := leftFunc([]*Node{node})
			r := rightFunc([]*Node{node})
			if tokensCompare(l, r, ">") {
				out = append(out, NewLiteralNode("true"))
			} else {
				out = append(out, NewLiteralNode("false"))
			}
		}
		return out
	}, nil
}

func isGreaterThanOrEqual(n *CallNode) bool {
	return n.Name == ">=" && len(n.Args) == 2
}

func lowerGreaterThanOrEqual(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if err != nil {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		var out []*Node
		for _, node := range nodes {
			l := leftFunc([]*Node{node})
			r := rightFunc([]*Node{node})
			if tokensCompare(l, r, ">=") {
				out = append(out, NewLiteralNode("true"))
			} else {
				out = append(out, NewLiteralNode("false"))
			}
		}
		return out
	}, nil
}

func isLessThan(n *CallNode) bool {
	return n.Name == "<" && len(n.Args) == 2
}

func lowerLessThan(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if err != nil {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		var out []*Node
		for _, node := range nodes {
			l := leftFunc([]*Node{node})
			r := rightFunc([]*Node{node})
			if tokensCompare(l, r, "<") {
				out = append(out, NewLiteralNode("true"))
			} else {
				out = append(out, NewLiteralNode("false"))
			}
		}
		return out
	}, nil
}

func isLessThanOrEqual(n *CallNode) bool {
	return n.Name == "<=" && len(n.Args) == 2
}

func lowerLessThanOrEqual(n *CallNode) (QueryFunc, error) {
	leftFunc, err := LowerDSL(n.Args[0])
	if err != nil {
		return nil, err
	}
	rightFunc, err := LowerDSL(n.Args[1])
	if err != nil {
		return nil, err
	}
	return func(nodes []*Node) []*Node {
		var out []*Node
		for _, node := range nodes {
			l := leftFunc([]*Node{node})
			r := rightFunc([]*Node{node})
			if tokensCompare(l, r, "<=") {
				out = append(out, NewLiteralNode("true"))
			} else {
				out = append(out, NewLiteralNode("false"))
			}
		}
		return out
	}, nil
}

// --- Utility functions for comparison ---

func tokensEqual(left, right []*Node) bool {
	if len(left) == 0 || len(right) == 0 {
		return false
	}
	return left[0].Token == right[0].Token
}

func tokensCompare(left, right []*Node, op string) bool {
	if len(left) == 0 || len(right) == 0 {
		return false
	}
	l, r := left[0].Token, right[0].Token
	lf, lerr := parseFloat(l)
	rf, rerr := parseFloat(r)
	if lerr == nil && rerr == nil {
		switch op {
		case ">":
			return lf > rf
		case ">=":
			return lf >= rf
		case "<":
			return lf < rf
		case "<=":
			return lf <= rf
		}
	}
	// fallback to string comparison
	switch op {
	case ">":
		return l > r
	case ">=":
		return l >= r
	case "<":
		return l < r
	case "<=":
		return l <= r
	}
	return false
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
