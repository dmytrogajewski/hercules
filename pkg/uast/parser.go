package uast

import (
	"embed"
	"slices"
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

var ProvidersMap map[string][]string = map[string][]string{
	"go": {".go"},
	//" csharp":     {".cs"},
	// "kotlin":     {".kt"},
	// "haskell":    {".hs"},
	// "ocaml":      {".ml"},
	// "fsharp":     {".fs"},
	// "clojure":    {".clj"},
	// "erlang":     {".erl"},
	// "elixir":     {".ex", ".exs"},
	// "bash":       {".sh", ".bash"},
	// "powershell": {".ps1"},
	// "javascript": {".js", ".jsx"},
	// "typescript": {".ts", ".tsx"},
	// "python":     {".py"},
	// "rust":       {".rs"},
	// "ruby":       {".rb"},
	// "php":        {".php"},
	// "sql":        {".sql"},
	// "html":       {".html", ".htm"},
	// "css":        {".css"},
	// "xml":        {".xml"},
	// "json":       {".json"},
	// "yaml":       {".yaml", ".yml"},
	// "toml":       {".toml"},
	// "ini":        {".ini"},
	// "markdown":   {".md"},
	// "dockerfile": {".dockerfile"},
	// "makefile":   {".makefile"},
	// "c":          {".c"},
	// "cpp":        {".cpp", ".cc", ".cxx"},
	// "java":       {".java"},
	// "scala":      {".scala"},
	// "lua":        {".lua"},
	// "perl":       {".pl"},
	// "r":          {".r"},
	// "dart":       {".dart"},
	// "sh":         {".sh", ".bash"},
	// "ps1":        {".ps1"},
}

//go:embed uastmaps
var uastMapFs embed.FS

// Parser implements Provider using embedded parsers
// Entry point for UAST parsing
// Parser is the main entry point for UAST parsing. It manages language providers and their configurations.
type Parser struct {
	providers map[string]Provider
}

// NewParser creates a new Parser with DSL-based language providers.
// It loads provider configurations and instantiates providers for each supported language.
// Returns a pointer to the Parser or an error if loading providers fails.
func NewParser() (*Parser, error) {
	p := &Parser{
		providers: make(map[string]Provider),
	}

	loader := NewLoader(uastMapFs)

	for lang := range ProvidersMap {
		pr, err := loader.LoadProvider(lang)
		if err != nil {
			return nil, err
		}
		p.providers[lang] = pr
	}

	return p, nil
}

// Parse parses the given file content using the appropriate language provider.
// filename is the name of the file being parsed.
// content is the file content as bytes.
// Returns the root Node or an error if parsing fails or the language is unsupported.
//
// Example:
//
//	parser, err := uast.NewParser()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	node, err := parser.Parse("main.go", []byte("package main\nfunc main() {}"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(node.Type)
func (p *Parser) Parse(filename string, content []byte) (*node.Node, error) {
	if filename == "" {
		return nil, ParseError{
			Filename: filename,
			Language: "unknown",
			Message:  "empty filename",
		}
	}
	language := p.detectLanguage(filename)
	if language == "" {
		return nil, UnsupportedLanguageError{
			Language: "unknown",
			Filename: filename,
		}
	}
	provider, exists := p.providers[language]
	if !exists {
		return nil, UnsupportedLanguageError{
			Language: language,
			Filename: filename,
		}
	}
	node, err := provider.Parse(filename, content)
	return node, err
}

func (p *Parser) detectLanguage(filename string) string {
	fileExt := strings.ToLower(getFileExtension(filename))
	if fileExt == "" {
		return ""
	}

	for lang, exts := range ProvidersMap {
		if slices.Contains(exts, fileExt) {
			return lang
		}
	}
	return ""
}

// SupportedLanguages returns a slice of all supported language names.
func (p *Parser) SupportedLanguages() []string {
	languages := make([]string, 0, len(p.providers))
	for lang := range p.providers {
		languages = append(languages, lang)
	}
	return languages
}

// IsSupported returns true if the given filename is supported by any provider.
func (p *Parser) IsSupported(filename string) bool {
	language := p.detectLanguage(filename)
	_, exists := p.providers[language]
	return exists
}
