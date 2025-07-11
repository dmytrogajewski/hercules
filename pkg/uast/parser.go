package uast

import (
	"strings"
)

// Parser implements Provider using embedded parsers
// Entry point for UAST parsing
// Parser is the main entry point for UAST parsing. It manages language providers and their configurations.
type Parser struct {
	providers map[string]LanguageProvider
	configs   Providers // language -> *ProviderConfig
}

// LanguageProvider defines the interface for language-specific UAST providers.
type LanguageProvider interface {
	// Parse parses the given file content and returns the root UAST node.
	// filename is the name of the file being parsed.
	// content is the file content as bytes.
	// Returns the root Node or an error if parsing fails.
	Parse(filename string, content []byte) (*Node, error)
	// Language returns the language name handled by this provider.
	Language() string
}

// NewParser creates a new Parser with configuration-based language providers.
// It loads provider configurations and instantiates providers for each supported language.
// Returns a pointer to the Parser or an error if loading providers fails.
func NewParser() (*Parser, error) {
	configs, err := LoadProviders()
	if err != nil {
		return nil, err
	}
	p := &Parser{
		providers: make(map[string]LanguageProvider),
		configs:   configs,
	}
	for language, providerConfig := range configs {
		provider := NewProviderForLanguage(language, providerConfig)
		if provider != nil {
			p.providers[language] = provider
		}
	}
	return p, nil
}

// NewProviderForLanguage instantiates a provider for a given language and configuration.
// Returns a LanguageProvider or nil if the language is not supported.
func NewProviderForLanguage(language string, config *ProviderConfig) LanguageProvider {
	return FactoryCreateProvider(language, config)
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
func (p *Parser) Parse(filename string, content []byte) (*Node, error) {
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
	ext := strings.ToLower(getFileExtension(filename))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	for language, config := range p.configs {
		for _, configExt := range config.Extensions {
			if configExt == ext {
				return language
			}
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
