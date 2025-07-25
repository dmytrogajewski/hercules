// Code generated by scripts/precompile.go. DO NOT EDIT.
//go:generate sh -c "cd ../../ && go run ./scripts/precompile.go -o pkg/uast/embedded_mappings.gen.go"

package uast

import (
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/mapping"
)

// PrecompiledMapping represents the pre-compiled mapping data
type PrecompiledMapping struct {
	Language   string                 `json:"language"`
	Extensions []string               `json:"extensions"`
	Rules      []mapping.MappingRule  `json:"rules"`
	Patterns   map[string]interface{} `json:"patterns"`
	CompiledAt string                 `json:"compiled_at"`
}

// Loader loads UAST parsers for different languages.
type Loader struct {
	embedFS    fs.FS
	parsers    map[string]LanguageParser
	extensions map[string]LanguageParser // extension -> parser mapping
}

// NewLoader creates a new loader with the given embedded filesystem.
func NewLoader(embedFS fs.FS) *Loader {
	l := &Loader{
		embedFS:    embedFS,
		parsers:    make(map[string]LanguageParser),
		extensions: make(map[string]LanguageParser),
	}

	l.loadUASTParsers()
	return l
}

// loadUASTParsers loads parsers, trying pre-compiled cache first
func (l *Loader) loadUASTParsers() {
	if l.loadFromCache() {
		return
	}

	l.loadFromFiles()
}

// loadFromCache attempts to load parsers from the pre-compiled cache
func (l *Loader) loadFromCache() bool {
	return l.loadFromEmbeddedMappings()
}

// loadFromEmbeddedMappings loads parsers from embedded mappings (if available)
func (l *Loader) loadFromEmbeddedMappings() bool {
	if embeddedMappingsAvailable() {
		return l.loadFromEmbeddedMappingsData()
	}
	return false
}

// loadFromFiles loads parsers from individual uastmap files
func (l *Loader) loadFromFiles() {
	if l.embedFS == nil {
		return
	}

	err := fs.WalkDir(l.embedFS, "uastmaps", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".uastmap") {
			return nil
		}

		file, err := l.embedFS.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		p, err := l.LoadParser(file)
		if err != nil {
			fmt.Printf("failed to load parser for %s: %v\n", d.Name(), err)
			return nil
		}

		l.parsers[p.Language()] = p

		for _, ext := range p.Extensions() {
			l.extensions[strings.ToLower(ext)] = p
		}

		return nil
	})

	if err != nil {
		fmt.Printf("error discovering parsers: %v\n", err)
	}
}

// LoadParser loads a parser by reading the uastmap file through the reader
func (l *Loader) LoadParser(reader io.Reader) (LanguageParser, error) {
	dslp := NewDSLParser(reader)

	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic while loading parser: %v", r)
			}
		}()
		err = dslp.Load()
	}()

	if err != nil {
		return nil, err
	}

	return dslp, nil
}

// LanguageParser returns the parser for the given file extension.
func (l *Loader) LanguageParser(extension string) (LanguageParser, bool) {
	parser, exists := l.extensions[strings.ToLower(extension)]
	return parser, exists
}

// GetParsers returns all loaded parsers.
func (l *Loader) GetParsers() map[string]LanguageParser {
	return l.parsers
}
