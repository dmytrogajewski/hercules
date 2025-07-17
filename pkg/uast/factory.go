package uast

import (
	"fmt"
	"strings"

	forest "github.com/alexaandru/go-sitter-forest"
	sitter "github.com/alexaandru/go-tree-sitter-bare"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/mapping"
)

// ProviderFactory creates UAST providers for different languages and mapping types.
type ProviderFactory struct {
	// Embed FS for loading mapping files
	embedFS interface{} // Replace with actual embed.FS when available
}

// NewProviderFactory creates a new provider factory.
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{}
}

// CreateProvider creates a provider for the given language.
func (pf *ProviderFactory) CreateProvider(langName string) (Provider, error) {
	// Check if we have a DSL mapping file for this language
	dslContent, err := pf.loadDSLMapping(langName)
	if err != nil {
		return nil, fmt.Errorf("no mapping configuration found for %s", langName)
	}

	return pf.CreateDSLProvider(langName, dslContent)
}

// CreateDSLProvider creates a provider using DSL-based mappings.
func (pf *ProviderFactory) CreateDSLProvider(langName string, dslContent string) (Provider, error) {
	// Get the language from the language registry
	lang, err := pf.getLanguage(langName)
	if err != nil {
		return nil, fmt.Errorf("unsupported language %s: %w", langName, err)
	}

	// Parse DSL content into mapping rules
	rules, err := (&mapping.MappingParser{}).ParseMapping(dslContent)
	if err != nil {
		return nil, fmt.Errorf("failed to load DSL mappings: %w", err)
	}

	// Removed ValidateMappings, parsing already validates rules

	// Type assertion for the language
	if sitterLang, ok := lang.(*sitter.Language); ok {
		return NewDSLProvider(sitterLang, langName, rules), nil
	}

	return nil, fmt.Errorf("invalid language type for %s", langName)
}

// getLanguage returns the Tree-sitter language for the given language name.
func (pf *ProviderFactory) getLanguage(langName string) (interface{}, error) {
	lang := forest.GetLanguage(langName)
	if lang == nil {
		return nil, fmt.Errorf("language %s not found", langName)
	}
	return lang, nil
}

// detectLanguageFromFile detects the language from a filename.
func (pf *ProviderFactory) detectLanguageFromFile(filename string) string {
	ext := pf.getFileExtension(filename)
	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".c":
		return "c"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".cs":
		return "csharp"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".hs":
		return "haskell"
	case ".ml":
		return "ocaml"
	case ".fs":
		return "fsharp"
	case ".clj":
		return "clojure"
	case ".erl":
		return "erlang"
	case ".ex", ".exs":
		return "elixir"
	case ".lua":
		return "lua"
	case ".pl":
		return "perl"
	case ".r":
		return "r"
	case ".dart":
		return "dart"
	case ".sh", ".bash":
		return "bash"
	case ".ps1":
		return "powershell"
	case ".sql":
		return "sql"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".xml":
		return "xml"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".ini":
		return "ini"
	case ".md":
		return "markdown"
	case ".dockerfile":
		return "dockerfile"
	case ".makefile":
		return "makefile"
	default:
		return ""
	}
}

// getFileExtension returns the file extension (with dot).
func (pf *ProviderFactory) getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return "." + parts[len(parts)-1]
}

// loadDSLMapping loads DSL mapping content for a language.
func (pf *ProviderFactory) loadDSLMapping(langName string) (string, error) {
	// This would load from embedded files
	// For now, return empty - this would load from embedded DSL files
	return "", fmt.Errorf("DSL mapping not found for %s", langName)
}

// FactoryCreateProvider creates a provider for the given language (legacy function).
func FactoryCreateProvider(language string, config interface{}) Provider {
	// This is a legacy function that should be removed
	// For now, return nil to indicate it's not supported
	return nil
}
