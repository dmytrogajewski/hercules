package imports

import (
	"runtime"
	"strings"
	"sync"

	"github.com/dmytrogajewski/hercules/internal/app/core"
	"github.com/dmytrogajewski/hercules/internal/pkg/importmodel"
	"github.com/dmytrogajewski/hercules/internal/pkg/plumbing"
	"github.com/dmytrogajewski/hercules/pkg/uast"
	"github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
	"github.com/go-git/go-git/v6"
	gitplumbing "github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/utils/merkletrie"
)

var _ core.PipelineItem = (*Extractor)(nil)

// Extractor reports the imports in the changed files.
type Extractor struct {
	core.NoopMerger
	// Goroutines is the number of goroutines to run for imports extraction.
	Goroutines int
	// MaxFileSize is the file size threshold. Files that exceed it are ignored.
	MaxFileSize int

	l core.Logger
}

const (
	// DependencyImports is the name of the dependency provided by Extractor.
	DependencyImports = "imports"
	// ConfigImportsGoroutines is the name of the configuration option for
	// Extractor.Configure() to set the number of parallel goroutines for imports extraction.
	ConfigImportsGoroutines = "Imports.Goroutines"
	// ConfigMaxFileSize is the name of the configuration option for
	// Extractor.Configure() to set the file size threshold after which they are ignored.
	ConfigMaxFileSize = "Imports.MaxFileSize"
	// DefaultMaxFileSize is the default value for Extractor.MaxFileSize.
	DefaultMaxFileSize = 1 << 20
)

// Name of this PipelineItem. Uniquely identifies the type, used for mapping keys, etc.
func (ex *Extractor) Name() string {
	return "Imports"
}

// Provides returns the list of names of entities which are produced by this PipelineItem.
// Each produced entity will be inserted into `deps` of dependent Consume()-s according
// to this list. Also used by core.Registry to build the global map of providers.
func (ex *Extractor) Provides() []string {
	return []string{DependencyImports}
}

// Requires returns the list of names of entities which are needed by this PipelineItem.
// Each requested entity will be inserted into `deps` of Consume(). In turn, those
// entities are Provides() upstream.
func (ex *Extractor) Requires() []string {
	return []string{plumbing.DependencyTreeChanges, plumbing.DependencyBlobCache}
}

// ListConfigurationOptions returns the list of changeable public properties of this PipelineItem.
func (ex *Extractor) ListConfigurationOptions() []core.ConfigurationOption {
	return []core.ConfigurationOption{{
		Name:        ConfigImportsGoroutines,
		Description: "Specifies the number of goroutines to run in parallel for the imports extraction.",
		Flag:        "import-goroutines",
		Type:        core.IntConfigurationOption,
		Default:     runtime.NumCPU()}, {
		Name:        ConfigMaxFileSize,
		Description: "Specifies the file size threshold. Files that exceed it are ignored.",
		Flag:        "import-max-file-size",
		Type:        core.IntConfigurationOption,
		Default:     DefaultMaxFileSize},
	}
}

// Configure sets the properties previously published by ListConfigurationOptions().
func (ex *Extractor) Configure(facts map[string]interface{}) error {
	if l, exists := facts[core.ConfigLogger].(core.Logger); exists {
		ex.l = l
	}
	if gr, exists := facts[ConfigImportsGoroutines].(int); exists {
		if gr < 1 {
			if ex.l != nil {
				ex.l.Warnf("invalid number of goroutines for the imports extraction: %d. Set to %d.",
					gr, runtime.NumCPU())
			}
			gr = runtime.NumCPU()
		}
		ex.Goroutines = gr
	}
	if size, exists := facts[ConfigMaxFileSize].(int); exists {
		if size <= 0 {
			if ex.l != nil {
				ex.l.Warnf("invalid maximum file size: %d. Set to %d.", size, DefaultMaxFileSize)
			}
			size = DefaultMaxFileSize
		}
		ex.MaxFileSize = size
	}
	return nil
}

// Initialize resets the temporary caches and prepares this PipelineItem for a series of Consume()
// calls. The repository which is going to be analysed is supplied as an argument.
func (ex *Extractor) Initialize(repository *git.Repository) error {
	ex.l = core.GetLogger()
	if ex.Goroutines < 1 {
		ex.Goroutines = runtime.NumCPU()
	}
	if ex.MaxFileSize == 0 {
		ex.MaxFileSize = DefaultMaxFileSize
	}
	return nil
}

