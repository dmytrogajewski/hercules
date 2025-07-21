package extractor

import (
	"regexp"
	"strings"

	"github.com/dmytrogajewski/hercules/internal/pkg/importmodel"
)

// Extractor extracts imports from source code
type Extractor struct {
	// No tree-sitter languages needed for text-based extraction
}

// NewExtractor creates a new imports extractor
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract extracts imports from a file based on its name and content
func Extract(filename string, data []byte) (*importmodel.File, error) {
	// Determine language from file extension
	language := getLanguageFromFilename(filename)

	// Create extractor and extract imports
	extractor := NewExtractor()
	defer extractor.Close()

	imports, err := extractor.ExtractImports(language, data)

	return &importmodel.File{
		Imports: imports,
		Lang:    language,
		Error:   err,
	}, nil
}

// getLanguageFromFilename determines the programming language from the file extension
func getLanguageFromFilename(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}

	ext := strings.ToLower(parts[len(parts)-1])
	switch ext {
	case "go":
		return "golang"
	case "py":
		return "python"
	case "java":
		return "java"
	case "js":
		return "javascript"
	case "ts":
		return "typescript"
	case "cpp", "cc", "cxx", "c++":
		return "cpp"
	case "cs":
		return "csharp"
	case "rs":
		return "rust"
	case "rb":
		return "ruby"
	case "php":
		return "php"
	default:
		return ""
	}
}

// ExtractImports extracts import statements from the given source code
func (e *Extractor) ExtractImports(language string, sourceCode []byte) ([]string, error) {
	// Use sophisticated text-based extraction for all languages
	return e.extractImportsAdvanced(language, sourceCode)
}

// extractImportsAdvanced uses sophisticated regex patterns for accurate import extraction
func (e *Extractor) extractImportsAdvanced(language string, sourceCode []byte) ([]string, error) {
	switch language {
	case "golang":
		return e.extractGoImportsAdvanced(sourceCode)
	case "python":
		return e.extractPythonImportsAdvanced(sourceCode)
	case "java":
		return e.extractJavaImportsAdvanced(sourceCode)
	case "javascript":
		return e.extractJavaScriptImportsAdvanced(sourceCode)
	case "typescript":
		return e.extractTypeScriptImportsAdvanced(sourceCode)
	case "csharp":
		return e.extractCSharpImportsAdvanced(sourceCode)
	case "cpp":
		return e.extractCppImportsAdvanced(sourceCode)
	case "rust":
		return e.extractRustImportsAdvanced(sourceCode)
	case "ruby":
		return e.extractRubyImportsAdvanced(sourceCode)
	case "php":
		return e.extractPhpImportsAdvanced(sourceCode)
	default:
		return nil, nil
	}
}

// extractGoImportsAdvanced extracts Go import statements with sophisticated parsing
func (e *Extractor) extractGoImportsAdvanced(sourceCode []byte) ([]string, error) {
	imports := []string{}

	// Handle single import: import "package"
	singleImportRegex := regexp.MustCompile(`(?m)^\s*import\s+"([^"]+)"`)
	matches := singleImportRegex.FindAllSubmatch(sourceCode, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, string(match[1]))
		}
	}

	// Handle grouped imports: import ( "package1" "package2" )
	groupedImportRegex := regexp.MustCompile(`(?s)import\s*\(\s*((?:\s*"[^"]+"\s*)+)\s*\)`)
	groupMatches := groupedImportRegex.FindAllSubmatch(sourceCode, -1)
	for _, match := range groupMatches {
		if len(match) > 1 {
			// Extract individual imports from the group
			importGroup := string(match[1])
			individualImports := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(importGroup, -1)
			for _, imp := range individualImports {
				if len(imp) > 1 {
					imports = append(imports, imp[1])
				}
			}
		}
	}

	return imports, nil
}

// extractPythonImportsAdvanced extracts Python import statements with sophisticated parsing
func (e *Extractor) extractPythonImportsAdvanced(sourceCode []byte) ([]string, error) {
	imports := []string{}

	// Handle: import module
	simpleImportRegex := regexp.MustCompile(`(?m)^\s*import\s+([a-zA-Z_][a-zA-Z0-9_.]*)\s*$`)
	matches := simpleImportRegex.FindAllSubmatch(sourceCode, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, string(match[1]))
		}
	}

	// Handle: from module import item
	fromImportRegex := regexp.MustCompile(`(?m)^\s*from\s+([a-zA-Z_][a-zA-Z0-9_.]*)\s+import`)
	fromMatches := fromImportRegex.FindAllSubmatch(sourceCode, -1)
	for _, match := range fromMatches {
		if len(match) > 1 {
			imports = append(imports, string(match[1]))
		}
	}

	return imports, nil
}

