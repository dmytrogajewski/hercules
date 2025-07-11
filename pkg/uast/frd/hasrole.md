# Feature Requirements Document: HasRole Utility Function

## Overview
Implement a robust, idiomatic, and efficient utility function for checking if a UAST node has a given role. The API must be safe for nil nodes, empty roles, and must be fully covered by unit and integration tests. The function must be O(1) and suitable for use in performance-sensitive code.

## Goals
- Provide a minimal, expressive API for role membership checks on UAST nodes.
- Ensure the API is idiomatic, efficient, and safe for all edge cases.
- Achieve 100% test coverage, including error and edge cases.

## API Requirements
- **Signature:** `func HasRole(node *Node, role Role) bool`
- **Description:** Returns true if the node has the given role.
- **Requirements:**
  - Must use O(1) lookup (slice scan is acceptable for small slices, but must be documented).
  - Must be safe for nil node (returns false, does not panic).
  - Must be safe for empty roles (returns false, does not panic).
  - Must be safe for nodes with no roles (returns false, does not panic).
  - Must be documented with usage examples.

## Extensibility & Edge Cases
- Must be extensible to support future role representations (e.g., map/set for large role sets).
- Must handle nil nodes, empty roles, and nodes with no roles gracefully.
- Must be robust to user mutation of the roles slice during checks (documented behavior).
- Must be safe for concurrent reads (documented: node should not be mutated during check).

## Out of Scope
- Role assignment or mutation APIs (future work).
- Role enumeration or listing (future work).

## Testability
- Must be covered by table-driven unit tests:
  - Nil node
  - Node with no roles
  - Node with one role
  - Node with multiple roles
  - Empty role string
  - Role not present
  - Role present
  - User mutation during check (documented, must not panic)
- Integration tests for performance (optional, for large role sets).
- Usage examples in documentation.

## Example Usage
```go
n := &Node{Roles: []Role{"Exported", "Test"}}
if HasRole(n, "Exported") {
    fmt.Println("Node is exported")
}
```

--- 