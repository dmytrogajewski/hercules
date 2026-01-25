# FRD-1.1: Single-Pass UAST Visitor

## 1. Overview
Current UAST analysis is inefficient because each analyzer traverses the AST independently. This results in O(A * N) complexity where A is the number of analyzers and N is the number of nodes. This FRD proposes a refactoring to a Visitor pattern where the AST is traversed exactly once, and events are dispatched to registered analyzers.

## 2. Goals
- Reduce AST traversal overhead by implementing a single-pass mechanism.
- Maintain backward compatibility with existing `StaticAnalyzer` interface where possible, or adapt it.
- Migrate core analyzers (Complexity, Halstead, Cohesion, Comments) to the new pattern.
- Achieve 2-5x performance improvement.

## 3. Architecture

### 3.1. NodeVisitor Interface
A new interface `NodeVisitor` will be defined in `pkg/analyzers/common/visitor.go`.

```go
type NodeVisitor interface {
    // OnEnter is called when entering a node.
    OnEnter(node *node.Node, depth int)
    // OnExit is called when exiting a node (after children).
    OnExit(node *node.Node, depth int)
}
```

Analyzers that only need specific node types can implement a filtering mechanism or the traverser can handle it. To optimize, we will implement a hook registry.

### 3.2. MultiAnalyzerTraverser
A `MultiAnalyzerTraverser` in `pkg/analyzers/common/multi_traverser.go` will manage multiple visitors.

```go
type MultiAnalyzerTraverser struct {
    visitors []NodeVisitor
    // map[NodeType][]NodeVisitor
    hooks    map[string][]NodeVisitor
}

func (t *MultiAnalyzerTraverser) RegisterVisitor(v NodeVisitor)
func (t *MultiAnalyzerTraverser) RegisterHook(nodeType string, v NodeVisitor)
func (t *MultiAnalyzerTraverser) Traverse(root *node.Node)
```

### 3.3. Analyzer Migration
Existing analyzers need to be refactored to implement `NodeVisitor` instead of (or in addition to) using `UASTTraverser` to find nodes.

For example, `ComplexityAnalyzer` currently uses `FindNodesByType` to find functions, then traverses each function.
New approach:
- `ComplexityAnalyzer` implements `OnEnter`.
- If node is Function, it initializes metrics for that function.
- It tracks nesting depth and other metrics as it traverses children.
- On `OnExit` of Function, it finalizes metrics.

## 4. Implementation Plan

1.  **Define Interfaces**: Create `pkg/analyzers/common/visitor.go` with `NodeVisitor`.
2.  **Implement Traverser**: Create `pkg/analyzers/common/multi_traverser.go` with `MultiAnalyzerTraverser` and hook registry.
3.  **Migrate Analyzers**:
    *   Complexity: `pkg/analyzers/complexity/visitor.go`
    *   Halstead: `pkg/analyzers/halstead/visitor.go`
    *   Cohesion: `pkg/analyzers/cohesion/visitor.go`
    *   Comments: `pkg/analyzers/comments/visitor.go`
4.  **Update Factory**: Modify `pkg/analyzers/analyze/analyzer.go` to use `MultiAnalyzerTraverser` when running multiple analyzers.

## 5. Testing Strategy
- **Unit Tests**: Test `MultiAnalyzerTraverser` with mock visitors to ensure correct dispatch and order.
- **Integration Tests**: Verify that migrated analyzers produce identical results to legacy ones.
- **Benchmarks**: Compare performance of old vs new approach on large ASTs.

## 6. Detailed Design

### 6.1 `NodeVisitor`
```go
package common

import "github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"

type NodeVisitor interface {
	OnEnter(n *node.Node, depth int)
	OnExit(n *node.Node, depth int)
}
```

### 6.2 `MultiAnalyzerTraverser`
```go
package common

import "github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"

type MultiAnalyzerTraverser struct {
	visitors []NodeVisitor
}

func NewMultiAnalyzerTraverser() *MultiAnalyzerTraverser {
	return &MultiAnalyzerTraverser{}
}

func (t *MultiAnalyzerTraverser) RegisterVisitor(v NodeVisitor) {
	t.visitors = append(t.visitors, v)
}

func (t *MultiAnalyzerTraverser) Traverse(n *node.Node) {
    if n == nil {
        return
    }
    t.traverseRecursive(n, 0)
}

func (t *MultiAnalyzerTraverser) traverseRecursive(n *node.Node, depth int) {
    for _, v := range t.visitors {
        v.OnEnter(n, depth)
    }

    for _, child := range n.Children {
        t.traverseRecursive(child, depth+1)
    }

    for _, v := range t.visitors {
        v.OnExit(n, depth)
    }
}
```

## 7. Acceptance Criteria
- All tests in `pkg/analyzers/...` pass.
- `MultiAnalyzerTraverser` implementation is complete and tested.
- `ComplexityAnalyzer` is migrated and passes tests.
- Benchmarks show improvement.
