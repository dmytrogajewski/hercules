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
	loader *Loader
}

// NewParser creates a new Parser with DSL-based language parsers.
// It loads parser configurations and instantiates parsers for each supported language.
// Returns a pointer to the Parser or an error if loading parsers fails.
func NewParser() (*Parser, error) {
	loader := NewLoader(uastMapFs)

	p := &Parser{
		loader: loader,
	}

	return p, nil
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
