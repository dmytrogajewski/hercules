# Feature Requirements Document: Core Go API (Navigation & Mutation)

## Overview
Implement core navigation and mutation methods for the canonical `Node` struct in the UAST package. This API will enable traversal, search, and transformation of UAST trees in a clean, idiomatic Go style.

## Goals
- Provide robust, efficient, and easy-to-use methods for navigating and mutating UAST trees.
- Support common tree operations: search, traversal (pre-order, post-order), and transformation.
- Ensure all methods follow clean code principles and are well-tested.

## Requirements

### Navigation Methods
- `Find(predicate func(*Node) bool) []*Node`
  - Returns all nodes in the tree (including root) for which `predicate(node)` is true.
- `PreOrder(fn func(*Node))`
  - Visits all nodes in pre-order (root, then children left-to-right), calling `fn(node)` for each.
- `PostOrder(fn func(*Node))`
  - Visits all nodes in post-order (children left-to-right, then root), calling `fn(node)` for each.
- `Ancestors(target *Node) []*Node`
  - Returns the list of ancestors from root to the parent of `target` (empty if not found).

### Mutation Methods
- `Transform(fn func(*Node) *Node) *Node`
  - Returns a new tree where each node is replaced by the result of `fn(node)` (post-order, non-recursive, preserves structure).
- `ReplaceChild(old, new *Node) bool`
  - Replaces the first occurrence of `old` in `Children` with `new`. Returns true if replaced.
- `RemoveChild(target *Node) bool`
  - Removes the first occurrence of `target` from `Children`. Returns true if removed.

### General
- All methods must be non-recursive (use explicit stack if needed).
- All methods must be covered by table-driven unit tests in `node_test.go`.
- All public methods must have clear, descriptive names and GoDoc comments.

## Out of Scope
- Query DSL, advanced pattern matching, or language-specific logic.
- Mutations that break tree invariants (e.g., cycles).

## Acceptance Criteria
- All navigation and mutation methods are implemented and tested.
- Tests cover edge cases: empty tree, single node, deep trees, wide trees, no matches, etc.
- Code is idiomatic, clean, and follows project style.
- No recursion is used in any method.

---

*Last updated: 2024-06-09* 