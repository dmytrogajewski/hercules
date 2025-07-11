# Feature Requirements Document: Node.FindDSL(query string) []*Node

## Motivation
- Enable querying of UAST trees using the project’s DSL, directly from Go code.
- Provide a simple, idiomatic API for users to run DSL queries on any UAST node and get matching nodes.
- Foundation for CLI, gRPC, and programmatic use cases.

## Requirements
- Add a method to the `Node` struct:
  ```go
  func (n *Node) FindDSL(query string) ([]*Node, error)
  ```
- The method must:
  - Parse the DSL query string using the existing DSL parser.
  - Lower the parsed AST to an executable Go closure or pipeline (as in the current runtime).
  - Execute the query starting from the receiver node (`n`), traversing the tree as needed.
  - Return all nodes matching the query as a slice.
  - Return an error if the query is invalid or execution fails.
- Must support all DSL features currently implemented (map, filter, reduce, field/literal access, pipelines, compositional logic).
- Must be concurrency-safe for read-only queries (no mutation of the tree).
- Must not require any global state.

## API
```go
// In node.go:
func (n *Node) FindDSL(query string) ([]*Node, error)
```
- Returns a slice of matching nodes, or an error.
- If the query is a reduce, returns the result as a single node (or a node wrapping the result value).

## Edge Cases
- Invalid DSL query: must return a descriptive error.
- Query returns no results: must return an empty slice, not nil.
- Query is a reduce: must return a single node (or node wrapping the result value).
- Query is a filter/map: must return all matching nodes.
- Query is empty: must return an error.
- Query references non-existent fields: must not panic, must return empty result or error.

## Test Plan
- Table-driven tests for:
  - Valid queries (map, filter, reduce, pipelines, field access, literals, composition)
  - Invalid queries (syntax errors, unknown fields, empty query)
  - Edge cases (no matches, reduce, deeply nested queries)
- Integration tests: run queries on real UAST trees and check results.
- Fuzz tests: (future) run random queries and ensure no panics.

## Out of Scope
- Iterator-based streaming API (future work, see roadmap)
- Mutation via DSL (future work)

## Notes
- This feature must not break or change the DSL syntax or semantics as currently implemented.
- If new facts are discovered during implementation, update this FRD and the spec as needed, as long as it does not contradict the roadmap. 