// Consume runs this PipelineItem on the next commit data.
// `deps` contain all the results from upstream PipelineItem-s as requested by Requires().
// Additionally, DependencyCommit is always present there and represents the analysed *object.Commit.
// This function returns the mapping with analysis results. The keys must be the same as
// in Provides(). If there was an error, nil is returned.
func (ex *Extractor) Consume(deps map[string]interface{}) (map[string]interface{}, error) {
	changes := deps[plumbing.DependencyTreeChanges].(object.Changes)
	cache := deps[plumbing.DependencyBlobCache].(map[gitplumbing.Hash]*plumbing.CachedBlob)
	result := map[gitplumbing.Hash]importmodel.File{}
	jobs := make(chan *object.Change, ex.Goroutines)
	resultSync := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(ex.Goroutines)
	for i := 0; i < ex.Goroutines; i++ {
		go func() {
			for change := range jobs {
				blob := cache[change.To.TreeEntry.Hash]
				if blob.Size > int64(ex.MaxFileSize) {
					ex.l.Warnf("skipped %s %s: size is too big: %d > %d",
						change.To.TreeEntry.Name, change.To.TreeEntry.Hash.String(),
						blob.Size, ex.MaxFileSize)
					continue
				}
				file, err := extractImports(change.To.TreeEntry.Name, blob.Data)
				if err != nil {
					ex.l.Errorf("failed to extract imports from %s %s: %v",
						change.To.TreeEntry.Name, change.To.TreeEntry.Hash.String(), err)
				} else {
					resultSync.Lock()
					result[change.To.TreeEntry.Hash] = *file
					resultSync.Unlock()
				}
			}
			wg.Done()
		}()
	}
	for _, change := range changes {
		action, err := change.Action()
		if err != nil {
			return nil, err
		}
		switch action {
		case merkletrie.Modify, merkletrie.Insert:
			jobs <- change
		case merkletrie.Delete:
			continue
		}
	}
	close(jobs)
	wg.Wait()
	return map[string]interface{}{DependencyImports: result}, nil
}

// Fork clones this PipelineItem.
func (ex *Extractor) Fork(n int) []core.PipelineItem {
	return core.ForkSamePipelineItem(ex, n)
}

// extractImports extracts imports from a file using UAST parsing
func extractImports(filename string, content []byte) (*importmodel.File, error) {
	// Create UAST parser
	parser, err := uast.NewParser()
	if err != nil {
		return &importmodel.File{
			Imports: []string{},
			Lang:    "",
			Error:   err,
		}, nil
	}

	// Check if file is supported
	if !parser.IsSupported(filename) {
		return &importmodel.File{
			Imports: []string{},
			Lang:    "",
			Error:   nil,
		}, nil
	}

	// Parse file to UAST
	uastNode, err := parser.Parse(filename, content)
	if err != nil {
		return &importmodel.File{
			Imports: []string{},
			Lang:    "",
			Error:   err,
		}, nil
	}

	if uastNode == nil {
		return &importmodel.File{
			Imports: []string{},
			Lang:    "",
			Error:   nil,
		}, nil
	}

	// Extract imports from UAST
	imports := extractImportsFromUAST(uastNode)

	return &importmodel.File{
		Imports: imports,
		Lang:    "",
		Error:   nil,
	}, nil
}

// extractImportsFromUAST extracts import strings from a UAST node tree
func extractImportsFromUAST(root *node.Node) []string {
	var imports []string
	seen := make(map[string]bool)

	// Traverse the UAST tree to find import nodes
	root.VisitPreOrder(func(n *node.Node) {
		// Look for nodes with Import type or Import role
		if n.Type == node.UASTImport || n.HasAnyRole(node.RoleImport) {
			// Extract import path from token or children
			if importPath := extractImportPath(n); importPath != "" {
				// Deduplicate imports
				if !seen[importPath] {
					imports = append(imports, importPath)
					seen[importPath] = true
				}
			}
		}
	})

	return imports
}

// extractImportPath extracts the import path from an import node
func extractImportPath(importNode *node.Node) string {
	// First try to get the import path from the token
	if importNode.Token != "" {
		return cleanImportPath(importNode.Token)
	}

	// For JavaScript imports, look for specific patterns in children
	if len(importNode.Children) > 0 {
		// Look for string literals that contain import paths
		for _, child := range importNode.Children {
			if child.Type == node.UASTLiteral && child.Token != "" {
				// This is likely a string literal containing the import path
				return cleanImportPath(child.Token)
			}
		}

		// Look for identifier nodes that might contain module names
		for _, child := range importNode.Children {
			if child.Type == node.UASTIdentifier && child.Token != "" {
				// This might be a module name
				return cleanImportPath(child.Token)
			}
		}

		// Recursively check children for import paths
		for _, child := range importNode.Children {
			if path := extractImportPath(child); path != "" {
				return path
			}
		}
	}

	return ""
}

// cleanImportPath cleans up an import path by removing quotes and extracting module names
func cleanImportPath(path string) string {
	// Remove surrounding quotes and trailing semicolons
	path = strings.Trim(path, `"';`)

	// Skip empty or invalid paths
	if path == "" || path == "{" || path == "}" {
		return ""
	}

	// Handle different import statement formats
	if strings.HasPrefix(path, "import ") {
		// Python: "import os" -> "os"
		parts := strings.Fields(path)
		if len(parts) >= 2 {
			return parts[1]
		}
	} else if strings.HasPrefix(path, "from ") {
		// Python: "from typing import List, Dict" -> "typing"
		parts := strings.Fields(path)
		if len(parts) >= 2 {
			return parts[1]
		}
	} else if strings.Contains(path, " from ") {
		// JavaScript: "React from 'react'" -> "react"
		parts := strings.Split(path, " from ")
		if len(parts) >= 2 {
			return strings.Trim(parts[1], `"'`)
		}
	} else if strings.Contains(path, "import ") {
		// JavaScript: "import './styles.css'" -> "./styles.css"
		parts := strings.Split(path, "import ")
		if len(parts) >= 2 {
			return strings.Trim(parts[1], `"'`)
		}
	} else if strings.HasPrefix(path, "{") && strings.Contains(path, "}") {
		// JavaScript destructuring: "{ useState, useEffect }" -> skip this
		return ""
	}

	// For simple module names, just return as is
	return path
}

func init() {
	core.Registry.Register(&Extractor{})
}
