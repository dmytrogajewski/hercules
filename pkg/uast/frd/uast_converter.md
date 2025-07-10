# Feature Requirements Document: UAST Converter (FromTree)

## Overview
Implement the core conversion logic from Tree-sitter parse trees to the canonical UAST node structure. This converter will be the heart of the system, producing language-agnostic UASTs from any supported Tree-sitter grammar, and will be extensible via mapping files.

## Goals
- Provide a fast, robust, and extensible conversion from Tree-sitter ASTs to canonical UAST nodes.
- Support mapping of Tree-sitter node kinds to UAST types and roles via YAML/JSON mapping files.
- Fill all required node fields: Type, Token, Roles, Positions, Props, Children.
- Expose a single entrypoint:
  - `func FromTree(tree *sitter.Tree, code []byte, lang string) *uast.Node`
- Ensure all code is non-recursive (explicit stack for traversal).
- Provide golden tests: input code → expected UAST (compare JSON).

## Requirements

### Conversion Logic
- Traverse the Tree-sitter AST using an explicit stack (no recursion).
- For each Tree-sitter node:
  - Map `Kind()` to canonical `Type` using a language-specific mapping (YAML/JSON, hot-reloadable).
  - Copy byte/line/col ranges into `Positions`.
  - For leaves, store `node.Content(code)` as `Token`.
  - Assign `Roles` using mapping and/or heuristics.
  - Copy any extra properties into `Props`.
  - Build the `Children` array in order.
- Compute a stable 64-bit hash for each node (FNV over Type, Token, child hashes).

### Mapping
- Mapping files (YAML/JSON) define Tree-sitter node kind → UAST type/roles for each language.
- Auto-generate a default mapping from `node_types.json` if no mapping is provided.
- Support hot-reloading of mapping files at runtime.

### API
- `func FromTree(tree *sitter.Tree, code []byte, lang string) *uast.Node`
- All code in package `uast`.

### Tests
- Golden tests: for each language, input code → expected UAST (compare JSON, ignore Id/hash).
- Edge cases: empty file, syntax error, deeply nested, wide trees.
- Table-driven tests for mapping correctness and error handling.

### General
- All code must be idiomatic Go and follow project style.
- All public methods must have GoDoc comments.
- No recursion in conversion code (explicit stack if needed).

## Out of Scope
- Query DSL, advanced role inference, or semantic analysis (future phases).
- Protobuf serialization (covered elsewhere).

## Acceptance Criteria
- FromTree is implemented, tested, and documented.
- Mapping files are supported and hot-reloadable.
- All golden tests pass for Go, Python, Java.
- Code is idiomatic, clean, and extensible.

---

*Last updated: 2024-06-09* 