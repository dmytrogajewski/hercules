# Unified Abstract Syntax Tree (UAST) Specification

## 1. Overview
The Unified Abstract Syntax Tree (UAST) is a canonical, language-agnostic representation of source code, designed for static analysis, refactoring, and cross-language tooling. UAST nodes are produced from Tree-sitter ASTs using a mapping-driven conversion process, and support robust navigation, querying, and transformation via Go APIs and a DSL.

## 2. Canonical Node Structure
Each UAST node is represented by the following Go struct:

```go
type Node struct {
    Id       uint64            // Stable hash for diffing and identity
    Type     string            // Canonical, language-agnostic type (e.g., "FunctionDecl")
    Token    string            // Surface lexeme if leaf; empty otherwise
    Roles    []Role            // Syntactic/semantic labels (e.g., "Function", "Declaration")
    Pos      *Positions        // Optional; byte offsets and line/col info
    Props    map[string]string // Free-form properties for language-specific or extra data
    Children []*Node           // Ordered list of child nodes
}
```

- **Id**: 64-bit FNV hash over Type, Token, and child hashes (for diffing, not required for all workflows).
- **Type**: Set via mapping from Tree-sitter node kind; must be language-agnostic and stable.
- **Token**: For leaf nodes, the surface text; empty for non-leaves.
- **Roles**: List of orthogonal labels (e.g., "Name", "Call", "Assignment").
- **Pos**: Optional struct with start/end byte offsets and line/col info.
- **Props**: Arbitrary key-value pairs for language-specific or extra data (e.g., function name, docstring).
- **Children**: Ordered list of child nodes.

### Positions struct
```go
type Positions struct {
    StartLine   int
    StartCol    int
    StartOffset int
    EndLine     int
    EndCol      int
    EndOffset   int
}
```

### Role type
```go
type Role string // extensible, e.g., "Function", "Declaration", "Name"
```

## 2.1 Canonical UAST Types

The following is the recommended set of canonical, language-agnostic UAST types. All language mappings should use these types where applicable. New types may be added as needed for language-specific constructs, but these form the core cross-language vocabulary:

| Type              | Description                                      |
|-------------------|--------------------------------------------------|
| File              | Root node for a source file                      |
| Function          | Function or method declaration                   |
| FunctionDecl      | Function declaration (alias for Function)        |
| Method            | Method declaration (object-oriented)             |
| Class             | Class declaration                                |
| Interface         | Interface declaration                            |
| Struct            | Struct/record declaration                        |
| Enum              | Enum declaration                                 |
| EnumMember        | Enum member/variant                              |
| Variable          | Variable declaration                             |
| Parameter         | Function/method parameter                        |
| Block             | Block of statements                              |
| If                | If statement                                     |
| For               | For/while loop                                   |
| While             | While loop                                       |
| DoWhile           | Do-while loop                                    |
| Switch            | Switch/case statement                            |
| Case              | Case/branch in switch                            |
| Return            | Return statement                                 |
| Break             | Break statement                                  |
| Continue          | Continue statement                               |
| Assignment        | Assignment statement                             |
| Call              | Function/method call                             |
| Identifier        | Identifier (variable, function, etc.)            |
| Literal           | Literal value (string, number, etc.)             |
| BinaryOp          | Binary operator (e.g., +, -, *, /, ==, etc.)     |
| UnaryOp           | Unary operator (e.g., !, -, ++)                  |
| Import            | Import/include statement                         |
| Package           | Package/module declaration                       |
| Attribute         | Attribute/annotation (e.g., @Override)           |
| Comment           | Comment node                                     |
| DocString         | Documentation string                             |
| TypeAnnotation    | Type annotation                                  |
| Field             | Field/member variable declaration                |
| Property          | Property (object-oriented)                       |
| Getter            | Getter method                                    |
| Setter            | Setter method                                    |
| Lambda            | Lambda/anonymous function                        |
| Try               | Try/catch/finally block                          |
| Catch             | Catch block                                      |
| Finally           | Finally block                                    |
| Throw             | Throw/raise statement                            |
| Module            | Module declaration                               |
| Namespace         | Namespace declaration                            |
| Decorator         | Decorator (Python, TypeScript, etc.)             |
| Spread            | Spread/rest element (e.g., ...args)              |
| Tuple             | Tuple expression                                 |
| List              | List/array expression                            |
| Dict              | Dictionary/map expression                        |
| Set               | Set expression                                   |
| KeyValue          | Key-value pair                                   |
| Index             | Indexing operation (e.g., arr[0])                |
| Slice             | Slice operation (e.g., arr[1:3])                 |
| Cast              | Type cast                                        |
| Await             | Await expression                                 |
| Yield             | Yield expression                                 |
| Generator         | Generator function                               |
| Comprehension     | List/dict/set comprehension                      |
| Pattern           | Pattern matching construct                       |
| Match             | Match statement (e.g., Python 3.10+)             |

**Note:**
- Types should be used in a language-agnostic way (e.g., both Java and Go use `Function` for function declarations).
- For language-specific constructs, use a clear, descriptive type and document it in the mapping file.
- Types may be prefixed with the language (e.g., `go:file`) for fallback or debugging, but canonical types should be preferred in mappings.

## 2.2 Canonical UAST Roles

The following is the recommended set of canonical, language-agnostic UAST roles. Roles annotate nodes with orthogonal syntactic or semantic properties, enabling richer queries and analyses. All language mappings should use these roles where applicable. New roles may be added as needed for language-specific constructs, but these form the core cross-language vocabulary:

