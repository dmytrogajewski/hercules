package uast

import (
	"strings"
)

// Parser implements Provider using embedded parsers
// Entry point for UAST parsing
type Parser struct {
	providers map[string]LanguageProvider
	configs   Providers // language -> *ProviderConfig
}

type LanguageProvider interface {
	Parse(filename string, content []byte) (*Node, error)
	Language() string
}

// NewParser creates a new Parser with configuration-based language providers
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

// NewProviderForLanguage instantiates a provider for a given language/config (delegates to factory)
func NewProviderForLanguage(language string, config *ProviderConfig) LanguageProvider {
	return FactoryCreateProvider(language, config)
}

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

func (p *Parser) SupportedLanguages() []string {
	languages := make([]string, 0, len(p.providers))
	for lang := range p.providers {
		languages = append(languages, lang)
	}
	return languages
}

func (p *Parser) IsSupported(filename string) bool {
	language := p.detectLanguage(filename)
	_, exists := p.providers[language]
	return exists
}
