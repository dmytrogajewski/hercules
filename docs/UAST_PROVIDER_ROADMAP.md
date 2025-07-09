# Embedded UAST Provider Roadmap

This document tracks the implementation plan, progress, and usage for the embedded UAST provider (Babelfish drop-in replacement) in Hercules.

## Overview
The embedded UAST provider allows Hercules to perform UAST-based analyses without an external Babelfish server, using built-in parsers (starting with Go). This enables offline, CI, and simplified deployments.

---

## Implementation Checklist

- [x] **Define UASTProvider interface**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Internal abstraction_

- [x] **Implement GoEmbeddedProvider (Go parser)**  
  _Status: Implemented_  
  _Test: [ ]_  
  _Usage: Used when `--uast-provider=embedded` and file is Go source_

- [x] **Add CLI/config flag to select provider**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: `./hercules --shotness --uast-provider=embedded <repo>`

- [x] **Refactor pipeline to use UASTProvider interface**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Transparent to user; Extractor uses Provider field, CLI flag selects backend_

- [x] **Add tests for GoEmbeddedProvider**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Run `go test ./internal/plumbing/uast`; see provider_test.go for unit tests_

- [x] **Document feature in README**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: See README section "Embedded UAST Provider" for usage, supported languages, and roadmap reference_

- [x] **Add Tree-sitter support for Java**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Used when `--uast-provider=embedded` and file is .java; see provider_test.go for test_

- [x] **Add Tree-sitter support for Kotlin**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Used when `--uast-provider=embedded` and file is .kt; see provider_test.go for test_

- [x] **Add Tree-sitter support for Swift**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Used when `--uast-provider=embedded` and file is .swift; see provider_test.go for test_

- [x] **Add Tree-sitter support for JavaScript**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Used when `--uast-provider=embedded` and file is .js/.jsx; see provider_test.go for test_

- [x] **Add Tree-sitter support for Rust**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Used when `--uast-provider=embedded` and file is .rs; see provider_test.go for test_

- [x] **Add Tree-sitter support for PHP**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Used when `--uast-provider=embedded` and file is .php; see provider_test.go for test_

- [x] **Add Tree-sitter support for Python**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Used when `--uast-provider=embedded` and file is .py; see provider_test.go for test_

- [x] **Use embedded provider by default**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Embedded provider is now default for all supported file types; see Extractor.Consume logic and provider_test.go_

- [x] **Add Tree-sitter support for TypeScript/TSX**  
  _Status: Implemented_  
  _Test: [x]_  
  _Usage: Used when `--uast-provider=embedded` and file is .ts/.tsx; see provider_test.go for test_

---

## Usage Examples

- **Use embedded provider for Go:**
  ```sh
  ./hercules --shotness --uast-provider=embedded <repo>
  ```

- **Default (Babelfish):**
  ```sh
  ./hercules --shotness <repo>
  ```

---

## Notes
- If a file's language is unsupported, the provider should skip or warn, not fail the analysis.
- This roadmap should be updated as features are implemented and tested. 

---

## Embedded UAST: Translation & Compatibility Roadmap

To achieve full drop-in compatibility with Babelfish UASTs, the embedded provider must translate Tree-sitter (sitter.Node) and Go AST nodes to the Babelfish UAST model (`nodes.Node`). The following features are required for full functionality:

### 1. Node Structure Translation
- [x] Convert `*sitter.Node` (and Go AST nodes) to `nodes.Object` recursively, preserving:
  - Node type/kind (e.g., "uast:Class", "uast:Function")
  - Children as `nodes.Array`
  - Text content, identifiers, and literals
  - Parent/child relationships
  
  _Complete: Recursive conversion of children and text content implemented in `internal/uastconvert`._

### 2. Position and Range Information
- [x] Map Tree-sitter node positions (start/end byte, line, column) to UAST node attributes (`@pos`, `@start`, `@end`)
- [x] Ensure all nodes have position info for downstream analyses

  _Complete: Position mapping from Tree-sitter to UAST format implemented in `internal/uastconvert`._

### 3. UAST XPath/Query Compatibility
- [x] Ensure translated nodes are compatible with `bblfsh/sdk.v2/uast/tools` for XPath queries
- [x] Implement or adapt node attributes so XPath queries (used in analyses like Shotness) work as expected

  _Complete: XPath-compatible node attributes (@type, @pos, @start, @end, @token) and structure implemented in `internal/uastconvert`._

### 4. Serialization and Hashing
- [x] Ensure converted nodes can be serialized to Protocol Buffers format
- [x] Implement proper hashing for node comparison and deduplication
- [x] Test serialization compatibility with existing UAST tools

  _Complete: Serialization and hashing support implemented and tested in `internal/uastconvert`._

### 5. Node Types and Attributes
- [x] Map Tree-sitter/Go AST node types and attributes to UAST node types and attributes (e.g., "uast:FunctionGroup", "Name", etc.)
- [x] Ensure compatibility with Babelfish UAST conventions for downstream analyses

  _Complete: Node type and attribute mapping implemented in `internal/uastconvert`._

### 6. Testing and Validation
- [x] Add integration tests to ensure converted UASTs are compatible with Hercules analyses and XPath queries
- [x] Validate against real code samples and compare with Babelfish UASTs

  _Complete: Integration tests implemented and robust for structure and type mapping. Attribute extraction (e.g., Name) from Tree-sitter nodes is a known limitation and marked as a TODO for future work._

### 7. Performance and Robustness
- [x] Ensure conversion is robust against panics and invalid input (Tree-sitter edge cases)
- [ ] Benchmark conversion speed and memory usage
- [ ] Add fuzz/invariant tests for edge cases

  _Complete: Conversion is robust against panics and invalid input. TODO: Add benchmarks and fuzz tests for further validation._

---

**This section should be updated as translation features are implemented.** 