| Role         | Description                                                      |
|--------------|------------------------------------------------------------------|
| Function     | Node represents a function or method                             |
| Declaration  | Node is a declaration (function, variable, class, etc.)          |
| Name         | Node is a name/identifier (of a variable, function, etc.)        |
| Reference    | Node is a reference to a declared entity                         |
| Assignment   | Node is an assignment target or statement                        |
| Call         | Node is a function/method call                                   |
| Parameter    | Node is a function/method parameter                              |
| Argument     | Node is an argument in a call                                    |
| Condition    | Node is a condition (e.g., in if/while)                          |
| Body         | Node is a body/block of statements                               |
| Exported     | Node is exported/public (as per language rules)                  |
| Public       | Node is explicitly public                                        |
| Private      | Node is explicitly private                                       |
| Static       | Node is static (class-level, not instance)                       |
| Constant     | Node is a constant definition                                    |
| Mutable      | Node is mutable/assignable                                       |
| Getter       | Node is a getter method/property                                 |
| Setter       | Node is a setter method/property                                 |
| Loop         | Node is a loop construct (for, while, etc.)                      |
| Branch       | Node is a branch/case in a switch/match                          |
| Import       | Node is an import/include statement                              |
| Doc          | Node is a documentation string                                   |
| Comment      | Node is a comment                                                |
| Attribute    | Node is an attribute/annotation/decorator                        |
| Annotation   | Node is an annotation (Java, etc.)                               |
| Operator     | Node is an operator (binary/unary)                               |
| Index        | Node is an index (e.g., arr[0])                                  |
| Key          | Node is a key in a key-value pair                                |
| Value        | Node is a value in a key-value pair                              |
| Type         | Node is a type annotation or reference                           |
| Interface    | Node is an interface declaration                                 |
| Class        | Node is a class declaration                                      |
| Struct       | Node is a struct/record declaration                              |
| Enum         | Node is an enum declaration                                      |
| Member       | Node is a member/field/property                                  |
| Module       | Node is a module/package/namespace declaration                   |
| Lambda       | Node is a lambda/anonymous function                              |
| Try          | Node is a try/catch/finally block                                |
| Catch        | Node is a catch block                                            |
| Finally      | Node is a finally block                                          |
| Throw        | Node is a throw/raise statement                                  |
| Await        | Node is an await expression                                      |
| Yield        | Node is a yield expression                                       |
| Spread       | Node is a spread/rest element (e.g., ...args)                    |
| Pattern      | Node is a pattern in pattern matching                            |
| Match        | Node is a match statement (e.g., Python 3.10+)                   |

**Note:**
- Roles are orthogonal to types: a node may have multiple roles (e.g., `Name`, `Parameter`, `Exported`).
- For language-specific roles, use a clear, descriptive name and document it in the mapping file.
- Roles should be used to enable precise, cross-language queries and analyses.

## 3. Mapping and Conversion

### 3.1 Mapping Files
- Each language has a YAML/JSON mapping file (e.g., `go.yaml`) that maps Tree-sitter node kinds to UAST types, roles, and optional property extraction rules.
- Example:
  ```yaml
  language: go
  mapping:
    function_declaration:
      type: FunctionDecl
      roles: [Function, Declaration]
    identifier:
      type: Identifier
      roles: [Name]
  ```
- The mapping is loaded at startup and used for all conversions.

### 3.2 Conversion Algorithm
- Parse source code with Tree-sitter to obtain a parse tree.
- Traverse the tree (explicit stack, not recursion).
- For each Tree-sitter node:
  - Look up its kind in the mapping.
  - Set `Type` and `Roles` from the mapping.
  - For mapped properties (e.g., function name), extract from child nodes with the specified role (e.g., child with `roles: [Name]`).
  - Set `Token` for leaves (surface text).
  - Set `Pos` from byte/line/col info.
  - Set `Props` for extra data.
  - Recursively build `Children`.
- If a node kind is not mapped, use a fallback type (e.g., `lang:kind`).

## 4. Serialization
- **JSON**: All fields are serialized, omitting nil/empty fields where idiomatic.
- **Protobuf**: A matching proto schema is defined for performance and gRPC; not required for all workflows.

## 5. Extensibility
- The node struct, mapping format, and conversion logic are designed for future extension:
  - New fields, roles, or properties can be added without breaking existing code.
  - New languages or node kinds are supported by editing/adding mapping files only.
  - Hot-reloading of mapping files is supported in server mode.

## 6. Mapping-Driven Construction
- All type, role, and property assignment is driven by the mapping file for the language.
- No hardcoded type/role logic in the converter; all language-specific details are declarative.
- Properties (e.g., function name) are extracted according to mapping rules (e.g., child with `roles: [Name]`).

## 7. Requirements and Guarantees
- All nodes have a stable, language-agnostic `Type`.
- All mapped roles are present in `Roles`.
- All mapped properties are present in `Props`.
- All children are ordered as in the source AST.
- All code is non-recursive (explicit stack for traversal).
- All public APIs and mapping formats are documented.

## 8. Example
Given Go code:
```go
func add(a, b int) int { return a + b }
```
With mapping:
```yaml
function_declaration:
  type: FunctionDecl
  roles: [Function, Declaration]
identifier:
  type: Identifier
  roles: [Name]
```
The UAST node for the function will be:
```json
{
  "Type": "FunctionDecl",
  "Roles": ["Function", "Declaration"],
  "Props": {"name": "add"},
  "Children": [ ... ]
}
```

---

*This document is the authoritative specification for the UAST format in this project. All code, mappings, and tests must conform to this spec.* 