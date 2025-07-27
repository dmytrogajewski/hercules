package node

import (
	"fmt"
)

// LowerDSL lowers a DSLNode AST to a QueryFunc
func LowerDSL(ast DSLNode) (QueryFunc, error) {
	return globalLowererRegistry.Lower(ast)
}

var globalLowererRegistry *DSLNodeLowererRegistry

func init() {
	globalLowererRegistry = NewDSLNodeLowererRegistry()
	globalLowererRegistry.Register(PipelineType, &PipelineLowerer{})
	globalLowererRegistry.Register(MapType, &MapLowerer{})
	globalLowererRegistry.Register(FilterType, &FilterLowerer{})
	globalLowererRegistry.Register(ReduceType, &ReduceLowerer{})
	globalLowererRegistry.Register(FieldType, &FieldLowerer{})
	globalLowererRegistry.Register(LiteralType, &LiteralLowerer{})
	globalLowererRegistry.Register(CallType, &CallLowerer{})
	globalLowererRegistry.Register(RMapType, &RMapLowerer{})
	globalLowererRegistry.Register(RFilterType, &RFilterLowerer{})
}

// DSLNodeLowererRegistry manages lowerers for different node types
func NewDSLNodeLowererRegistry() *DSLNodeLowererRegistry {
	return &DSLNodeLowererRegistry{
		lowerers: make(map[DSLNodeType]DSLNodeLowerer),
	}
}

func (r *DSLNodeLowererRegistry) Register(nodeType DSLNodeType, lowerer DSLNodeLowerer) {
	r.lowerers[nodeType] = lowerer
}

func (r *DSLNodeLowererRegistry) Lower(node DSLNode) (QueryFunc, error) {
	nodeType := ClassifyDSLNode(node)
	lowerer, exists := r.lowerers[nodeType]
	if !exists {
		return nil, fmt.Errorf("unsupported DSL node type: %T", node)
	}
	return lowerer.Lower(node)
}

// Lowerer implementations
type PipelineLowerer struct{}

func (l *PipelineLowerer) Lower(node DSLNode) (QueryFunc, error) {
	return lowerPipeline(node.(*PipelineNode))
}

type MapLowerer struct{}

func (l *MapLowerer) Lower(node DSLNode) (QueryFunc, error) { return lowerMap(node.(*MapNode)) }

type FilterLowerer struct{}

func (l *FilterLowerer) Lower(node DSLNode) (QueryFunc, error) {
	return lowerFilter(node.(*FilterNode))
}

type ReduceLowerer struct{}

func (l *ReduceLowerer) Lower(node DSLNode) (QueryFunc, error) {
	return lowerReduce(node.(*ReduceNode))
}

type FieldLowerer struct{}

func (l *FieldLowerer) Lower(node DSLNode) (QueryFunc, error) { return lowerField(node.(*FieldNode)) }

type LiteralLowerer struct{}

func (l *LiteralLowerer) Lower(node DSLNode) (QueryFunc, error) {
	return lowerLiteral(node.(*LiteralNode))
}

type CallLowerer struct{}

func (l *CallLowerer) Lower(node DSLNode) (QueryFunc, error) { return lowerCall(node.(*CallNode)) }

type RMapLowerer struct{}

func (l *RMapLowerer) Lower(node DSLNode) (QueryFunc, error) { return lowerRMap(node.(*RMapNode)) }

type RFilterLowerer struct{}

func (l *RFilterLowerer) Lower(node DSLNode) (QueryFunc, error) {
	return lowerRFilter(node.(*RFilterNode))
}

func lowerCall(n *CallNode) (QueryFunc, error) {
	return globalOperatorRegistry.Handle(n)
}

// Core lowering functions
func lowerField(n *FieldNode) (QueryFunc, error) {
	return func(nodes []*Node) []*Node {
		var out []*Node
		for _, node := range nodes {
			result := processFieldAccess(n, node)
			if result != nil {
				out = append(out, result...)
			}
		}
		return out
	}, nil
}

func lowerLiteral(n *LiteralNode) (QueryFunc, error) {
	return func(nodes []*Node) []*Node {
		return []*Node{NewLiteralNode(fmt.Sprint(n.Value))}
	}, nil
}
