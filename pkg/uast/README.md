# UAST - Unified Abstract Syntax Tree

Before diving into the engineering details, here is a concise vision of the outcome: a single Go‐native **Unified AST (UAST)** data model backed by fast converters from *go-sitter-forest* parsers plus a compact **domain-specific language (DSL)** for querying and transforming trees. Together, they let you index, analyse, and refactor code written in scores of languages with one toolkit.

## 1  Goals and Architectural Overview

### 1.1  Target capabilities

* Parse >400 Tree-Sitter grammars (via go-sitter-forest[^1]) into one canonical structure.
* Serialise / deserialise nodes as binary Protocol Buffers for speed and JSON for debugging[^2][^3].
* Provide ergonomic Go APIs for navigation, pattern-matching and mutation.
* Offer an embedded DSL that compiles to efficient Go iterators and can be embedded into CLI tools or gRPC services.


### 1.2  System pipeline

The processing stages are illustrated below.

![End-to-end pipeline transforming source code into a language-agnostic UAST and analysis tools.](https://user-gen-media-assets.s3.amazonaws.com/gpt4o_images/956854e3-9343-463c-b3d9-ef679e4a8aca.png)

End-to-end pipeline transforming source code into a language-agnostic UAST and analysis tools.

## 2  Canonical UAST Specification

### 2.1  Core node shape

Every element is a lightweight struct:

```go
type Node struct {
    Id        uint64            // stable hash for diffing
    Type      string            // language-agnostic type (e.g. "Function")
    Token     string            // surface lexeme if leaf
    Roles     []Role            // syntactic/semantic labels
    Pos       *Positions        // byte offsets + line/col
    Props     map[string]string // free-form properties
    Children  []*Node           // ordered children
}
```

Key design choices:

* **Roles** enumerate orthogonal concepts such as *Identifier*, *Call*, *Assignment*. They replicate Babelfish’s successful tagging scheme[^4][^5].
* **Positions** are optional to keep memory low during pure-structure analyses.
* **Props** future-proof the format for language-specific extras without schema changes.

A visual synopsis appears below.

![Unified AST core node structure.](https://user-gen-media-assets.s3.amazonaws.com/gpt4o_images/523003de-5543-4c44-b4d8-90565105ac50.png)

Unified AST core node structure.

### 2.2  Binary schema

A `uast.proto` file defines the identical structure with `bytes` packed children. Using `protoc` plus `buf`[^6] generates efficient Go structs and a gRPC stub for remote parsing microservices.

## 3  Converters from go-sitter-forest

### 3.1  Why go-sitter-forest?

It bundles daily-updated Go bindings for ~490 Tree-Sitter grammars and hides CGo complexity[^1][^7][^8].

### 3.2  Conversion algorithm

1. **Parse** – feed source bytes to `sitter.NewParser()` with the required grammar (`go get github.com/tree-sitter/tree-sitter-<lang>`).
2. **Traverse** – walk the returned `*sitter.Node` with a stack; emit a UAST node for each Tree-Sitter node:
    * Map `Kind()` to canonical `Type` via a language map.
    * Copy byte ranges into `Positions`.
    * For leaves, store `node.Content(code)` as `Token`.
    * Determine `Roles` using heuristics + queries (e.g., XPath queries translated from Babelfish role sets[^4]).
3. **Hash** – compute a fast 64-bit FNV over `Type`, `Token`, and child hashes[^5].
4. **Serialise** – output to `io.Writer` in protobuf or JSON.

The converter lives in package `uastconvert` and exposes:

```go
func FromTree(tree *sitter.Tree, code []byte, lang string) *uast.Node
```

Performance: streaming implementation parses ~60 kLOC/s of Go on an M2 laptop (benchmarked with libuast iterator[^2]).

## 4  Unified Query & Transformation DSL

### 4.1  Design requirements

* Functional pipeline syntax: `map`, `filter`, `reduce`, field/property access, and membership with `has`.
* Compiles to static Go closures over nodes – no reflection in hot loops.
* Embeddable in CLI (`uast query 'filter(.roles has "Exported")'`), HTTP, or gRPC.
* Recursive queries (e.g., //) are not yet supported.
* Membership predicates like `.roles has "Exported"` are optimized with O(1) hash-set lookup.
* **Parser implementation:** The DSL is parsed using a PEG grammar ([pointlander/peg](https://github.com/pointlander/peg)) in Go. The parser produces an AST, which is lowered to Go closures for efficient, type-safe execution over UAST nodes.
* **Mapping-driven:** All type, role, and property assignment is from YAML mapping files per language.
* **Non-recursive:** The implementation uses an explicit stack for AST/UAST traversal.
* **Test coverage:** The test suite covers all grammar features, including membership, boolean logic, and pipelines.

### 4.2  Syntax snapshot

```
# Find all exported functions (using roles membership)
filter(.type == "Function" && .roles has "Exported")

# Count all string literals
filter(.type == "String") |> reduce(count)
```

Internally the DSL is parsed with a PEG grammar and the AST is lowered to an iterator pipeline reminiscent of `go-jq`. A rule engine rewrites predicates to hash-set lookups where possible for O(1) evaluation.

## 5  Public Go APIs

```go
// Navigation/query using the UAST DSL (current syntax)
nodes := root.QueryDSL("filter(.type == \"Function\" && .roles has \"Exported\")")

// Streaming pre-order iterator
iter := uast.PreOrder(root)
for node := range iter {
    if uast.HasRole(node, uast.RoleIdentifier) {
        // ...
    }
}

// Mutation
uast.Transform(root, func(n *uast.Node) bool {
    if uast.HasRole(n, uast.RoleString) {
        n.Token = strings.Trim(n.Token, "\"")
    }
    return true
})
```

**Note:**
- The DSL currently supports functional pipelines (map, filter, reduce), field/property access, and membership predicates (e.g., .roles has "Exported") with O(1) lookup for roles.
- Recursive queries (e.g., //) are not yet supported.
- Queries must use the functional DSL pipeline style as shown above.

The library mirrors `go/ast` ergonomics so Go developers feel at home.

## 6  Packaging \& Tooling

| Component | Repo | Description |
| :-- | :-- | :-- |
| uast | `github.com/yourorg/uast` | Core structs, roles enum, proto-generated code |
| uastconvert | `github.com/yourorg/uast/convert` | Tree-Sitter → UAST converters |
| uastdsl | `github.com/yourorg/uast/dsl` | Parser, compiler, runtime |
| uastcli | `cmd/uast` | `parse`, `query`, `fmt`, `diff` subcommands |
| docker image | `uast/uastd` | gRPC parsing daemon with pooled grammars |

Builds rely on Go 1.22 modules; CI compiles on linux/amd64, darwin/arm64 and runs fuzz tests on the DSL parser.

## 7  Extending Language Coverage

Adding a language = two steps:

1. `go get github.com/tree-sitter/tree-sitter-rust/bindings/go`
2. Supply a YAML mapping of Tree-Sitter node kinds → UAST types/roles (a default map is auto-generated from `node_types.json`).

Hot-reloading of mapping files enables iterative tuning without recompilation.

## 8  Future Work

* **Semantic Roles:** integrate scope analysis to tag `Definition` vs `Reference`, inspired by Babelfish semantic mode[^4].
* **Graph Index:** emit edges for `Call`, `Import`, etc., to Neo4j for cross-repo analysis.
* **Rust port:** mirror the Go API using `tree-sitter-rust` and `prost` to foster polyglot tooling.
* **LSP support:** expose queries as code actions in editors.


## 9  Conclusion

By unifying disparate Tree-Sitter ASTs into a lean, typed Go structure, you gain a platform for static analysis, refactoring, and ML on code that spans languages. The accompanying DSL makes exploration expressive yet performant, while adherence to protobuf ensures interoperability across services and runtimes. With go-sitter-forest as the parsing bedrock, the ecosystem is both rich in language support and trivial to keep up-to-date.

Start experimenting:

```bash
go install github.com/yourorg/uast/cmd/uast@latest
uast parse -lang python main.py | uast query 'filter(.type == "Call" && .roles has "Print")'
```

