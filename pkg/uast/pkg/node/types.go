package node

import (
	"fmt"
)

// --- Node Types ---
type DSLNode any

type MapNode struct{ Expr DSLNode }
type FilterNode struct{ Expr DSLNode }
type ReduceNode struct{ Expr DSLNode }
type FieldNode struct{ Fields []string }
type LiteralNode struct{ Value any }
type CallNode struct {
	Name string
	Args []DSLNode
}
type PipelineNode struct{ Stages []DSLNode }
type RMapNode struct{ Expr DSLNode }
type RFilterNode struct{ Expr DSLNode }

type QueryFunc func([]*Node) []*Node

type DSLNodeType string

const (
	PipelineType DSLNodeType = "Pipeline"
	MapType      DSLNodeType = "Map"
	FilterType   DSLNodeType = "Filter"
	ReduceType   DSLNodeType = "Reduce"
	FieldType    DSLNodeType = "Field"
	LiteralType  DSLNodeType = "Literal"
	CallType     DSLNodeType = "Call"
	RMapType     DSLNodeType = "RMap"
	RFilterType  DSLNodeType = "RFilter"
)

// --- Lowering Interfaces ---
type DSLNodeLowerer interface {
	Lower(node DSLNode) (QueryFunc, error)
}

type DSLNodeLowererRegistry struct {
	lowerers map[DSLNodeType]DSLNodeLowerer
}

// --- Field Access Interfaces ---
type FieldAccessStrategy interface {
	Access(node *Node) []*Node
}

type FieldAccessStrategyRegistry struct {
	strategies map[string]FieldAccessStrategy
}

func NewFieldAccessStrategyRegistry() *FieldAccessStrategyRegistry {
	registry := &FieldAccessStrategyRegistry{
		strategies: make(map[string]FieldAccessStrategy),
	}

	// Register strategies
	registry.strategies["children"] = &ChildrenFieldStrategy{}
	registry.strategies["token"] = &TokenFieldStrategy{}
	registry.strategies["id"] = &IdFieldStrategy{}
	registry.strategies["roles"] = &RolesFieldStrategy{}
	registry.strategies["type"] = &TypeFieldStrategy{}
	registry.strategies["first"] = &FirstFieldStrategy{}
	registry.strategies["last"] = &LastFieldStrategy{}

	return registry
}

func (r *FieldAccessStrategyRegistry) Access(node *Node, fieldName string) []*Node {
	strategy, exists := r.strategies[fieldName]
	if exists {
		return strategy.Access(node)
	}

	// Check for props access
	if hasProp(node, fieldName) {
		return []*Node{NewLiteralNode(node.Props[fieldName])}
	}

	return nil
}

func hasProp(node *Node, fieldName string) bool {
	_, exists := node.Props[fieldName]
	return exists
}

type ChildrenFieldStrategy struct{}

func (s *ChildrenFieldStrategy) Access(node *Node) []*Node {
	return node.Children
}

type TokenFieldStrategy struct{}

func (s *TokenFieldStrategy) Access(node *Node) []*Node {
	return []*Node{NewLiteralNode(node.Token)}
}

type IdFieldStrategy struct{}

func (s *IdFieldStrategy) Access(node *Node) []*Node {
	return []*Node{NewLiteralNode(fmt.Sprint(node.Id))}
}

type RolesFieldStrategy struct{}

func (s *RolesFieldStrategy) Access(node *Node) []*Node {
	var out []*Node
	for _, r := range node.Roles {
		out = append(out, NewLiteralNode(string(r)))
	}
	return out
}

type TypeFieldStrategy struct{}

func (s *TypeFieldStrategy) Access(node *Node) []*Node {
	if node.Type != "" {
		return []*Node{NewLiteralNode(string(node.Type))}
	}
	return nil
}

type FirstFieldStrategy struct{}

func (s *FirstFieldStrategy) Access(node *Node) []*Node {
	return getFirstFieldValue(node, "first")
}

type LastFieldStrategy struct{}

func (s *LastFieldStrategy) Access(node *Node) []*Node {
	return getLastFieldValue(node, "last")
}

// Note: Node is expected to be defined elsewhere in the package.
