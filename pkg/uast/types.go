package uast

import (
	"fmt"
	"strings"

	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
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

// Parser is responsible for parsing source code into UAST nodes
type LanguageParser interface {
	Parse(filename string, content []byte) (*node.Node, error)
	Language() string
	Extensions() []string
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

type LanguageParserError struct {
	Parser  string
	Message string
}

func (e LanguageParserError) Error() string {
	return fmt.Sprintf("parser error (%s): %s", e.Parser, e.Message)
}

// getFileExtension returns the file extension (with dot)
func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return ""
	}
	return "." + parts[len(parts)-1]
}

// UASTMap represents a custom UAST mapping configuration
type UASTMap struct {
	Extensions []string `json:"extensions"`
	UAST       string   `json:"uast"`
}
