# UAST - Unified Abstract Syntax Tree

[![Go Report Card](https://goreportcard.com/badge/github.com/dmytrogajewski/hercules)](https://goreportcard.com/report/github.com/dmytrogajewski/hercules)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dmytrogajewski/hercules)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)

A Go-native **Unified AST (UAST)** data model backed by Tree-sitter parsers plus a compact **domain-specific language (DSL)** for querying and transforming trees. Parse, analyze, and refactor code written in 66+ languages with one toolkit.

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [What is UAST?](#what-is-uast)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Basic Parsing](#basic-parsing)
  - [DSL Queries](#dsl-queries)
  - [Go API](#go-api)
  - [CLI Tool](#cli-tool)
- [Language Support](#language-support)
- [Performance](#performance)
- [Contributing](#contributing)
- [Documentation](#documentation)
- [License](#license)

## 🚀 Quick Start

```bash
# Install the CLI tool
go install github.com/dmytrogajewski/hercules/cmd/uast@latest

# Parse a Go file and find all functions
uast parse main.go | uast query 'filter(.type == "Function")'

# Parse Python and find function calls
uast parse -lang python script.py | uast query 'filter(.type == "Call")'
```

## 🤔 What is UAST?

UAST (Unified Abstract Syntax Tree) provides a language-agnostic representation of source code. Instead of dealing with different AST formats for each programming language, UAST gives you a single, consistent structure for analyzing code across 100+ languages.

### How it works:

```
Source Code → Tree-sitter Parser → Mapping-driven Conversion → UAST → DSL Queries → Analysis
```

## ✨ Features

- **🌍 Multi-language Support**: Parse 66+ programming languages with Tree-sitter grammars
- **🔍 Powerful DSL**: Query and filter nodes with a functional pipeline syntax
- **⚡ High Performance**: Optimized for speed with streaming iterators and memory pools
- **🛠️ Go-native API**: Ergonomic Go APIs for navigation, mutation, and transformation
- **📊 Change Detection**: Language-agnostic diffing and change analysis
- **🎯 Mapping-driven**: DSL-based configuration for language-specific conversions

## 📦 Installation

### Prerequisites

- Go 1.22 or later
- Git

### Install CLI Tool

```bash
go install github.com/dmytrogajewski/hercules/cmd/uast@latest
```

### Use as Library

```bash
go get github.com/dmytrogajewski/hercules/pkg/uast
```

## 📖 Usage

### Basic Parsing

```go
package main

import (
    "fmt"
    "log"
    "github.com/dmytrogajewski/hercules/pkg/uast"
)

func main() {
    // Create parser
    parser, err := uast.NewParser()
    if err != nil {
        log.Fatal(err)
    }

    // Parse Go code
    code := []byte(`package main
func hello() {
    fmt.Println("Hello, World!")
}`)

    node, err := parser.Parse("main.go", code)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Parsed %s with %d children\n", node.Type, len(node.Children))
}
```

### DSL Queries

The UAST DSL provides a functional pipeline syntax for querying nodes:

```go
// Find all exported functions
nodes, err := node.FindDSL("filter(.type == \"Function\" && .roles has \"Exported\")")
if err != nil {
    log.Fatal(err)
}

// Count all string literals
nodes, err := node.FindDSL("filter(.type == \"Literal\") |> reduce(count)")
if err != nil {
    log.Fatal(err)
}

// Find function calls with specific names
nodes, err := node.FindDSL("filter(.type == \"Call\" && .props.name == \"printf\")")
if err != nil {
    log.Fatal(err)
}
```

**Supported DSL Operations:**
- **Filtering**: `filter(.type == "Function")`
- **Boolean Logic**: `&&`, `||`
- **Equality**: `==`, `!=`
- **Membership**: `.roles has "Exported"`
- **Field Access**: `.token`, `.type`, `.props.name`
- **Pipelines**: `|>` for chaining operations

### Go API

#### Navigation and Querying

```go
// Streaming pre-order iterator
iter := node.PreOrder()
for node := range iter {
    if node.HasRole("RoleName") {
        // process identifier
    }
}

// Find nodes with predicate
functions := node.Find(func(n *uast.Node) bool {
    return n.Type == "Function"
})
```

#### Transformation

```go
// Transform nodes in-place
node.TransformInPlace(func(n *uast.Node) bool {
    if node.HasRole(n, uast.RoleString) {
        n.Token = strings.Trim(n.Token, "\"")
    }
    return true
})
```

#### Change Detection

```go
// Detect structural changes between two versions
changes := uast.DetectChanges(before, after)
for _, change := range changes {
    fmt.Printf("%s: %s\n", change.Type, change.File)
}
```

### CLI Tool

The UAST CLI provides command-line access to all features:

```bash
# Parse a file and output UAST as JSON
uast parse main.go

# Query UAST using DSL
uast parse main.go | uast query 'filter(.type == "Function")'

# Format UAST output
uast parse main.go | uast fmt

# Detect changes between files
uast diff before.go after.go

# Get help
uast --help
```

## 🌍 Language Support

UAST supports 66+ programming languages including:

**Popular Languages:**
- Go, Python, Java, JavaScript, TypeScript
- Rust, C++, C#, Ruby, PHP, Kotlin, Swift

**Web Technologies:**
- HTML, CSS, JSON, YAML, XML, Markdown

**Configuration Files:**
- Dockerfile, Makefile, CMake, TOML, INI

**Specialized Languages:**
- SQL, Haskell, OCaml, Scala, Elixir, Erlang
- F#, Clojure, Lua, Perl
- And 50+ more languages

See the [language roadmap](LANGUAGE_ROADMAP.md) for the complete list and status.

## ⚡ Performance

UAST is optimized for high-performance code analysis:

### Parsing Performance
- **Small files (~50 lines):** ~32μs, 6KB memory
- **Medium files (~100 lines):** ~270μs, 57KB memory
- **Large files (~200 lines):** ~1ms, 208KB memory

### DSL Query Performance
- **Simple field access:** ~1.9μs, 2.6KB memory
- **Filter operations:** ~4.5μs, 5.7KB memory
- **Complex pipelines:** ~10μs, 12KB memory

### Tree Traversal
- **Pre-order streaming:** ~18μs, 384B memory
- **Find with predicate:** ~0.7μs, 248B memory

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on:

- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Development Setup](DEVELOPMENT.md)
- [Adding New Languages](docs/ADDING_LANGUAGE.md)
- [Reporting Issues](../../issues)
- [Submitting Pull Requests](../../pulls)

### Quick Contribution

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📚 Documentation

- **[API Reference](docs/API.md)**: Complete Go API documentation
- **[DSL Syntax](docs/DSL_SYNTAX.md)**: Query language reference
- **[Language Mapping](docs/MAPPING_FORMAT.md)**: How to add new languages
- **[Configuration](docs/CONFIGURATION.md)**: Setup and configuration options
- **[Deployment](docs/DEPLOYMENT.md)**: Production deployment guide
- **[Plugins](docs/PLUGINS.md)**: Extending UAST with plugins
- **[Recipes](docs/RECIPES.md)**: Common use cases and examples

## 📄 License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## 🙏 Acknowledgments

- [Tree-sitter](https://tree-sitter.github.io/tree-sitter/) for the parsing foundation
- [go-sitter-forest](https://github.com/alexaandru/go-sitter-forest) for Go bindings
- [pointlander/peg](https://github.com/pointlander/peg) for DSL parsing
- All our [contributors](../../graphs/contributors) who make this project possible

---

**Ready to start analyzing code across languages?** [Get started with the Quick Start](#quick-start) or explore the [API Reference](docs/API.md).
