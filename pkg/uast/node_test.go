package uast

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNodeCreationAndFields(t *testing.T) {
	n := &Node{
		Id:       12345,
		Type:     "Function",
		Token:    "main",
		Roles:    []Role{"Identifier", "Declaration"},
		Pos:      &Positions{StartLine: 1, StartCol: 1, StartOffset: 0, EndLine: 1, EndCol: 10, EndOffset: 10},
		Props:    map[string]string{"lang": "go"},
		Children: []*Node{},
	}
	if n.Id != 12345 || n.Type != "Function" || n.Token != "main" {
		t.Errorf("Node fields not set correctly: %+v", n)
	}
	if n.Pos == nil || n.Pos.StartLine != 1 || n.Pos.EndOffset != 10 {
		t.Errorf("Positions not set correctly: %+v", n.Pos)
	}
	if len(n.Roles) != 2 || n.Roles[0] != "Identifier" {
		t.Errorf("Roles not set correctly: %+v", n.Roles)
	}
	if n.Props["lang"] != "go" {
		t.Errorf("Props not set correctly: %+v", n.Props)
	}
}

func TestNodeJSONSerialization(t *testing.T) {
	n := &Node{
		Id:    1,
		Type:  "Literal",
		Token: "42",
		Roles: []Role{"Constant"},
		Pos:   nil,
		Props: map[string]string{},
	}
	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("Failed to marshal Node: %v", err)
	}
	var out Node
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("Failed to unmarshal Node: %v", err)
	}
	if n.Id != out.Id || n.Type != out.Type || n.Token != out.Token || !reflect.DeepEqual(n.Roles, out.Roles) || !reflect.DeepEqual(n.Pos, out.Pos) {
		t.Errorf("Node mismatch after JSON roundtrip: got %+v, want %+v", &out, n)
	}
	if (n.Props == nil && len(out.Props) != 0) || (out.Props == nil && len(n.Props) != 0) || (n.Props != nil && out.Props != nil && !reflect.DeepEqual(n.Props, out.Props)) {
		t.Errorf("Props mismatch after JSON roundtrip: got %+v, want %+v", out.Props, n.Props)
	}
}

func TestNodeAddRemoveChildren(t *testing.T) {
	parent := &Node{Type: "Parent"}
	child1 := &Node{Type: "Child1"}
	child2 := &Node{Type: "Child2"}
	parent.AddChild(child1)
	parent.AddChild(child2)
	if len(parent.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(parent.Children))
	}
	parent.RemoveChild(child1)
	if len(parent.Children) != 1 || parent.Children[0] != child2 {
		t.Errorf("RemoveChild failed: %+v", parent.Children)
	}
}

func TestNodeString(t *testing.T) {
	n := &Node{Type: "Function", Token: "main"}
	str := n.String()
	if str == "" || str == "<nil>" {
		t.Errorf("String() should return non-empty string, got %q", str)
	}
}

func TestNodeEdgeCases(t *testing.T) {
	n := &Node{}
	if n.Pos != nil {
		t.Errorf("Default Pos should be nil")
	}
	if len(n.Roles) != 0 {
		t.Errorf("Default Roles should be empty")
	}
	if len(n.Props) != 0 {
		t.Errorf("Default Props should be empty")
	}
	if len(n.Children) != 0 {
		t.Errorf("Default Children should be empty")
	}
}

func makeTestTree() *Node {
	//      root
	//     / |  \
	//   c1 c2  c3
	//  /      /  \
	// gc1   gc2 gc3
	root := &Node{Id: 1, Type: "Root"}
	c1 := &Node{Id: 2, Type: "Child", Token: "c1"}
	c2 := &Node{Id: 3, Type: "Child", Token: "c2"}
	c3 := &Node{Id: 4, Type: "Child", Token: "c3"}
	gc1 := &Node{Id: 5, Type: "Grandchild", Token: "gc1"}
	gc2 := &Node{Id: 6, Type: "Grandchild", Token: "gc2"}
	gc3 := &Node{Id: 7, Type: "Grandchild", Token: "gc3"}
	c1.Children = []*Node{gc1}
	c3.Children = []*Node{gc2, gc3}
	root.Children = []*Node{c1, c2, c3}
	return root
}

