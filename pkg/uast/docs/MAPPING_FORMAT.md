# UAST Mapping Format

Mapping files are YAML or JSON, keyed by language, and are **embedded into the binary at build time**.

Each language maps Tree-sitter node kinds to:
- `type`: UAST node type (string)
- `roles`: optional list of roles (strings)
- `properties`: optional map of string to value

Example YAML:

```yaml
go:
  function_declaration:
    type: FunctionDecl
    roles: [Function, Declaration]
  identifier:
    type: Identifier
    roles: [Name]
```

**Important:**
- Mapping files are not loaded from disk at runtime.
- To update mappings, edit the YAML, rebuild the binary, and redeploy.
- No runtime reload or hot-reloading is possible. 