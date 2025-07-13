package uast

import (
	"fmt"
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/uast/internal/node"
)

// Remove old Node interface and Position struct; replaced by canonical Node and Positions in node.go

// ChangeType represents the type of change between two nodes
type ChangeType int

const (
	ChangeAdded ChangeType = iota
	ChangeRemoved
	ChangeModified
)

func (ct ChangeType) String() string {
	switch ct {
	case ChangeAdded:
		return "added"
	case ChangeRemoved:
		return "removed"
	case ChangeModified:
		return "modified"
	default:
		return "unknown"
	}
}

// Change represents a structural change between two versions of code
type Change struct {
	Before *node.Node
	After  *node.Node
	Type   ChangeType
	File   string
}

// Provider is responsible for parsing source code into UAST nodes
type Provider interface {
	Parse(filename string, content []byte) (*node.Node, error)
	SupportedLanguages() []string
	IsSupported(filename string) bool
}

// ProviderType represents the type of UAST provider
type ProviderType string

const (
	ProviderEmbedded  ProviderType = "embedded"
	ProviderBabelfish ProviderType = "babelfish"
)

// ConditionalFilter defines conditions for including/excluding child nodes
type ConditionalFilter struct {
	Type          string            `yaml:"type,omitempty"`
	ParentContext string            `yaml:"parent_context,omitempty"`
	HasField      string            `yaml:"has_field,omitempty"`
	Props         map[string]string `yaml:"props,omitempty"`
}

// ChildMapping defines how to map child nodes with optional filtering
type ChildMapping struct {
	Type        string             `yaml:"type"`
	ExcludeIf   *ConditionalFilter `yaml:"exclude_if,omitempty"`
	IncludeOnly *ConditionalFilter `yaml:"include_only,omitempty"`
}

// NameExtraction defines how to extract names from nodes
type NameExtraction struct {
	Source string `yaml:"source"` // "fields.name", "props.name", or "text"
}

// Error types for better error handling
type UnsupportedLanguageError struct {
	Language string
	Filename string
}

func (e UnsupportedLanguageError) Error() string {
	return fmt.Sprintf("unsupported language for file %s: %s", e.Filename, e.Language)
}

type ParseError struct {
	Filename string
	Language string
	Message  string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error in %s (%s): %s", e.Filename, e.Language, e.Message)
}

type ProviderError struct {
	Provider string
	Message  string
}

func (e ProviderError) Error() string {
	return fmt.Sprintf("provider error (%s): %s", e.Provider, e.Message)
}

// getFileExtension returns the file extension (with dot)
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return "." + parts[len(parts)-1]
}
