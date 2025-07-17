# Tree-Sitter to UAST Mapping System: Core Technical Specification

## 1 Core Problem and Solution Architecture

### 1.1 Problem Statement

Convert Tree-Sitter concrete syntax trees to unified abstract syntax trees with minimal manual effort while maintaining O(1) dispatch performance. Each Tree-Sitter grammar contains 200-500 node types that must be mapped to canonical UAST representations.

### 1.2 Solution Architecture

**Three-phase pipeline**: Grammar introspection → Heuristic mapping generation → Code compilation

```
Phase 1: Grammar Analysis
├── Input: node-types.json from Tree-Sitter grammar
├── Process: Extract node metadata, field information, structural patterns
└── Output: NodeTypeMetadata[] with classification data

Phase 2: Mapping Generation  
├── Input: NodeTypeMetadata[] + heuristic rules + manual overrides
├── Process: Apply pattern matching, role assignment, inheritance resolution
└── Output: MappingRule[] with Tree-Sitter patterns and UAST specifications

Phase 3: Code Generation
├── Input: MappingRule[] 
├── Process: Compile to Go switch statement with inlined converters
└── Output: High-performance converter with O(1) dispatch
```


## 2 Grammar Analysis Engine

### 2.1 Node-Types.json Processing

Tree-Sitter grammars generate `node-types.json` containing structural metadata:

```json
{
  "type": "function_declaration",
  "named": true,
  "fields": {
    "name": {"required": true, "types": [{"type": "identifier", "named": true}]},
    "body": {"required": true, "types": [{"type": "block", "named": true}]},
    "parameters": {"required": false, "types": [{"type": "parameter_list", "named": true}]}
  }
}
```

**Processing Algorithm**:

1. Parse JSON into structured metadata
2. Extract field dependencies and type relationships
3. Classify nodes by structural patterns (leaf, container, operator)
4. Build inheritance hierarchies from grammar analysis

### 2.2 Heuristic Classification System

**Pattern-based classification rules**:


| Pattern | UAST Type | Roles | Confidence |
| :-- | :-- | :-- | :-- |
| `*_literal` | `"Literal"` | `["Literal"]` | 95% |
| `identifier` | `"Identifier"` | `["Identifier"]` | 98% |
| `*_statement` | Auto-detect | `["Statement"]` | 80% |
| `*_declaration` | Auto-detect | `["Declaration"]` | 85% |
| `*_expression` | Auto-detect | `["Expression"]` | 75% |

**Field-based role assignment**:

- Field name `"condition"` → Role `"Condition"`
- Field name `"body"` → Role `"Body"`
- Field name `"name"` → Role `"Name"`
- Field name `"arguments"` → Role `"Arguments"`


### 2.3 Coverage Analysis

**Automated coverage calculation**:

- Target: 95% of node types auto-mapped
- Measurement: Count mapped vs unmapped nodes
- Validation: Ensure no duplicate mappings or missing critical nodes


## 3 Pattern Matching System

### 3.1 Tree-Sitter Query Integration

**S-expression pattern syntax**:

```
(function_declaration
  name: (identifier) @name
  parameters: (parameter_list) @params
  body: (block) @body)
```

**Query compilation process**:

1. Parse S-expression into query AST
2. Compile to `*sitter.Query` object
3. Cache compiled queries for reuse
4. Execute with `QueryCursor.Matches()`

### 3.2 Capture System

**Capture naming and reference**:

- Captures use `@name` syntax in patterns
- Referenced in UAST construction as `"@name"`
- Validated at compile time for existence
- Support for optional captures with `?` modifier

**Capture resolution algorithm**:

```
1. Parse pattern and extract capture names
2. Validate all captures exist in UAST mapping
3. Generate capture-to-field binding code
4. Compile to direct field access (no reflection)
```


## 4 Mapping DSL Implementation

### 4.1 PEG Grammar Specification

**Using pointlander/peg implementation**:

```peg
Grammar     <- Rule+
Rule        <- Identifier '<-' Pattern '=>' UastSpec
Pattern     <- TSQuery
UastSpec    <- 'uast' '(' UastField+ ')'
UastField   <- FieldName ':' (CaptureRef / Literal)
CaptureRef  <- '@' Identifier
TSQuery     <- '(' NodeType FieldSpec* ')' CaptureSpec*
FieldSpec   <- FieldName ':' Pattern
CaptureSpec <- '@' Identifier
```


### 4.2 DSL Syntax Examples

**Basic mapping**:

```
function_declaration <- (function_declaration 
                        name: (identifier) @name
                        body: (block) @body) => uast(
    type: "Function",
    token: "@name", 
    roles: ["Declaration"],
    children: ["@body"]
)
```

**Inheritance and overrides**:

```
method_declaration <- extends(function_declaration) => uast(
    type: "Function",
    token: "@name",
    roles: ["Declaration", "Method"],
    props: ["receiver": "true"]
)
```


