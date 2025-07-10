# Feature Requirements Document: Tree-sitter Integration

## Overview
Integrate the go-sitter-forest library to enable multi-language parsing in the UAST package. Provide a generic parser interface and ensure robust, tested support for at least Go, Python, and Java source files.

## Goals
- Enable parsing of multiple programming languages using Tree-sitter grammars via go-sitter-forest.
- Provide a generic, idiomatic Go interface for parsing source code into canonical UAST nodes.
- Ensure the system is easily extensible to new languages.
- Provide tests for at least Go, Python, and Java.

## Requirements

### Integration
- Add go-sitter-forest as a dependency.
- Ensure all required Tree-sitter grammars (Go, Python, Java) are available and registered.
- Provide a mechanism to detect language from filename (extension-based, config-driven).

### Parser Interface
- Define a `Provider` interface:
  - `Parse(filename string, content []byte) (*Node, error)`
  - `SupportedLanguages() []string`
  - `IsSupported(filename string) bool`
- Implement an `EmbeddedProvider` that uses go-sitter-forest for supported languages.
- Each language must have a `LanguageProvider` implementation that wraps the corresponding Tree-sitter grammar.

### Tests
- Table-driven tests for parsing simple files in Go, Python, and Java.
- Tests must verify:
  - Correct language detection from filename.
  - Successful parsing (non-nil root node, correct type/token for root).
  - Reasonable tree structure (root has children, etc.).
- Edge cases: empty file, unsupported extension, invalid code.

### General
- All code must be idiomatic Go and follow project style.
- All public methods must have GoDoc comments.
- No recursion in conversion code (explicit stack if needed).

## Out of Scope
- Language mapping (node kind → UAST type/roles) and advanced conversion logic (covered in later phases).
- Babelfish/external providers.

## Acceptance Criteria
- go-sitter-forest is integrated and working for Go, Python, and Java.
- EmbeddedProvider and LanguageProvider are implemented and tested.
- All tests for parsing and language detection pass.
- Code is idiomatic, clean, and extensible.

---

*Last updated: 2024-06-09* 