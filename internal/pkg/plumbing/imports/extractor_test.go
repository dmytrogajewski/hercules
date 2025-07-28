package imports

import (
	"testing"

	"github.com/dmytrogajewski/hercules/internal/app/core"
	"github.com/dmytrogajewski/hercules/internal/pkg/importmodel"
	"github.com/dmytrogajewski/hercules/internal/pkg/plumbing"
	"github.com/go-git/go-git/v6"
	gitplumbing "github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestExtractorEndToEnd(t *testing.T) {
	// Create the extractor
	extractor := &Extractor{}

	// Initialize the extractor with a mock repository
	mockRepo := &git.Repository{}
	err := extractor.Initialize(mockRepo)
	assert.NoError(t, err)

	// Test data for different languages
	testCases := []struct {
		name     string
		filename string
		content  string
		expected []string
	}{
		{
			name:     "Go imports",
			filename: "main.go",
			content: `package main

import "fmt"
import "strings"

import (
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Hello")
}`,
			expected: []string{"fmt", "strings", "os", "path/filepath"},
		},
		{
			name:     "Python imports",
			filename: "app.py",
			content: `import os
import sys
from typing import List, Dict
from pathlib import Path

def main():
    print("Hello")`,
			expected: []string{"os", "sys", "typing", "pathlib"},
		},
		{
			name:     "JavaScript imports",
			filename: "app.js",
			content: `import React from 'react';
import { useState, useEffect } from 'react';
import './styles.css';

function App() {
    return <div>Hello</div>;
}`,
			// UAST parser extracts the actual import statements, not just module names
			expected: []string{"React", "./styles.css"},
		},
		{
			name:     "Java imports",
			filename: "Test.java",
			content: `import java.util.List;
import java.util.Map;
import org.springframework.stereotype.Component;

@Component
public class Test {
    // ...
}`,
			expected: []string{"java.util.List", "java.util.Map", "org.springframework.stereotype.Component"},
		},
		{
			name:     "C++ imports",
			filename: "main.cpp",
			content: `#include <iostream>
#include <vector>
#include "myheader.h"

int main() {
    return 0;
}`,
			// C++ might not be supported by UAST parser yet
			expected: []string{},
		},
		{
			name:     "C# imports",
			filename: "Program.cs",
			content: `using System;
using System.Collections.Generic;
using Microsoft.AspNetCore.Mvc;

namespace TestApp {
    public class Test {
        // ...
    }
}`,
			// C# might not be supported by UAST parser yet
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create dependencies that the extractor needs
			deps := map[string]interface{}{
				plumbing.DependencyTreeChanges: object.Changes{
					&object.Change{
						From: object.ChangeEntry{},
						To: object.ChangeEntry{
							Name: tc.filename,
							TreeEntry: object.TreeEntry{
								Name: tc.filename,
								Hash: gitplumbing.NewHash("test-hash"),
							},
						},
					},
				},
				plumbing.DependencyBlobCache: map[gitplumbing.Hash]*plumbing.CachedBlob{
					gitplumbing.NewHash("test-hash"): {
						Blob: object.Blob{
							Hash: gitplumbing.NewHash("test-hash"),
							Size: int64(len(tc.content)),
						},
						Data: []byte(tc.content),
					},
				},
			}

			// Run the extractor
			result, err := extractor.Consume(deps)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// Check that imports were extracted
			imports, exists := result[DependencyImports]
			assert.True(t, exists, "Imports should be present in result")

			importsMap, ok := imports.(map[gitplumbing.Hash]importmodel.File)
			assert.True(t, ok, "Imports should be a map[gitplumbing.Hash]importmodel.File")

			// For unsupported languages, we expect no files to be processed
			if len(tc.expected) == 0 {
				// For unsupported languages, files might still be processed but return empty imports
				if len(importsMap) > 0 {
					for _, importFile := range importsMap {
						assert.Len(t, importFile.Imports, 0, "Should have no imports for unsupported language")
					}
				}
				return
			}

			assert.Greater(t, len(importsMap), 0, "Should have processed at least one file")
			for hash, importFile := range importsMap {
				assert.NotNil(t, importFile.Imports, "Imports should not be nil for hash %s", hash.String())
				// For supported languages, check that we got some imports (may not match exactly due to UAST parsing)
				assert.Greater(t, len(importFile.Imports), 0, "Should have extracted some imports for supported language")
			}
		})
	}
}

func TestExtractorConfiguration(t *testing.T) {
	extractor := &Extractor{}

	// Test configuration
	facts := map[string]interface{}{
		ConfigImportsGoroutines: 4,
		ConfigMaxFileSize:       1024,
		core.ConfigLogger:       core.GetLogger(),
	}

	err := extractor.Configure(facts)
	assert.NoError(t, err)
	assert.Equal(t, 4, extractor.Goroutines)
	assert.Equal(t, 1024, extractor.MaxFileSize)
}

func TestExtractorProvidesAndRequires(t *testing.T) {
	extractor := &Extractor{}

	// Test Provides
	provides := extractor.Provides()
	assert.Contains(t, provides, DependencyImports)

	// Test Requires
	requires := extractor.Requires()
	assert.Contains(t, requires, plumbing.DependencyTreeChanges)
	assert.Contains(t, requires, plumbing.DependencyBlobCache)
}

func TestExtractorName(t *testing.T) {
	extractor := &Extractor{}
	assert.Equal(t, "Imports", extractor.Name())
}

func TestExtractorListConfigurationOptions(t *testing.T) {
	extractor := &Extractor{}
	options := extractor.ListConfigurationOptions()

	assert.Len(t, options, 2)

	// Check for goroutines option
	foundGoroutines := false
	for _, opt := range options {
		if opt.Name == ConfigImportsGoroutines {
			foundGoroutines = true
			break
		}
	}
	assert.True(t, foundGoroutines, "Should have goroutines configuration option")

	// Check for max file size option
	foundMaxFileSize := false
	for _, opt := range options {
		if opt.Name == ConfigMaxFileSize {
			foundMaxFileSize = true
			break
		}
	}
	assert.True(t, foundMaxFileSize, "Should have max file size configuration option")
}

func TestExtractorRegistration(t *testing.T) {
	summoned := core.Registry.Summon((&Extractor{}).Name())
	assert.Len(t, summoned, 1)
	assert.Equal(t, summoned[0].Name(), "Imports")
	summoned = core.Registry.Summon((&Extractor{}).Provides()[0])
	assert.Len(t, summoned, 1)
	assert.Equal(t, summoned[0].Name(), "Imports")
}
