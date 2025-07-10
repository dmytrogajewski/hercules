# Feature Requirements Document: Canonical UAST Node Struct

## Overview
Implement the core `Node` struct for the Unified Abstract Syntax Tree (UAST) in Go. This struct is the foundation for all UAST operations, serialization, and language conversions.

## Requirements

### 1. Data Structure
- The `Node` struct must have the following fields:
  - `Id uint64` — Stable hash for diffing and identity.
  - `Type string` — Canonical, language-agnostic type (e.g., "Function").
  - `Token string` — Surface lexeme if the node is a leaf; empty otherwise.
  - `Roles []Role` — Syntactic/semantic labels (enum or string constants).
  - `Pos *Positions` — Optional; byte offsets and line/col info. May be nil for memory efficiency.
  - `Props map[string]string` — Free-form properties for language-specific or extra data.
  - `Children []*Node` — Ordered list of child nodes.

### 2. Roles and Positions
- `Role` should be an extensible enum or string type, supporting both standard and custom roles.
- `Positions` struct must include at least: `StartLine`, `StartCol`, `StartOffset`, `EndLine`, `EndCol`, `EndOffset`.

### 3. Serialization
- Must support JSON serialization/deserialization for debugging and tests.
- Must support Protocol Buffers serialization for performance (define a matching proto schema, but initial implementation may stub this).
- JSON output should omit nil/empty fields where idiomatic.

### 4. API & Methods
- Provide constructor(s) for easy node creation.
- Provide methods for:
  - Adding/removing children
  - Navigating children (e.g., `Find`, `PreOrder`, `PostOrder` — may be stubbed for now)
  - String representation for debugging

### 5. Extensibility
- The struct and related types must be designed for future extension (e.g., new fields, roles, or properties) without breaking existing code.

### 6. Testing
- Unit tests must cover:
  - Node creation with all fields
  - Serialization/deserialization (JSON)
  - Adding/removing children
  - String representation
  - Edge cases: nil positions, empty roles/props/children

### 7. Documentation
- All exported types and methods must have GoDoc comments.

## Out of Scope
- DSL, query, and transformation logic (covered in later phases)
- Language-specific conversion (covered in converter phase)

## Acceptance Criteria
- All requirements above are met and verified by tests.
- Code is idiomatic, documented, and ready for integration with the rest of the UAST pipeline. 