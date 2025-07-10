# Feature Requirements Document: Language Mapping Format for UAST

## Overview
This document specifies requirements for a mapping system that translates Tree-sitter node kinds to canonical UAST node types and roles. The mapping must be defined in YAML or JSON, support multiple languages, be extensible, and allow for hot-reloading and validation.

## Goals
- Provide a declarative mapping from Tree-sitter node kinds (per language) to UAST node types and roles.
- Support both YAML and JSON formats for mapping files.
- Enable easy extension for new languages and node kinds.
- Allow hot-reloading of mapping files at runtime (for server mode).
- Validate mapping files for correctness and completeness.
- Facilitate testability: mappings should be easily testable in isolation.

## Requirements

### 1. Mapping Format
- Each mapping file is keyed by language (e.g., `go`, `python`, `java`).
- For each language, map Tree-sitter node kinds to:
  - UAST node type (string, e.g., `FunctionDecl`)
  - Optional roles (list of strings, e.g., `["Function", "Declaration"]`)
  - Optional properties (map of string to value)
- Example YAML:
  ```yaml
  go:
    function_declaration:
      type: FunctionDecl
      roles: [Function, Declaration]
    identifier:
      type: Identifier
      roles: [Name]
  python:
    function_definition:
      type: FunctionDef
      roles: [Function, Declaration]
  ```
- Example JSON:
  ```json
  {
    "go": {
      "function_declaration": {"type": "FunctionDecl", "roles": ["Function", "Declaration"]},
      "identifier": {"type": "Identifier", "roles": ["Name"]}
    }
  }
  ```

### 2. Loader
- Implement a loader that reads mapping files (YAML/JSON) from disk.
- Loader must support loading a single file or a directory of files.
- Loader must merge mappings for the same language if split across files.
- Loader must provide a Go API for querying mappings by language and node kind.

### 3. Validator
- Validate that all required fields are present (e.g., `type` for each mapping).
- Warn or error on duplicate or conflicting mappings.
- Optionally, check for coverage of all node kinds in a language (if a reference list is available).

### 4. Hot-Reloading
- In server mode, support reloading mappings on SIGHUP or via an explicit API call.
- Ensure thread safety for concurrent access during reload.

### 5. Extensibility
- Adding a new language or node kind should require only editing/adding a mapping file.
- No code changes should be needed for new mappings.

### 6. Testability
- Provide test fixtures for mapping files (valid and invalid cases).
- Expose test helpers for loading and validating mappings in unit tests.

## Out of Scope
- Mapping Tree-sitter properties/fields to UAST properties (future work).
- UI for editing mappings (future work).

## Acceptance Criteria
- Loader and validator pass all provided test fixtures.
- Mappings can be hot-reloaded without server restart.
- Adding a new language or node kind is possible via mapping file only.
- Mapping format is documented in SPEC.md and README.md. 