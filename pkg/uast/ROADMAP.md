# UAST Package Implementation Roadmap (Detailed Plan)

## Progress Summary (2024-07-10)
- **UAST SPEC.md** written, including canonical types and roles
- **Mapping-driven construction** fully implemented and tested (integration test)
- **Integration test for SPEC compliance** added (Go function: type, roles, props)
- All mapping, provider, and loader logic refactored to be SPEC-driven
- **CLI tool 'uast' scaffolded** with cobra/viper, including parse/query/fmt/diff subcommands
- **'parse' command fully implemented and tested** (integration test, output capture, language detection, UAST JSON output)
- **CLI structure and help tested**
- Next: implement and test query/fmt/diff commands

## Current Architecture (2024-06)
- **Parser** (`parser.go`): Main entrypoint for UAST parsing. Handles provider loading and language detection.
- **Provider Loader** (`loader.go`): Loads and validates per-language provider configs and mappings.
- **Provider Factory** (`factory.go`): Instantiates native, TreeSitter, or external providers by language/config.
- **TreeSitterProvider**: Remains as the core implementation for Tree-sitter-based parsing.
- **All legacy EmbeddedProvider/factory abstractions removed.**

---

## Phase 1: Canonical Node & Core API

### 1.1 Define Canonical Node Structure
- [x] Implement `Node` struct with fields: `Id`, `Type`, `Token`, `Roles`, `Pos`, `Props`, `Children`.
- [x] Implement `Role` enum and `Positions` struct.
- [x] Add serialization (JSON, Protobuf) and deserialization (JSON implemented; Protobuf planned).
- [x] Write unit tests for node creation, serialization, and deserialization.

### 1.2 Core Go API
- [x] Implement navigation methods: `Find`, `PreOrder`, `PostOrder`, etc.
- [x] Implement mutation API: `Transform`.
- [x] Write table-driven tests for navigation and mutation (using hand-crafted trees).
- [x] All navigation and mutation methods are implemented, tested, and verified as of 2024-06-09.

---

## Phase 2: Tree-sitter Integration & Language Mapping

### 2.1 Tree-sitter Integration
- [x] Integrate go-sitter-forest for multi-language parsing. (2024-06-09)
- [x] Implement a generic parser interface. (2024-06-09)
- [x] Write tests for parsing simple files in at least 3 languages (Go, Python, Java). (2024-06-09)

### 2.2 Language Mapping
- [x] Design YAML/JSON mapping format: Tree-sitter node kinds → UAST types/roles. (2024-06-10)
- [x] Implement mapping loader and validator. (2024-06-10)
- [x] Write tests for mapping correctness and hot-reloading. (2024-06-10)
- [x] Refactor to merged provider/mapping config (2024-06-10)
- [x] Integrate loader with factory and provider initialization (2024-06-10)
- [x] Remove all legacy config and compatibility logic (2024-06-10)
- [x] Achieve full test coverage for loader, provider, and factory (2024-06-10)

### 2.3 UAST Converter
- [x] Implement `FromTree(tree *sitter.Tree, code []byte, lang string) *uast.Node`. (2024-06-09)
- [x] Map node kinds, fill positions, tokens, roles, and props. (2024-06-09)
- [x] Write golden tests: input code → expected UAST (compare JSON/types). (2024-06-09)

---

## Phase 3: Query & Transformation DSL

### 3.1 DSL Parser & Compiler
- [x] Design PEG grammar for DSL (using pigeon). (2024-07-10)
- [x] Implement parser and AST construction, with robust error handling. (2024-07-10)
- [x] Write table-driven unit tests for parsing and error cases. (2024-07-10)
- [x] Implement AST lowering to Go closures. (2024-07-10)

#### Progress Summary (2024-07-10)
- Canonical PEG grammar for the DSL designed and documented in `dsl_parser.peg`.
- Parser implemented using pigeon, with explicit action blocks for all rules.
- Comprehensive table-driven tests for valid and invalid DSL expressions.
- All tuple-wrapping, type assertion, and error propagation issues resolved.
- Error handling for unknown input and invalid tokens matches test suite requirements.
- **AST lowering to Go closures is complete.**
- **DSL runtime supports map, filter, reduce(count), field/literal access, pipelines, and compositional logic.**
- **All table-driven tests for parsing, lowering, and execution pass.**
- The language is now usable for real queries and transformations on UAST data.
- Next: Expand DSL runtime (3.2), add more operators, transformations, and integration.

### 3.2 DSL Runtime
- [x] Implement query execution engine (iterator pipeline).
- [x] Implement transformation primitives (`map`, `filter`, `reduce`, etc).
- [x] Write integration tests: query strings → expected node sets.

---

## Phase 4: CLI, gRPC, and Tooling

### 4.1 CLI Tool
- [x] Implement `uast parse`, `uast query`, `uast fmt`, `uast diff` commands. (parse implemented and tested, others in progress)
- [x] Write end-to-end tests (shell scripts or Go tests) for CLI commands. (parse tested, others in progress)

### 4.2 gRPC Daemon
- [ ] Define protobuf/gRPC service for remote parsing and querying.
- [ ] Implement server and client.
- [ ] Write integration tests for gRPC endpoints.

---

## Phase 5: Language Expansion & Performance

### 5.1 Add More Languages
- [ ] Add at least 5 more Tree-sitter grammars (JS, TS, Rust, Kotlin, C#).
- [ ] Write mapping files and golden tests for each.

### 5.2 Performance & Memory
- [ ] Benchmark parsing and query performance.
- [ ] Profile memory usage; optimize node allocation and traversal.
- [ ] Add fuzz tests for parser and DSL.

---

## Phase 6: Documentation & Examples

- [ ] Write GoDoc for all public APIs.
- [ ] Add usage examples for each major feature.
- [ ] Document mapping format and DSL syntax.
- [ ] Add a “How to add a new language” guide.

---

## Test Plan Summary

- **Unit tests** for all core structs, methods, and DSL parsing.
- **Golden tests** for language conversion (input code → expected UAST JSON).
- **Integration tests** for CLI and gRPC.
- **Fuzz tests** for parser and DSL.
- **Performance benchmarks** for parsing and queries.

---

**Note:** ucyt-based generic tests are currently disabled during core refactor and feature development.
