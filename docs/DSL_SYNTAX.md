# UAST Query DSL Syntax

## Nested Property Selectors

The DSL now supports nested property selectors, allowing queries like `.props.name` or `.type`. This enables direct access to properties within the UAST node structure.

### Syntax
- `.field1.field2...fieldN` selects nested properties.
- For the current Hercules UAST, only `.props.name` (one level) is meaningful, since `Props` is a flat map.

### Examples
- `.type` — selects the node type
- `.props.name` — selects the value of the `name` property in the node's `Props` map
- `.token` — selects the node's token value

### Limitations
- For the current UAST structure, only one level of nesting is supported in practice (e.g., `.props.name`).
- Deeper nesting (e.g., `.props.meta.name`) will return no results, as `Props` is a flat `map[string]string`.
- The DSL and parser are future-proofed for deeper nesting if the UAST structure evolves. 