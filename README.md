# Hercules

> **Disclaimer:**
> This project originates from [src-d/hercules](https://github.com/src-d/hercules), but follows a **completely different concept and architecture**. It is not a drop-in replacement and is maintained independently.

[![CI](https://github.com/dmytrogajewski/hercules/actions/workflows/ci.yml/badge.svg)](https://github.com/dmytrogajewski/hercules/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmytrogajewski/hercules.svg)](https://pkg.go.dev/github.com/dmytrogajewski/hercules)

Hercules is a comprehensive code analysis platform providing both static analysis (via UAST) and git repository history analysis.

---

## Features

- **Static Analysis**: Complexity, cohesion, comments, Halstead metrics, imports extraction
- **History Analysis**: Burndown, couples, developer stats, file history, sentiment, shotness, typos
- **Multi-Language UAST**: Universal Abstract Syntax Tree supporting 60+ languages via tree-sitter
- **Flexible Output**: Text, JSON, YAML formats
- **Extensible**: Add new analyzers easily

---

## Quick Start

### Installation

```sh
go install github.com/dmytrogajewski/hercules/cmd/herr@latest
go install github.com/dmytrogajewski/hercules/cmd/uast@latest
```

Or build from source:

```sh
git clone https://github.com/dmytrogajewski/hercules.git
cd hercules
make build
```

---

## Command-Line Tools

Hercules provides two CLI tools:

| Tool | Purpose |
|------|---------|
| `uast` | Parse source code into Universal Abstract Syntax Tree |
| `herr` | Analyze code (static analysis and git history) |

### Static Analysis

Analyze source code structure and complexity:

```sh
# Analyze a single file with all analyzers
uast parse main.go | herr analyze

# Specific analyzers
uast parse main.go | herr analyze -a complexity,halstead

# JSON output
uast parse main.go | herr analyze -f json

# Analyze multiple files
uast parse *.go | herr analyze
```

**Available static analyzers:**

| Analyzer | Description |
|----------|-------------|
| `complexity` | Cyclomatic and cognitive complexity |
| `cohesion` | Code cohesion metrics |
| `comments` | Comment quality and documentation coverage |
| `halstead` | Halstead complexity measures |
| `imports` | Import statement extraction |

### History Analysis

Analyze git repository history:

```sh
# Burndown analysis
herr history -a burndown .

# Multiple analyzers
herr history -a burndown,couples,devs .

# Analyze only latest commit
herr history -a devs --head .

# JSON output
herr history -a burndown -f json .
```

**Available history analyzers:**

| Analyzer | Description |
|----------|-------------|
| `burndown` | Line burndown statistics over time |
| `couples` | File/developer coupling matrix |
| `devs` | Developer activity (commits, lines added/removed) |
| `file-history` | File change history with contributors |
| `imports` | Import usage per developer over time |
| `sentiment` | Comment sentiment classification |
| `shotness` | Structural hotspots (fine-grained coupling) |
| `typos` | Typo-fix pairs extracted from diffs |

### UAST Operations

```sh
# Parse a file to UAST
uast parse main.go

# Parse with JSON output
uast parse main.go -f json

# Validate UAST mapping
uast validate go

# Start development server
uast server --port 8080
```

---

## Examples

### Complexity Analysis

```sh
$ uast parse pkg/uast/parser.go | herr analyze -a complexity

=== COMPLEXITY ===
High complexity - functions are complex and should be refactored
nesting_depth: 24.00 | total_complexity: 15.00 | total_functions: 10.00
```

### Developer Statistics

```sh
$ herr history -a devs --head .

hercules (v2):
  version: 0
Devs:
  ticks:
    0:
      -1: [1, 391669, 0, 0, {Go: [325974, 0, 0], JavaScript: [1757, 0, 0], ...}]
  tick_size: 86400
```

### Import Analysis

```sh
$ uast parse main.go | herr analyze -a imports -f json

{
  "imports": {
    "count": 6,
    "imports": ["fmt", "os", "strings", ...]
  }
}
```

---

## Project Structure

```
hercules/
├── cmd/
│   ├── herr/             # Main analyzer CLI
│   └── uast/             # UAST parser CLI
├── pkg/
│   ├── analyzers/        # All analyzers
│   │   ├── complexity/   # Static: complexity
│   │   ├── cohesion/     # Static: cohesion
│   │   ├── comments/     # Static: comments
│   │   ├── halstead/     # Static: halstead
│   │   ├── imports/      # Static & history: imports
│   │   ├── burndown/     # History: burndown
│   │   ├── couples/      # History: couples
│   │   ├── devs/         # History: devs
│   │   ├── file_history/ # History: file-history
│   │   ├── sentiment/    # History: sentiment
│   │   ├── shotness/     # History: shotness
│   │   └── typos/        # History: typos
│   └── uast/             # UAST parsing (60+ languages)
├── docs/                 # Documentation
└── test/                 # Test data
```

---

## Supported Languages

The UAST parser supports 60+ languages including:

Go, Python, JavaScript, TypeScript, Java, Kotlin, Rust, C, C++, C#, Ruby, PHP, Swift, Scala, Haskell, Clojure, Lua, Perl, R, SQL, GraphQL, and many more.

See [pkg/uast/LANGUAGE_ROADMAP.md](pkg/uast/LANGUAGE_ROADMAP.md) for the complete list.

---

## Documentation

- [Development Guide](docs/DEVELOPMENT.md)
- [Configuration](docs/CONFIGURATION.md)
- [API Reference](docs/API.md)
- [UAST Language Roadmap](pkg/uast/LANGUAGE_ROADMAP.md)

---

## Project History

This project was originally forked from [src-d/hercules](https://github.com/src-d/hercules), but has since been **completely re-architected**. It is not API or feature compatible with the original.

---

## License

[Apache 2.0](./LICENSE.md)
