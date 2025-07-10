# Feature Requirements Document: DSL Runtime (Phase 3.2)

## Overview
This document specifies the requirements for the DSL Runtime component of the UAST package. The DSL Runtime is responsible for executing parsed and lowered DSL queries on UASTs, supporting transformation primitives, and providing a composable, testable query engine.

## Goals
- Provide a robust, composable query execution engine for the UAST DSL.
- Support transformation primitives: `map`, `filter`, `reduce`, and pipelines.
- Enable integration tests: query strings → expected node sets.
- Ensure extensibility for future operators and transformations.

## Functional Requirements
1. **Query Execution Engine**
   - Accepts a lowered DSL query (Go closure) and a UAST root node (or node set).
   - Executes the query, returning a set of result nodes.
   - Supports pipelined execution (chaining of map/filter/reduce/etc.).
   - Handles errors gracefully (invalid queries, runtime errors, etc.).

2. **Transformation Primitives**
   - **map:** Applies a field or expression to each node in the input set.
   - **filter:** Selects nodes matching a predicate.
   - **reduce:** Aggregates a node set (e.g., `count`).
   - **pipeline:** Composes multiple transformations in sequence.
   - **Extensibility:** New primitives/operators can be added with minimal changes.

3. **Integration Tests**
   - Table-driven tests: (query string, input UAST) → expected output node set.
   - Cover all primitives, pipelines, and edge cases.
   - Test error handling (invalid queries, runtime errors).

## Non-Functional Requirements
- **Performance:** Should handle UASTs of at least 10,000 nodes with <1s latency for simple queries.
- **Testability:** All logic must be covered by unit and integration tests.
- **Extensibility:** Adding new operators/primitives should not require major refactoring.
- **Documentation:** All public APIs and DSL features must be documented.

## Design Constraints
- Must not break existing DSL parser/lowering logic.
- Must use the canonical `Node` struct and compositional QueryFunc interface.
- Must not introduce recursion in a way that risks stack overflows (prefer explicit stacks/loops).
- Must be compatible with future CLI/gRPC integration.

## Out of Scope
- CLI and gRPC integration (covered in later phases).
- Language-specific UAST quirks (handled by provider/mapping layer).

## Acceptance Criteria
- All table-driven integration tests pass for map, filter, reduce, and pipelines.
- Query execution engine is composable, robust, and documented.
- Code is reviewed and merged to mainline.

---

*Last updated: 2024-07-10* 