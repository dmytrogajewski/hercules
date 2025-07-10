# Feature Requirements Document: UAST CLI Tool

## Overview
Implement a command-line interface (CLI) tool for UAST operations, including parsing, querying, formatting, and diffing source code using the canonical UAST and DSL. The CLI should be ergonomic, scriptable, and consistent with other tools in the codebase (spf13/cobra/viper stack).

## Goals
- Provide a single binary (`uast`) for all UAST-related operations.
- Support the following subcommands:
  - `uast parse`: Parse source code to UAST (JSON/protobuf output)
  - `uast query`: Run DSL queries on UAST (input: file/stdin, output: JSON)
  - `uast fmt`: Pretty-print or normalize UAST JSON
  - `uast diff`: Compare two UASTs and report structural changes
- Enable scripting and integration in CI/CD pipelines.
- Support all languages and mappings available in the core library.

## Requirements

### 1. CLI Structure
- Use spf13/cobra for command structure and help.
- Use spf13/viper for configuration (flags, env, config file).
- Top-level command: `uast`
- Subcommands:
  - `parse`:
    - Args: `--lang`, `--out`, `--format` (json/proto), `--mapping`
    - Input: source file(s) or stdin
    - Output: UAST (JSON/protobuf)
  - `query`:
    - Args: `--query`, `--in`, `--out`, `--format`
    - Input: UAST file (JSON/protobuf)
    - Output: Query result (JSON)
  - `fmt`:
    - Args: `--in`, `--out`, `--pretty`
    - Input: UAST file (JSON)
    - Output: Formatted UAST JSON
  - `diff`:
    - Args: `--a`, `--b`, `--out`, `--format`
    - Input: Two UAST files (JSON/protobuf)
    - Output: Diff report (JSON)
- All commands must support stdin/stdout for piping.
- All commands must have `--help` and clear usage examples.

### 2. Behavior
- `parse` auto-detects language by extension or uses `--lang`.
- `query` runs a DSL query string or file on a UAST input.
- `fmt` pretty-prints or normalizes UAST JSON.
- `diff` compares two UASTs and outputs a structural diff.
- All commands must exit nonzero on error and print errors to stderr.
- Support for both JSON and protobuf (where applicable).

### 3. Extensibility
- Adding new subcommands or options should require minimal code changes.
- New languages and mappings are supported automatically via the core library.
- CLI config can be extended via viper (env/config file/flags).

### 4. Implementation
- Use spf13/cobra for CLI structure.
- Use spf13/viper for config.
- Use the canonical UAST parser, mapping, and DSL runtime from the core package.
- Write idiomatic, testable Go code.

### 5. Test Plan
- Unit tests for all command handlers (table-driven, error cases, help output).
- Integration tests for end-to-end workflows (parse → query → fmt → diff).
- Golden tests for CLI output (input → expected output JSON/proto).
- Fuzz tests for CLI input (optional).

## Out of Scope
- GUI or web interface (future work)
- Advanced diff/merge UI (future work)

## Acceptance Criteria
- All subcommands implemented and tested.
- CLI is scriptable, ergonomic, and consistent with other project tools.
- All tests pass and coverage is high.
- Documentation and usage examples are provided. 