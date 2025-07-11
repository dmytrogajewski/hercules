# Feature Requirements Document: PreOrder Streaming Iterator

## Overview
Implement a robust, idiomatic, and efficient streaming iterator for pre-order traversal of UAST trees. The API must be safe for large trees, handle all edge cases (including nil and empty trees), and be fully covered by unit and integration tests. The iterator must be usable in Go range loops and must close the channel when traversal is complete.

## Goals
- Provide a minimal, expressive API for streaming pre-order traversal of UAST trees.
- Ensure the API is idiomatic and safe for use in concurrent and large-tree scenarios.
- Handle all edge cases gracefully (nil root, empty tree, single node, deep trees).
- Document usage and edge cases with examples.
- Achieve 100% test coverage, including error and edge cases.

## API Requirements
- **Signature:** `func PreOrder(root *Node) <-chan *Node`
- **Description:** Returns a channel streaming nodes in pre-order (root, then children left-to-right).
- **Requirements:**
  - Must be non-blocking for the caller (traversal runs in a goroutine).
  - Must close the channel when traversal is complete.
  - Must be safe for nil root (channel closes immediately, no panic).
  - Must be safe for empty trees and single-node trees.
  - Must be usable in Go `for n := range PreOrder(root)` loops.
  - Must not leak goroutines or channels.
  - Must not panic on malformed trees (e.g., cycles, though cycles are not expected in UAST).
  - Must be documented with usage examples.

## Extensibility & Edge Cases
- Must be extensible to support future traversal options (e.g., post-order, breadth-first).
- Must handle nil nodes, empty trees, and very deep trees (stack safety).
- Must be robust to user mutation of the tree during traversal (documented behavior).
- Must be safe for concurrent reads (documented: tree should not be mutated during traversal).

## Out of Scope
- Post-order, breadth-first, or custom traversals (future work).
- Parallel traversal (future work).
- Mutation during traversal (undefined behavior, but must not panic).

## Testability
- Must be covered by table-driven unit tests:
  - Nil root
  - Empty tree
  - Single node
  - Multi-level tree
  - Deep tree (stack safety)
  - User mutation during traversal (documented, must not panic)
- Integration tests for large trees (performance, channel closure).
- Usage examples in documentation.

## Example Usage
```go
root := &Node{Type: "Root", Children: []*Node{...}}
for n := range PreOrder(root) {
    fmt.Println(n.Type)
}
```

--- 