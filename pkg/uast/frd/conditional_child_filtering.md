# Feature Requirements Document: Conditional Child Filtering in UAST Mapping Engine

## Overview

This document describes the requirements for enhancing the UAST mapping engine to support conditional child inclusion/exclusion and field-based name extraction, enabling precise, mapping-driven control over UAST output for complex grammars.

## Motivation

Current mapping logic cannot:
- Exclude unwanted children (e.g., extra `Identifier` nodes in Rust functions)
- Extract names from fields (e.g., `fields.name`)
- Express context-aware child filtering in mapping YAML

This leads to incorrect UASTs for languages with complex grammars, making it impossible to pass language-agnostic tests and to maintain clean, minimal UASTs.

## Requirements

### 1. Conditional Child Filtering
- Allow mapping YAML to specify, for each node type, which children to include or exclude under certain conditions.
- Conditions may include:
  - Child node type
  - Parent node type/context
  - Node properties (fields, roles, etc.)
  - Position in tree (optional, for future)
- Filtering must be context-aware (parent/child relationship).

### 2. Field-based Name Extraction
- Allow mapping YAML to specify that a node's name should be extracted from a specific field (e.g., `fields.name`), not just from properties or text.
- Fallback chain: `fields.name` → `props.name` → node text.

### 3. Mapping YAML Syntax
- Extend the mapping format to support:

```yaml
children:
  - type: identifier
    exclude_if:
      type: identifier
      parent_context: function_item
  - type: ParameterList
  - type: Block
name:
  source: fields.name
```

### 4. Backward Compatibility
- Existing mappings must continue to work unchanged.
- New features must be opt-in via new YAML fields.

### 5. Performance
- The new logic must not introduce more than 5% overhead in parsing time for large files.

## Success Criteria
- Rust and other complex grammars can produce correct UASTs that pass all language-agnostic tests.
- No regressions in other languages.
- Mapping YAML is expressive and maintainable.

## Out of Scope
- Position-based filtering (for now)
- Arbitrary user-defined filter functions (for now)

## Implementation Plan
1. Extend types and loader to support new YAML fields.
2. Implement context-aware child filtering in the mapping engine.
3. Implement field-based name extraction using Tree-sitter field API.
4. Add/adjust tests for Rust and other languages.
5. Validate with all language tests. 