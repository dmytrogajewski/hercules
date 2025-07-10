# UAST Provider Language Integration Checklist

When adding a new language provider to `pkg/uast`, ensure the following default tests are written. Each test should be provided as a YAML use case file (see examples below).

## Checklist

- [ ] 1. **Basic Parsing**: Can parse a minimal valid file for the language and produce a non-empty UAST.
- [ ] 2. **Node Type Extraction**: Can extract key node types (e.g., function, class, method) from a sample file.
- [ ] 3. **Change Detection**: Detects changes between two versions of a file (e.g., function added/removed/modified).
- [ ] 4. **Language Detection**: Correctly identifies the language from a sample file.
- [ ] 5. **Error Handling**: Returns a clear error for invalid or malformed input.
- [ ] 6. **Edge Cases**: Handles empty files, comments-only files, and files with only whitespace.

---

## YAML Examples

### 1. Basic Parsing
```yaml
name: Basic parsing for Go
provider: go
input: |
  package main
  func main() {}
assert: node != null && node.children.size() > 0
```

### 2. Node Type Extraction
```yaml
name: Extract function node in Python
provider: python
input: |
  def foo():
    pass
assert: node.children.exists(n, n.type == "function_definition")
```

### 3. Change Detection
```yaml
name: Detect function addition in JavaScript
provider: javascript
before: |
  function a() {}
after: |
  function a() {}
  function b() {}
assert: changes.exists(c, c.type == "function" && c.action == "added")
```

### 4. Language Detection
```yaml
name: Detect language for Rust
provider: rust
input: |
  fn main() {}
assert: detect_language(input) == "rust"
```

### 5. Error Handling
```yaml
name: Error on invalid Kotlin
provider: kotlin
input: |
  fun { invalid syntax
assert: error != null && error.contains("syntax")
```

### 6. Edge Cases
```yaml
name: Handle empty file in PHP
provider: php
input: ""
assert: node == null || node.children.size() == 0
```

---

**Note:**
- Use CEL expressions for assertions.
- Place YAML files in the appropriate test directory (e.g., `usecases/<language>/`).
- Add more tests as needed for language-specific features. 