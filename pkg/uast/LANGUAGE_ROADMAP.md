# UAST Language Support Roadmap

> **Note:** The UAST query DSL now supports nested property selectors (e.g., `.props.name`). For the current UAST structure, only one level of nesting is meaningful, but the implementation is future-proofed for deeper nesting if/when the UAST structure evolves.

Simple checklist for fixing UAST language mappings based on test results.

## Current Status

### ✅ Working
- [x] Go - All tests pass
- [x] Rust - All tests pass (basic functions + advanced features: structs, traits, enums, macros, modules)
- [x] Kotlin - All tests pass (basic functions + advanced features: classes, objects, data classes, enums, interfaces, extensions)
- [x] Python - All tests pass (TreeSitterPythonProvider with comprehensive mapping)
- [x] Swift - All tests pass

### 🔄 Partially Working (Need Mapping Fixes)
- [x] Perl - Function detection works, structure/mapping incomplete

### ❌ Not Working (Providers Exist But Mappings Broken or No Provider)
- [ ] JavaScript - Provider exists, parsing fails
- [ ] Java - Provider exists, parsing fails
- [ ] C++ - Provider exists, parsing fails
- [ ] Ruby - Provider exists, parsing fails
- [ ] PHP - Provider exists, parsing fails
- [ ] C# - Provider exists, parsing fails
- [ ] Scala - Provider exists, parsing fails
- [ ] Dart - Provider exists, parsing fails
- [ ] Lua - Provider exists, parsing fails
- [ ] CSS - Provider exists, parsing fails
- [ ] SQL - Provider exists, parsing fails
- [ ] Markdown - Provider exists, parsing fails
- [ ] TOML - Provider exists, parsing fails
- [ ] XML - Provider exists, parsing fails
- [ ] YAML - Provider exists, parsing fails
- [ ] Haskell - Provider exists, parsing fails
- [ ] OCaml - Provider exists, parsing fails
- [ ] Dockerfile - Provider exists, parsing fails
- [ ] Makefile - Provider exists, parsing fails
- [ ] Bash - Provider exists, parsing fails
- [ ] Clojure - Provider exists, parsing fails
- [ ] Elixir - No provider found
- [ ] F# - No provider found
- [ ] Erlang - No provider found
- [ ] HTML - No provider found
- [ ] JSON - No provider found
- [ ] INI - No provider found
- [ ] PowerShell - No provider found

## Priority Order

### Phase 1: Fix Partially Working (4 languages)
1. [x] Swift
2. [ ] Haskell
3. [ ] Perl
4. [ ] Dockerfile

### Phase 2: Fix Broken Mappings (15 languages)
1. [ ] JavaScript
2. [ ] Java
3. [ ] C++
4. [ ] Ruby
5. [ ] PHP
6. [ ] C#
7. [ ] Scala
8. [ ] Dart
9. [ ] Lua
10. [ ] CSS
11. [ ] SQL
12. [ ] Markdown
13. [ ] TOML
14. [ ] XML
15. [ ] YAML

### Phase 3: Add Missing Providers (11 languages)
1. [ ] Elixir
2. [ ] Clojure
3. [ ] OCaml
4. [ ] F#
5. [ ] Erlang
6. [ ] HTML
7. [ ] JSON
8. [ ] INI
9. [ ] Bash
10. [ ] PowerShell
11. [ ] Makefile

## Progress Summary
- Total languages: 33
- Working: 5 (Go, Rust, Kotlin, Python, Swift)
- Partially working: 1 (Perl)
- Not working: 27

> This reflects the latest automated test run.

## Next Steps
1. Start with Phase 1 - fix partially working languages
2. Use raw AST dumps to understand current parsing
3. Update mappings to match expected UAST structure
4. Run tests to validate fixes
5. Move to Phase 2 when Phase 1 is complete 