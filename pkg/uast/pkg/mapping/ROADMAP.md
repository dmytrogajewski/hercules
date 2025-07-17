# Mapping System Implementation Roadmap

## Core Modules

- [x] Grammar analysis engine
  - [x] Parse node-types.json from Tree-sitter grammars
  - [x] Extract field dependencies and type relationships
  - [x] Classify nodes by structural patterns (leaf, container, operator)
  - [x] Build inheritance hierarchies from grammar analysis
  - [x] Heuristic classification system
    - [x] Pattern-based classification rules (e.g., *_literal → Literal, *_statement → Statement)
    - [x] Field-based role assignment (e.g., field name "condition" → role "Condition")
  - [x] Automated coverage analysis
    - [x] Calculate % of node types auto-mapped
    - [x] Validate no duplicate mappings or missing critical nodes
  - [x] CLI helper: generate .uastmap from node-types.json analysis

- [x] Grammar analysis engine
  - [x] Parse node-types.json
  - [x] Build node type registry with dependency tracking
  - [x] Apply heuristic classification rules
  - [x] Coverage analysis and validation

- [x] Mapping DSL parser
  - [x] PEG grammar for mapping DSL (Tree-sitter S-expr compatible)
  - [x] Whitespace-tolerant, multi-line/indented support
  - [x] AST-to-MappingRule conversion logic
  - [x] Literal unquoting and robust field extraction
  - [x] Unit tests for parser and conversion

- [x] Pattern matching engine
  - [x] Compile S-expression patterns to Tree-sitter queries
  - [x] Cache compiled queries
  - [x] Match patterns against Tree-sitter nodes (real, tested)
  - [x] Unit tests for matcher and cache
  - [x] Integration with go-tree-sitter-bare and go-sitter-forest
  - [x] Real end-to-end test for pattern matching

- [x] Code generation engine
  - [x] Generate Go code for UAST converters from mapping rules
  - [x] Template-based code generation (stub)
  - [x] Output compatible with node and spec modules
  - [x] Unit test for code generator

## Outstanding TODOs and Feature Gaps

- [x] **Code Generation Data Preparation**
  - [x] Implement logic in `prepareTemplateData` (code_generator.go) to transform mapping rules for code generation templates.
  - [x] Expand code generation templates for real converter output.

- [x] **Mapping API Implementation**
  - [x] Implement `LoadMappings` to parse DSL into rules (mapping.go).
  - [x] Implement `ValidateMappings` for rule consistency and completeness (mapping.go).
  - [x] Implement `GenerateConverters` for the full pipeline (mapping.go).

- [x] **End-to-End and Integration Testing**
  - [x] Implement `TestEndToEndMappingPipeline` to cover the full flow from DSL to generated code (integration_test.go).

- [x] **Feature Extensions**
  - [x] Extend `MappingRule` and `UASTSpec` for advanced mapping features (inheritance, roles, conditional logic) as needed.
  - [x] Add error reporting/diagnostics to mapping and validation.

## Next Steps

- [x] Integration
  - [x] Integrate mapping system with UAST pipeline (replace YAML-based pipeline)
  - [x] Ensure end-to-end compatibility with node and spec modules
  - [x] Provide migration helpers and compatibility shims
  - [x] CLI can generate .uastmap mapping files from grammar analysis

- [ ] End-to-end and integration tests
  - [ ] Test full mapping pipeline on real grammars and mapping DSL
  - [ ] Validate UAST output against spec

- [ ] Documentation
  - [ ] Update/add docs for new mapping system usage and migration