func TestNodeFind(t *testing.T) {
	tree := makeTestTree()
	tests := []struct {
		name      string
		predicate func(*Node) bool
		expectIds []uint64
	}{
		{"Find all", func(n *Node) bool { return true }, []uint64{1, 2, 5, 3, 4, 6, 7}},
		{"Find children", func(n *Node) bool { return n.Type == "Child" }, []uint64{2, 3, 4}},
		{"Find none", func(n *Node) bool { return false }, nil},
		{"Find leaf", func(n *Node) bool { return n.Token == "gc2" }, []uint64{6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []uint64
			for _, n := range tree.Find(tt.predicate) {
				got = append(got, n.Id)
			}
			if !reflect.DeepEqual(got, tt.expectIds) {
				t.Errorf("Find: got %v, want %v", got, tt.expectIds)
			}
		})
	}
}

func TestNodePreOrder(t *testing.T) {
	tree := makeTestTree()
	var got []uint64
	tree.PreOrder(func(n *Node) { got = append(got, n.Id) })
	want := []uint64{1, 2, 5, 3, 4, 6, 7}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PreOrder: got %v, want %v", got, want)
	}
}

func TestNodePostOrder(t *testing.T) {
	tree := makeTestTree()
	var got []uint64
	tree.PostOrder(func(n *Node) { got = append(got, n.Id) })
	want := []uint64{5, 2, 3, 6, 7, 4, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("PostOrder: got %v, want %v", got, want)
	}
}

func TestNodeAncestors(t *testing.T) {
	tree := makeTestTree()
	gc2 := tree.Children[2].Children[0] // gc2
	anc := tree.Ancestors(gc2)
	var got []uint64
	for _, n := range anc {
		got = append(got, n.Id)
	}
	want := []uint64{1, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Ancestors: got %v, want %v", got, want)
	}
	// Not found
	fake := &Node{Id: 999}
	anc = tree.Ancestors(fake)
	if len(anc) != 0 {
		t.Errorf("Ancestors: expected empty for not found, got %v", anc)
	}
}

func TestNodeTransform(t *testing.T) {
	tree := makeTestTree()
	newTree := tree.Transform(func(n *Node) *Node {
		copy := *n
		copy.Type = "X" + n.Type
		return &copy
	})
	if newTree.Type != "XRoot" || newTree.Children[0].Type != "XChild" {
		t.Errorf("Transform: did not apply transformation correctly")
	}
	if newTree == tree || newTree.Children[0] == tree.Children[0] {
		t.Errorf("Transform: did not deep copy nodes")
	}
}

func TestNodeReplaceRemoveChild(t *testing.T) {
	parent := &Node{Type: "P"}
	c1 := &Node{Type: "C1"}
	c2 := &Node{Type: "C2"}
	parent.Children = []*Node{c1, c2}
	c3 := &Node{Type: "C3"}
	ok := parent.ReplaceChild(c2, c3)
	if !ok || parent.Children[1] != c3 {
		t.Errorf("ReplaceChild failed")
	}
	ok = parent.RemoveChild(c1)
	if !ok || len(parent.Children) != 1 || parent.Children[0] != c3 {
		t.Errorf("RemoveChild failed")
	}
	ok = parent.RemoveChild(&Node{Type: "X"})
	if ok {
		t.Errorf("RemoveChild should fail for non-existent child")
	}
}

func TestNodeNavigationEdgeCases(t *testing.T) {
	empty := &Node{}
	if len(empty.Find(func(*Node) bool { return true })) != 1 {
		t.Errorf("Find on single node should return itself")
	}
	var called int
	empty.PreOrder(func(*Node) { called++ })
	if called != 1 {
		t.Errorf("PreOrder on single node should call once")
	}
	called = 0
	empty.PostOrder(func(*Node) { called++ })
	if called != 1 {
		t.Errorf("PostOrder on single node should call once")
	}
	anc := empty.Ancestors(&Node{Id: 999})
	if len(anc) != 0 {
		t.Errorf("Ancestors on single node should be empty")
	}
}