// extractJavaImportsAdvanced extracts Java import statements with sophisticated parsing
func (e *Extractor) extractJavaImportsAdvanced(sourceCode []byte) ([]string, error) {
	imports := []string{}

	// Handle: import package.Class;
	javaImportRegex := regexp.MustCompile(`(?m)^\s*import\s+([a-zA-Z_][a-zA-Z0-9_.]*);`)
	matches := javaImportRegex.FindAllSubmatch(sourceCode, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, string(match[1]))
		}
	}

	return imports, nil
}

// extractJavaScriptImportsAdvanced extracts JavaScript import statements with sophisticated parsing
func (e *Extractor) extractJavaScriptImportsAdvanced(sourceCode []byte) ([]string, error) {
	imports := []string{}
	seen := make(map[string]bool)

	// Comprehensive regex that handles all JavaScript import syntax in one pass
	// Matches:
	// - import defaultExport from 'module'
	// - import { namedExport } from 'module'
	// - import * as namespace from 'module'
	// - import 'module'
	// - import defaultExport, { namedExport } from 'module'
	// - import defaultExport, * as namespace from 'module'
	jsImportRegex := regexp.MustCompile(`(?m)^\s*import\s+(?:(?:[\w$*{}, \n]+)\s+from\s+)?['"]([^'"]+)['"]`)
	matches := jsImportRegex.FindAllSubmatch(sourceCode, -1)

	for _, match := range matches {
		if len(match) > 1 {
			module := string(match[1])
			if !seen[module] {
				imports = append(imports, module)
				seen[module] = true
			}
		}
	}

	return imports, nil
}

// extractTypeScriptImportsAdvanced extracts TypeScript import statements
func (e *Extractor) extractTypeScriptImportsAdvanced(sourceCode []byte) ([]string, error) {
	// TypeScript imports are similar to JavaScript
	return e.extractJavaScriptImportsAdvanced(sourceCode)
}

// extractCSharpImportsAdvanced extracts C# using statements with sophisticated parsing
func (e *Extractor) extractCSharpImportsAdvanced(sourceCode []byte) ([]string, error) {
	imports := []string{}

	// Handle: using System;
	csharpImportRegex := regexp.MustCompile(`(?m)^\s*using\s+([a-zA-Z_][a-zA-Z0-9_.]*);`)
	matches := csharpImportRegex.FindAllSubmatch(sourceCode, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, string(match[1]))
		}
	}

	return imports, nil
}

// extractCppImportsAdvanced extracts C++ include statements with sophisticated parsing
func (e *Extractor) extractCppImportsAdvanced(sourceCode []byte) ([]string, error) {
	imports := []string{}

	// Handle: #include <header>
	// Handle: #include "header"
	cppIncludeRegex := regexp.MustCompile(`(?m)^\s*#include\s+[<"]([^>"]+)[>"]`)
	matches := cppIncludeRegex.FindAllSubmatch(sourceCode, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, string(match[1]))
		}
	}

	return imports, nil
}

// extractRustImportsAdvanced extracts Rust use statements with sophisticated parsing
func (e *Extractor) extractRustImportsAdvanced(sourceCode []byte) ([]string, error) {
	imports := []string{}

	// Handle: use crate::module;
	// Handle: use serde::{Deserialize, Serialize};
	rustUseRegex := regexp.MustCompile(`(?m)^\s*use\s+([a-zA-Z_][a-zA-Z0-9_:]*(?:\s*::\s*\{[^}]+\})?);`)
	matches := rustUseRegex.FindAllSubmatch(sourceCode, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, string(match[1]))
		}
	}

	return imports, nil
}

// extractRubyImportsAdvanced extracts Ruby require/load statements with sophisticated parsing
func (e *Extractor) extractRubyImportsAdvanced(sourceCode []byte) ([]string, error) {
	imports := []string{}

	// Handle: require 'gem'
	// Handle: load 'file'
	rubyImportRegex := regexp.MustCompile(`(?m)^\s*(?:require|load)\s+['"]([^'"]+)['"]`)
	matches := rubyImportRegex.FindAllSubmatch(sourceCode, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, string(match[1]))
		}
	}

	return imports, nil
}

// extractPhpImportsAdvanced extracts PHP use statements with sophisticated parsing
func (e *Extractor) extractPhpImportsAdvanced(sourceCode []byte) ([]string, error) {
	imports := []string{}

	// Handle: use Namespace\Class;
	phpUseRegex := regexp.MustCompile(`(?m)^\s*use\s+([a-zA-Z_][a-zA-Z0-9_\\]*);`)
	matches := phpUseRegex.FindAllSubmatch(sourceCode, -1)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, string(match[1]))
		}
	}

	return imports, nil
}

// Close closes all parsers and frees resources
func (e *Extractor) Close() {
	// No resources to close in text-based extraction
}
