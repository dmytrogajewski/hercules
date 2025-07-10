# Feature Requirements Document: DSL Parser & Compiler (Phase 3.1)

## Overview
Implement a parser and compiler for the UAST Query & Transformation DSL. The parser will use a PEG grammar (via pigeon) to parse DSL query strings into an AST, which will then be lowered to Go closures for execution in the runtime. This is the foundation for all query and transformation features in the UAST system.

## Goals
- Parse DSL query strings into a well-typed AST.
- Support basic query and transformation primitives: `map`, `filter`, `reduce`, field access, literals, and function calls.
- Lower the AST to Go closures (functions) for later execution.
- Provide clear error messages for invalid queries.
- Be extensible for future DSL features (e.g., let bindings, subqueries).

## DSL Syntax & Grammar
- Use PEG grammar, implemented with [pigeon](https://github.com/mna/pigeon).
- Minimal initial syntax:
  - Expressions: `map`, `filter`, `reduce`, `.` (field access), literals (string, int, bool), function calls.
  - Example: `map(.children) | filter(.type == "FunctionDecl") | reduce(count)`
- Grammar must be documented in the code and in the roadmap/spec.

## AST Structure
- Define Go structs for each AST node type: Map, Filter, Reduce, Field, Literal, Call, etc.
- AST nodes must be typed and support source position info for error reporting.
- AST must be easily traversable and lowerable to Go closures.

## Lowering to Go Closures
- Implement lowering logic: AST → Go closure (func(Node) Node / []Node / value).
- Each DSL primitive must have a corresponding Go implementation.
- Lowering must be testable in isolation.

## Error Handling
- All parse errors must include line/column and a helpful message.
- Lowering errors (e.g., type errors) must be clear and actionable.

## Extensibility
- Grammar and AST must be designed for easy extension (e.g., add let, subqueries, custom functions).

## Testing Requirements
- Table-driven unit tests for:
  - Parsing valid/invalid DSL strings to AST.
  - Lowering AST to Go closures and evaluating on sample nodes.
- Golden tests for: DSL string → AST → closure → result.

## Out of Scope
- DSL runtime (iterator pipeline, execution engine) is Phase 3.2.
- CLI and gRPC integration is Phase 4.
- Performance optimization is Phase 5.

---

**Owner:** UAST Team
**Status:** Draft (2024-06) 