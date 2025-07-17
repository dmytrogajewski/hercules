package uast

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	forest "github.com/alexaandru/go-sitter-forest"
	sitter "github.com/alexaandru/go-tree-sitter-bare"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/mapping"
)

// Loader loads UAST providers for different languages.
type Loader struct {
	embedFS fs.FS
}

// NewLoader creates a new loader with the given embedded filesystem.
func NewLoader(embedFS fs.FS) *Loader {
	return &Loader{
		embedFS: embedFS,
	}
}

// LoadProvider loads a provider for the given language.
func (l *Loader) LoadProvider(language string) (Provider, error) {
	dslContent, err := l.loadDSLMapping(language)

	if err != nil {
		return nil, fmt.Errorf("failed to load mapping for %s: %w", language, err)
	}

	rules, err := (&mapping.MappingParser{}).ParseMapping(string(dslContent))

	if err != nil {
		return nil, fmt.Errorf("failed to parse DSL mappings for %s: %w", language, err)
	}

	// Removed ValidateMappings, parsing already validates rules

	lang := l.getLanguage(language)

	return NewDSLProvider(lang, language, rules), nil
}

// loadDSLMapping loads DSL mapping content for a language.
func (l *Loader) loadDSLMapping(language string) ([]byte, error) {
	if l.embedFS == nil {
		return nil, fmt.Errorf("embedFS is nil")
	}

	mappingPath := filepath.Join("uastmaps", language+".uastmap")
	content, err := fs.ReadFile(l.embedFS, mappingPath)

	if err != nil {
		return nil, fmt.Errorf("mapping file not found: %s", mappingPath)
	}

	return content, nil
}

// getLanguage returns the Tree-sitter language for the given language name.
func (l *Loader) getLanguage(language string) *sitter.Language {
	return forest.GetLanguage(language)
}

// LoadAllProviders loads all available providers.
func (l *Loader) LoadAllProviders() (map[string]Provider, error) {
	providers := make(map[string]Provider)

	// Walk through the providers directory to find all .uastmap files
	err := fs.WalkDir(l.embedFS, "providers", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".uastmap") {
			return nil
		}

		// Extract language name from filename
		base := filepath.Base(path)
		language := strings.TrimSuffix(base, ".uastmap")

		provider, err := l.LoadProvider(language)
		if err != nil {
			return fmt.Errorf("failed to load provider for %s: %w", language, err)
		}

		providers[language] = provider
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to load providers: %w", err)
	}

	return providers, nil
}
