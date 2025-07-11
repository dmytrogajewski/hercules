# How to Add a New Language (UAST)

1. Install the Tree-sitter grammar for your language:
   - `go get github.com/tree-sitter/tree-sitter-<lang>/bindings/go`
2. Create or edit a YAML mapping file in the provider directory:
   - Map Tree-sitter node kinds to UAST `type` and (optionally) `roles`.
   - Example:
     ```yaml
     rust:
       function_item:
         type: Function
         roles: [Function, Declaration]
       identifier:
         type: Identifier
         roles: [Name]
     ```
3. **Rebuild the binary and redeploy** to apply the new or updated mapping. (Mapping files are embedded at build time; no runtime reload is possible for now.)
4. No code changes are required for new mappings—just update the mapping file, rebuild, and redeploy.
5. Test parsing and queries using the CLI:
   - `uast parse -lang rust main.rs | uast query 'filter(.type == "Function")'`

See GoDoc and the mapping format summary for more details. 