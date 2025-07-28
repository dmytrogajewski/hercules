package uast

import (
	"embed"
	"fmt"
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

//go:embed uastmaps/*.uastmap
var uastMapFs embed.FS

// Parser implements LanguageParser using embedded parsers
// Entry point for UAST parsing
// Parser is the main entry point for UAST parsing. It manages language parsers and their configurations.
type Parser struct {
	loader     *Loader
	customMaps map[string]UASTMap
}

// NewParser creates a new Parser with DSL-based language parsers.
// It loads parser configurations and instantiates parsers for each supported language.
// Returns a pointer to the Parser or an error if loading parsers fails.
func NewParser() (*Parser, error) {
	loader := NewLoader(uastMapFs)

	p := &Parser{
		loader:     loader,
		customMaps: make(map[string]UASTMap),
	}

	return p, nil
}

// WithUASTMap adds custom UAST mappings to the parser using the option pattern.
// This method allows passing custom UAST map configurations that will be used
// in addition to or as a replacement for the embedded mappings.
func (p *Parser) WithUASTMap(maps map[string]UASTMap) *Parser {
	// Store custom maps
	for name, uastMap := range maps {
		p.customMaps[name] = uastMap
	}

	// Load custom parsers from the provided mappings
	p.loadCustomParsers()

	return p
}

// loadCustomParsers loads parsers from custom UAST mappings
func (p *Parser) loadCustomParsers() {
	for name, uastMap := range p.customMaps {
		// Create a reader from the UAST string
		reader := strings.NewReader(uastMap.UAST)

		// Load parser from the custom mapping
		parser, err := p.loader.LoadParser(reader)
		if err != nil {
			fmt.Printf("failed to load custom parser for %s: %v\n", name, err)
			continue
		}

		// Register the parser with the loader
		p.loader.parsers[parser.Language()] = parser

		// Register extensions
		for _, ext := range parser.Extensions() {
			p.loader.extensions[strings.ToLower(ext)] = parser
		}
	}
}

// IsSupported returns true if the given filename is supported by any parser.
func (p *Parser) IsSupported(filename string) bool {
	// Get file extension
	ext := strings.ToLower(getFileExtension(filename))
	if ext == "" {
		return false
	}

	// Check if any parser supports this file extension
	_, exists := p.loader.LanguageParser(ext)
	return exists
}

// Parse parses a file and returns its UAST.
func (p *Parser) Parse(filename string, content []byte) (*node.Node, error) {
	// Get file extension
	ext := strings.ToLower(getFileExtension(filename))
	if ext == "" {
		return nil, fmt.Errorf("no file extension found for %s", filename)
	}

	// Get parser for this extension
	parser, exists := p.loader.LanguageParser(ext)
	if !exists {
		return nil, fmt.Errorf("no parser found for extension %s", ext)
	}

	// Parse using the parser
	return parser.Parse(filename, content)
}
