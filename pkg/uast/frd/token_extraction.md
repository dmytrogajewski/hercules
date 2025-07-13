# Feature Requirements Document: Token Extraction in Language Mapping

## Overview
Extend the UAST language mapping system to support extracting token values from source code using Tree-sitter node positions. This feature enables mappings to extract literal values (strings, numbers, identifiers) from the source code and include them as tokens in the resulting UAST nodes.

## Goals
- Enable mappings to extract token values from source code using byte positions
- Support multiple extraction strategies (self, child fields, descendants)
- Maintain backward compatibility with existing mapping format
- Provide a clean, declarative syntax for token extraction
- Ensure robust error handling for invalid byte ranges

## Requirements

### 1. Token Extraction Mapping
- Extend the existing mapping format to support token extraction
- Add a `token` field to mapping entries that specifies how to extract the token value
- Support multiple extraction sources:
  - `"self"` - Extract from the node's own byte range
  - `"child:fieldname"` - Extract from a child node with the specified field name
  - `"descendant:nodetype"` - Extract from the first descendant of the specified type

### 2. Mapping Format Extension
- Extend the current YAML mapping format:
  ```yaml
  language: perl
  extensions: [.pl, .pm]
  parser: tree-sitter
  mapping:
    interpolated_string_literal:
      type: Literal
      roles: [Literal]
      token: "self"  # Extract from node's own text
    string_content:
      type: Literal
      roles: [Literal]
      token: "child:fieldname"  # Extract from a child node with the specified field name
    bareword:
      type: Identifier
      roles: [Name]
      token: "descendant:nodetype"  #  Extract from the first descendant of the specified type
  ```

### 3. Implementation Requirements
- Extend the `Mapping` struct to include a `Token` field
- Add token extraction logic to the `TreeSitterProvider.ToCanonicalNode()` method
- Implement `extractTokenFromNode()` function with support for:
  - Self-extraction (current node's byte range)
  - Child field extraction (using Tree-sitter's field API)
  - Descendant extraction (recursive search)
- Ensure proper error handling for invalid byte ranges

### 4. Backward Compatibility
- Existing mappings without `token` field must continue to work
- Default behavior should remain unchanged for existing mappings
- Token extraction should be optional and non-breaking

### 5. Error Handling
- Gracefully handle cases where byte ranges are invalid
- Log warnings for extraction failures but don't fail the entire mapping
- Provide fallback behavior when token extraction fails

### 6. Testing Requirements
- Unit tests for each extraction strategy
- Integration tests with real Tree-sitter ASTs
- Test cases for error conditions (invalid byte ranges, missing children)
- Update existing language tests to use token extraction where appropriate

## Technical Design

### Mapping Struct Extension
```go
type Mapping struct {
    Type     string            `yaml:"type"`
    Roles    []string          `yaml:"roles,omitempty"`
    Props    map[string]string `yaml:"props,omitempty"`
    Token    string            `yaml:"token,omitempty"`  // NEW: token extraction source
    // ... existing fields
}
```

### Token Extraction Function
```go
func extractTokenFromNode(node *TreeSitterNode, source string) string {
    switch source {
    case "text", "self":
        // Extract from node's own text
        if node.Root.ChildCount() == 0 {
            return extractTextFromRange(node.Root, node.Source)
        }
    case strings.HasPrefix(source, "child:"):
        // Extract from child field
        fieldName := strings.TrimPrefix(source, "child:")
        return extractTokenFromChildField(node, fieldName)
    case strings.HasPrefix(source, "descendant:"):
        // Extract from descendant
        nodeType := strings.TrimPrefix(source, "descendant:")
        return extractTokenFromDescendant(node, nodeType)
    }
    return ""
}
```

### Integration Points
- Modify `TreeSitterProvider.ToCanonicalNode()` to call token extraction
- Update `NewNode()` calls to include extracted tokens
- Ensure token extraction happens before child processing

## Out of Scope
- Complex token transformation (e.g., string unescaping)
- Token extraction from non-leaf nodes with complex structure
- Performance optimizations for large files (future work)

## Acceptance Criteria
- Token extraction works for all supported extraction strategies
- Existing mappings continue to work without modification
- All new functionality is covered by unit and integration tests
- Token extraction is used successfully in Perl mapping to extract string literals
- Error handling is robust and doesn't break parsing
- Documentation is updated to reflect the new mapping format

## Success Metrics
- Perl language tests pass with proper token extraction
- No regression in existing language support
- Token extraction works for at least 3 different extraction strategies
- Error handling prevents crashes on malformed input

---

*Last updated: 2024-12-19* 