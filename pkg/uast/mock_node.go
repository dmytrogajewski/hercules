package uast

// MockNode is a test double for Node, now using the canonical struct.
type MockNode struct {
	Node
}

func NewMockNode() *MockNode {
	return &MockNode{Node: Node{}}
}
