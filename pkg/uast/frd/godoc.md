# Feature Requirements Document: GoDoc Documentation for Public APIs

## Overview
Document all public APIs in the UAST package using idiomatic GoDoc comments. Documentation must be clear, concise, and provide usage examples for each major function, type, and method. The documentation must be kept up to date with the codebase and reviewed for completeness and accuracy.

## Goals
- Provide comprehensive, idiomatic GoDoc for all public APIs in the UAST package.
- Ensure documentation is clear, concise, and includes usage examples.
- Make the package easy to use and understand for new and existing users.
- Achieve 100% documentation coverage for all exported types, functions, and methods.

## API Requirements
- All exported types, functions, and methods must have GoDoc comments.
- Comments must describe the purpose, parameters, return values, and any important edge cases or behaviors.
- Usage examples must be provided for all major APIs (Node, PreOrder, HasRole, Transform, FindDSL, etc).
- Comments must follow GoDoc style and be suitable for godoc.org/pkg.go.dev rendering.
- Documentation must be kept up to date with code changes.

## Extensibility & Edge Cases
- Documentation must be extensible to cover new APIs and features as they are added.
- Edge cases, panics, and error behaviors must be documented.
- Any concurrency or mutation caveats must be clearly described.

## Out of Scope
- Non-GoDoc documentation formats (e.g., Markdown, Sphinx) for API docs (these are for guides/examples, not API docs).
- Internal/private APIs (not exported from the package).

## Testability
- Documentation coverage must be checked using `golint`, `godoc`, or equivalent tools.
- Usage examples must be runnable as GoDoc examples (where possible).
- Documentation must be reviewed for clarity and completeness.

## Example GoDoc
```go
// PreOrder returns a channel streaming nodes in pre-order traversal (root, then children left-to-right).
// The channel is closed when traversal is complete. If root is nil, the channel is closed immediately.
//
// Example:
//   for n := range PreOrder(root) {
//       fmt.Println(n.Type)
//   }
func PreOrder(root *Node) <-chan *Node
```

--- 