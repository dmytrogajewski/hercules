# Feature Requirements Document: Transform Mutation API

## Overview
Implement a robust, idiomatic, and efficient mutation API for UAST trees. The API must allow in-place mutation of nodes, support both pre-order and post-order traversal (with pre-order as default), and be fully covered by unit and integration tests. The function must be safe for nil roots, empty trees, and must document concurrency and mutation semantics.

## Goals
- Provide a minimal, expressive API for mutating UAST trees in-place.
- Ensure the API is idiomatic, efficient, and safe for all edge cases.
- Support both pre-order and post-order traversal (pre-order as default; post-order as future extensibility).
- Achieve 100% test coverage, including error and edge cases.

## API Requirements
- **Signature:** `func Transform(root *Node, fn func(*Node) bool)`
- **Description:** Applies the given function to each node in the tree, allowing mutation. Traversal is pre-order by default. If `fn` returns false, traversal does not descend into that node's children.
- **Requirements:**
  - Must traverse the tree in pre-order (root, then children left-to-right).
  - Must allow in-place mutation of nodes (fields, children, etc).
  - If `fn` returns false, must skip traversal of that node's children.
  - Must be safe for nil root (no panic, no-op).
  - Must be safe for empty trees and single-node trees.
  - Must be documented with usage examples.
  - Must not leak goroutines or stack (non-recursive implementation preferred).
  - Must be robust to user mutation of the tree during traversal (documented behavior).
  - Must be safe for concurrent reads (documented: tree should not be mutated concurrently).

## Extensibility & Edge Cases
- Must be extensible to support post-order and custom traversals in the future.
- Must handle nil nodes, empty trees, and very deep trees (stack safety).
- Must be robust to user mutation of the tree during traversal (documented: behavior is best-effort, but must not panic).
- Must be safe for concurrent reads (documented: tree should not be mutated concurrently).

## Out of Scope
- Post-order or custom traversals (future work).
- Parallel or concurrent mutation (future work).

## Testability
- Must be covered by table-driven unit tests:
  - Nil root
  - Empty tree
  - Single node
  - Multi-level tree
  - Deep tree (stack safety)
  - Mutation of node fields and children
  - Skipping children by returning false
  - User mutation during traversal (documented, must not panic)
- Integration tests for large trees (performance, stack safety).
- Usage examples in documentation.

## Example Usage
```go
root := &Node{Type: "Root", Children: []*Node{...}}
Transform(root, func(n *Node) bool {
    n.Token = strings.ToUpper(n.Token)
    return true // continue traversal
})
```

--- 