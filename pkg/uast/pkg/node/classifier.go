package node

// ClassifyDSLNode classifies DSL nodes by type
func ClassifyDSLNode(node DSLNode) DSLNodeType {
	switch node.(type) {
	case *PipelineNode:
		return PipelineType
	case *MapNode:
		return MapType
	case *FilterNode:
		return FilterType
	case *ReduceNode:
		return ReduceType
	case *FieldNode:
		return FieldType
	case *LiteralNode:
		return LiteralType
	case *CallNode:
		return CallType
	case *RMapNode:
		return RMapType
	case *RFilterNode:
		return RFilterType
	default:
		return ""
	}
}

// DSLNodeClassifier provides a struct-based interface for backward compatibility
type DSLNodeClassifier struct{}

func NewDSLNodeClassifier() *DSLNodeClassifier {
	return &DSLNodeClassifier{}
}

func (c *DSLNodeClassifier) Classify(node DSLNode) DSLNodeType {
	return ClassifyDSLNode(node)
}

// isLiteralNode checks if a node is a literal node
func isLiteralNode(n DSLNode) bool {
	_, ok := n.(*LiteralNode)
	return ok
}