### 4.3 Semantic Actions

**Compile-time validation**:

- Verify capture references exist in pattern
- Check UAST type/role enum values
- Validate field type compatibility
- Ensure no circular dependencies


## 5 Code Generation Architecture

### 5.1 Switch Statement Generation

**Target structure**:

```go
func convertNode(node *sitter.Node, source []byte) *uast.Node {
    switch node.Type() {
    case "function_declaration":
        return convertFunctionDeclaration(node, source)
    case "identifier":
        return convertIdentifier(node, source)
    default:
        return createGenericNode(node, source)
    }
}
```


### 5.2 Converter Function Generation

**Per-rule converter template**:

```go
func convertFunctionDeclaration(node *sitter.Node, source []byte) *uast.Node {
    query := cachedQueries["function_declaration"]
    matches := executeQuery(query, node, source)
    
    if len(matches) == 0 {
        return createGenericNode(node, source)
    }
    
    captures := matches[^0].Captures
    return &uast.Node{
        Type:  "Function",
        Token: extractToken(captures["name"], source),
        Roles: []uast.Role{uast.Declaration},
        Children: []*uast.Node{
            convertNode(captures["body"], source),
        },
        Id: generateHash(node),
    }
}
```


### 5.3 Performance Optimizations

**Compile-time optimizations**:

- Inline simple conversions
- Pre-compile all Tree-Sitter queries
- Generate typed capture accessors
- Eliminate reflection in hot paths

**Runtime optimizations**:

- Query result caching
- Memory pool for UAST nodes
- Single-pass traversal
- Minimal allocations


## 6 Implementation Guidelines

### 6.1 Grammar Analysis Implementation

**Steps**:

1. Parse `node-types.json` using `encoding/json`
2. Build node type registry with dependency tracking
3. Apply heuristic classification rules
4. Generate mapping file templates with coverage analysis
5. Validate completeness and consistency

**Data structures**:

```go
type NodeTypeInfo struct {
    Name     string
    IsNamed  bool
    Fields   map[string]FieldInfo
    Children []ChildInfo
    Category NodeCategory  // Leaf, Container, Operator
}

type FieldInfo struct {
    Name     string
    Required bool
    Types    []string
    Multiple bool
}
```


### 6.2 PEG Parser Implementation

**Using pointlander/peg**:

1. Define grammar in `.peg` file
2. Generate parser with `peg -inline -switch mapping.peg`
3. Implement semantic actions for AST construction
4. Add validation and error reporting

**Parser integration**:

```go
type MappingParser struct {
    rules map[string]*Rule
    errors []ParseError
}

func (p *MappingParser) ParseMapping(input string) (*Mapping, error) {
    // Generated PEG parsing code
    // Returns validated mapping rules
}
```


### 6.3 Code Generation Implementation

**Template-based generation**:

1. Load mapping rules from parsed DSL
2. Generate switch statement with all node types
3. Create converter functions for each rule
4. Add imports and package declarations
5. Format with `go fmt` and validate with `go vet`

**Generation pipeline**:

```go
type CodeGenerator struct {
    templates map[string]*template.Template
    rules     []MappingRule
    language  string
}

func (g *CodeGenerator) GenerateConverter() ([]byte, error) {
    data := struct {
        Language string
        Rules    []MappingRule
    }{g.language, g.rules}
    
    return g.executeTemplate("converter.go.tmpl", data)
}
```


## 7 Performance Specifications

### 7.1 Conversion Performance

**Targets**:

- 50,000+ lines/second/core
- <100MB memory for 1GB source
- <5% overhead vs direct Tree-Sitter parsing
- Linear scaling with input size


### 7.2 Compilation Performance

**Targets**:

- <5 seconds for 1000 mapping rules
- <100ms converter startup time
- <1MB memory for compiled converter
- Incremental compilation support


### 7.3 Memory Management

**Strategies**:

- Object pooling for frequent allocations
- Streaming processing for large files
- Lazy evaluation of expensive operations
- Memory-mapped files for large grammars


## 8 Testing and Validation

### 8.1 Test Categories

**Unit tests**:

- Grammar analysis accuracy
- Pattern matching correctness
- Code generation validity
- Performance benchmarks

**Integration tests**:

- End-to-end mapping pipeline
- Multi-language support
- Error handling and recovery
- Memory usage validation


### 8.2 Validation Framework

**Automated validation**:

- Coverage analysis (100% node types mapped)
- Type safety verification
- Performance regression detection
- Memory leak detection

**Quality metrics**:

- Mapping accuracy (>99%)
- Conversion speed (target thresholds)
- Memory efficiency (target limits)
- Error rate (<0.1%)

This specification provides the core technical details needed to implement a high-performance Tree-Sitter to UAST mapping system. The focus is on concrete algorithms, data structures, and implementation patterns rather than high-level concepts.

