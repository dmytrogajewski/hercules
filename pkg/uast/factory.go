package uast

import (
	forest "github.com/alexaandru/go-sitter-forest"
)

type ParserType string

const (
	ParserNative     ParserType = "native"
	ParserTreeSitter ParserType = "tree-sitter"
	ParserExternal   ParserType = "external"
)

// FactoryCreateProvider instantiates a provider for a given language/config
func FactoryCreateProvider(language string, config *ProviderConfig) LanguageProvider {
	switch ParserType(config.Parser) {
	case ParserNative:
		// Add native providers here as needed
		return nil
	case ParserTreeSitter:
		if !forest.SupportedLanguage(language) {
			return nil
		}
		tsLang := forest.GetLanguage(language)
		if tsLang == nil {
			return nil
		}
		return &TreeSitterProvider{
			language: tsLang,
			langName: language,
			mapping:  config.Mapping,
		}
	case ParserExternal:
		// TODO: Implement external provider creation
		return nil
	default:
		return nil
	}
}

// FactoryCreateProviderWithOptions instantiates a provider for a given language/config with options
func FactoryCreateProviderWithOptions(language string, config *ProviderConfig, opts ParserOptions) LanguageProvider {
	switch ParserType(config.Parser) {
	case ParserNative:
		return nil
	case ParserTreeSitter:
		if !forest.SupportedLanguage(language) {
			return nil
		}
		tsLang := forest.GetLanguage(language)
		if tsLang == nil {
			return nil
		}
		return &TreeSitterProvider{
			language:        tsLang,
			langName:        language,
			mapping:         config.Mapping,
			IncludeUnmapped: opts.IncludeUnmapped,
		}
	case ParserExternal:
		return nil
	default:
		return nil
	}
}
