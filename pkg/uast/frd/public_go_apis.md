# Feature Requirements Document: Public Go APIs (Navigation, Streaming, Mutation via DSL)

## Overview
Implement a set of ergonomic, composable, and testable Go APIs for navigating, querying, and mutating UASTs using the project’s DSL. These APIs must be idiomatic, efficient, and extensible, and must use the UAST DSL for all queries (no XPath support).

## Goals
- Provide a minimal, expressive API surface for UAST navigation, streaming, and mutation.
- All query/navigation APIs must use the UAST DSL (not XPath).
- Support streaming traversal for large trees.
- Support efficient role checks and membership predicates.
- Enable safe, composable mutation of UASTs.
- Ensure all APIs are fully covered by unit and integration tests.

## API Requirements

### 1. Query API
- **Signature:** `func (n *Node) FindDSL(query string) []*Node`
- **Description:** Returns all nodes matching the given DSL query, starting from the receiver node.
- **Requirements:**
  - Must parse and lower the DSL query string to an executable pipeline.
  - Must support all current DSL features (filter, map, reduce, field/property access, membership with `has`).
  - Must return a slice of nodes matching the query.
  - Must propagate and return errors for invalid queries.

### 2. Streaming Iterator
- **Signature:** `func PreOrder(root *Node) <-chan *Node`
- **Description:** Returns a channel streaming nodes in pre-order traversal.
- **Requirements:**
  - Must be non-blocking and safe for large trees.
  - Must close the channel when traversal is complete.
  - Must be usable in range loops.

### 3. Role Check Utility
- **Signature:** `func HasRole(node *Node, role Role) bool`
- **Description:** Returns true if the node has the given role.
- **Requirements:**
  - Must use O(1) lookup (hash-set or equivalent).
  - Must be safe for nil nodes and empty roles.

### 4. Mutation API
- **Signature:** `func Transform(root *Node, fn func(*Node) bool)`
- **Description:** Applies the given function to each node in the tree, allowing mutation.
- **Requirements:**
  - Must traverse the tree in pre-order or post-order (documented).
  - Must allow in-place mutation of nodes.
  - Must be safe for concurrent use if possible.

## Extensibility & Edge Cases
- APIs must be extensible to support future DSL features (e.g., recursive queries, additional operators).
- Must handle nil nodes, empty trees, and invalid queries gracefully.
- Must be documented with usage examples.
- Must be covered by table-driven and integration tests.

## Out of Scope
- XPath support (explicitly not supported).
- Non-DSL query languages.

## Testability
- All APIs must be covered by unit tests (including error cases and edge cases).
- Integration tests must cover end-to-end usage (parse → query → mutate → serialize).

--- 