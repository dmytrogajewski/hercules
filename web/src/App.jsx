import React, { useState, useEffect, useCallback } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useToast } from '@/components/ui/use-toast.jsx'
import { Code, Search, FileText, Play, Zap, X, Settings, ChevronDown, ChevronUp, Plus, Trash2 } from 'lucide-react'
import ReactJsonView from '@microlink/react-json-view'

function App() {
  const [language, setLanguage] = useState('go')
  const [code, setCode] = useState('')
  const [uastOutput, setUastOutput] = useState('No UAST data yet. Start typing code to parse automatically.')
  const [queryInput, setQueryInput] = useState('')
  const [queryOutput, setQueryOutput] = useState('No query results yet. Enter a query to see results.')
  const [isParsing, setIsParsing] = useState(false)
  const [isQuerying, setIsQuerying] = useState(false)
  const [parseTimeout, setParseTimeout] = useState(null)
  const [queryTimeout, setQueryTimeout] = useState(null)
  const [showExamples, setShowExamples] = useState(false)
  const [showCustomMappings, setShowCustomMappings] = useState(false)
  const [customMappings, setCustomMappings] = useState([])
  const [showRawJson, setShowRawJson] = useState(false)
  const { toast } = useToast()

  const languages = [
    { value: 'go', label: 'Go' },
    { value: 'python', label: 'Python' },
    { value: 'javascript', label: 'JavaScript' },
    { value: 'typescript', label: 'TypeScript' },
    { value: 'java', label: 'Java' },
    { value: 'cpp', label: 'C++' },
    { value: 'c', label: 'C' },
    { value: 'rust', label: 'Rust' },
    { value: 'ruby', label: 'Ruby' },
    { value: 'php', label: 'PHP' },
    { value: 'csharp', label: 'C#' },
    { value: 'kotlin', label: 'Kotlin' },
    { value: 'swift', label: 'Swift' },
    { value: 'scala', label: 'Scala' },
    { value: 'dart', label: 'Dart' },
    { value: 'lua', label: 'Lua' },
    { value: 'bash', label: 'Bash' },
    { value: 'html', label: 'HTML' },
    { value: 'css', label: 'CSS' },
    { value: 'json', label: 'JSON' },
    { value: 'yaml', label: 'YAML' },
    { value: 'xml', label: 'XML' },
    { value: 'sql', label: 'SQL' }
  ]

  const exampleQueries = [
    {
      title: 'Find all Import nodes',
      query: 'rfilter(.type == "Import")'
    },
    {
      title: 'Find all Function declarations',
      query: 'rfilter(.type == "Function")'
    },
    {
      title: 'Find all nodes with Import role',
      query: 'rfilter(.roles has "Import")'
    },
    {
      title: 'Get all import tokens',
      query: 'rfilter(.type == "Import") |> rmap(.token)'
    },
    {
      title: 'Find all identifiers',
      query: 'rfilter(.type == "Identifier")'
    }
  ]

  const exampleCustomMappings = [
    {
      name: 'custom_json',
      extensions: ['.json'],
      uast: `[language "json", extensions: ".json"]

_value <- (_value) => uast(
    type: "CustomValue"
)

array <- (array) => uast(
    token: "self",
    type: "CustomArray"
)

document <- (document) => uast(
    type: "CustomDocument"
)

object <- (object) => uast(
    token: "self",
    type: "CustomObject"
)

pair <- (pair) => uast(
    type: "CustomPair",
    children: "_value", "string"
)

string <- (string) => uast(
    token: "self",
    type: "CustomString"
)`
    },
    {
      name: 'simple_config',
      extensions: ['.config', '.cfg'],
      uast: `[language "json", extensions: ".config", ".cfg"]

_value <- (_value) => uast(
    type: "ConfigValue"
)

document <- (document) => uast(
    type: "ConfigDocument"
)

object <- (object) => uast(
    token: "self",
    type: "ConfigObject"
)

pair <- (pair) => uast(
    type: "ConfigPair",
    children: "_value", "string"
)

string <- (string) => uast(
    token: "self",
    type: "ConfigString"
)`
    }
  ]

  const exampleCode = {
    'go': `package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Hello, World!")
}`,
    'python': `import os
import sys
from typing import List

def main():
    print("Hello, World!")
    
if __name__ == "__main__":
    main()`,
    'javascript': `import { readFile } from 'fs';
import path from 'path';

function main() {
    console.log("Hello, World!");
}

export default main;`,
    'java': `import java.util.List;
import java.util.ArrayList;

public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
    }
}`,
    'cpp': `#include <iostream>
#include <vector>

int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}`,
    'rust': `use std::io;

fn main() {
    println!("Hello, World!");
}`,
    'ruby': `require 'json'
require 'net/http'

def main
  puts "Hello, World!"
end

main`,
    'php': `<?php

use Symfony\\Component\\HttpFoundation\\Request;

function main() {
    echo "Hello, World!";
}

main();`,
    'csharp': `using System;
using System.Collections.Generic;

namespace MyApp
{
    class Program
    {
        static void Main(string[] args)
        {
            Console.WriteLine("Hello, World!");
        }
    }
}`,
    'kotlin': `import kotlin.io.println

fun main() {
    println("Hello, World!")
}`,
    'swift': `import Foundation

func main() {
    print("Hello, World!")
}

main()`,
    'scala': `import scala.io.StdIn

object Main {
  def main(args: Array[String]): Unit = {
    println("Hello, World!")
  }
}`,
    'dart': `import 'dart:io';

void main() {
  print('Hello, World!');
}`,
    'lua': `local io = require("io")

function main()
    print("Hello, World!")
end

main()`,
    'bash': `#!/bin/bash

echo "Hello, World!"`,
    'html': `<!DOCTYPE html>
<html>
<head>
    <title>Hello World</title>
</head>
<body>
    <h1>Hello, World!</h1>
</body>
</html>`,
    'css': `body {
    font-family: Arial, sans-serif;
    margin: 0;
    padding: 20px;
}

h1 {
    color: #333;
}`,
    'json': `{
  "name": "example",
  "version": "1.0.0",
  "dependencies": {
    "express": "^4.17.1"
  }
}`,
    'yaml': `name: example
version: 1.0.0
dependencies:
  express: ^4.17.1`,
    'xml': `<?xml version="1.0" encoding="UTF-8"?>
<root>
    <item>Hello, World!</item>
</root>`,
    'sql': `SELECT * FROM users 
WHERE age > 18 
ORDER BY name;`
  }

  useEffect(() => {
    setCode(exampleCode[language] || '// Enter your code here...')
  }, [language])

  // Re-parse when language changes (might affect custom mappings)
  useEffect(() => {
    if (code.trim()) {
      debouncedParse(code)
    }
  }, [language, customMappings])

  // Debounced parsing function
  const debouncedParse = useCallback((codeToParse) => {
    if (parseTimeout) {
      clearTimeout(parseTimeout)
    }

    const timeout = setTimeout(() => {
      if (codeToParse.trim()) {
        parseCode(codeToParse)
      } else {
        setUastOutput('No UAST data yet. Start typing code to parse automatically.')
        // Clear query output when there's no UAST data
        if (queryOutput !== 'No query results yet. Enter a query to see results.') {
          setQueryOutput('No UAST data available. Parse some code first.')
        }
      }
    }, 1000) // 1 second debounce

    setParseTimeout(timeout)
  }, [parseTimeout, queryOutput])

  // Debounced query function
  const debouncedQuery = useCallback((queryToExecute) => {
    if (queryTimeout) {
      clearTimeout(queryTimeout)
    }

    const timeout = setTimeout(() => {
      if (queryToExecute.trim() && uastOutput !== 'No UAST data yet. Start typing code to parse automatically.' && uastOutput !== 'No UAST data available. Parse some code first.') {
        executeQuery(queryToExecute)
      } else {
        setQueryOutput('No query results yet. Enter a query to see results.')
      }
    }, 500) // 0.5 second debounce for queries

    setQueryTimeout(timeout)
  }, [queryTimeout, uastOutput])

  // Handle code changes with debouncing
  const handleCodeChange = (newCode) => {
    setCode(newCode)
    debouncedParse(newCode)
  }

  // Handle query changes with debouncing
  const handleQueryChange = (newQuery) => {
    setQueryInput(newQuery)
    debouncedQuery(newQuery)
  }

  // Handle custom mappings changes - trigger re-parse if code exists
  const handleCustomMappingsChange = (newMappings) => {
    setCustomMappings(newMappings)
    // If we have code, re-parse with new mappings
    if (code.trim()) {
      debouncedParse(code)
    }
  }

  // Update custom mapping and trigger re-parse if needed
  const updateCustomMapping = (index, field, value) => {
    const updated = [...customMappings]
    updated[index] = { ...updated[index], [field]: value }
    setCustomMappings(updated)
    
    // If we have code, re-parse with updated mappings
    if (code.trim()) {
      debouncedParse(code)
    }
  }

  // Add custom mapping and trigger re-parse if needed
  const addCustomMapping = () => {
    const newMapping = {
      name: `mapping_${customMappings.length + 1}`,
      extensions: ['.custom'],
      uast: `[language "json", extensions: ".custom"]

_value <- (_value) => uast(
    type: "Synthetic"
)

document <- (document) => uast(
    type: "Synthetic"
)`
    }
    const newMappings = [...customMappings, newMapping]
    setCustomMappings(newMappings)
    
    // If we have code, re-parse with new mappings
    if (code.trim()) {
      debouncedParse(code)
    }
  }

  // Remove custom mapping and trigger re-parse if needed
  const removeCustomMapping = (index) => {
    const newMappings = customMappings.filter((_, i) => i !== index)
    setCustomMappings(newMappings)
    
    // If we have code, re-parse with updated mappings
    if (code.trim()) {
      debouncedParse(code)
    }
  }

  // Load example mapping and trigger re-parse if needed
  const loadExampleMapping = (example) => {
    const newMappings = [...customMappings, example]
    setCustomMappings(newMappings)
    
    // If we have code, re-parse with new mappings
    if (code.trim()) {
      debouncedParse(code)
    }
  }

  const parseCode = async (codeToParse = code) => {
    if (!codeToParse.trim()) {
      setUastOutput('No UAST data yet. Start typing code to parse automatically.')
      return
    }

    setIsParsing(true)
    setUastOutput('Parsing code...')

    try {
      // Convert custom mappings to the format expected by the server
      const uastMaps = {}
      customMappings.forEach(mapping => {
        uastMaps[mapping.name] = {
          extensions: mapping.extensions,
          uast: mapping.uast
        }
      })

      const requestBody = {
        code: codeToParse,
        language: language
      }

      // Only include UASTMaps if there are custom mappings
      if (Object.keys(uastMaps).length > 0) {
        requestBody.uastmaps = uastMaps
      }

      const response = await fetch('/api/parse', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody)
      })

      const data = await response.json()

      if (data.error) {
        throw new Error(data.error)
      }

      setUastOutput(formatJSON(data.uast))

    } catch (error) {
      setUastOutput(`Error: ${error.message}`)
    } finally {
      setIsParsing(false)
    }
  }

  const executeQuery = async (queryToExecute = queryInput) => {
    if (uastOutput === 'No UAST data yet. Start typing code to parse automatically.' || uastOutput === 'No UAST data available. Parse some code first.') {
      setQueryOutput('No UAST data available. Parse some code first.')
      return
    }

    if (!queryToExecute.trim()) {
      setQueryOutput('No query results yet. Enter a query to see results.')
      return
    }

    setIsQuerying(true)
    setQueryOutput('Executing query...')

    try {
      const response = await fetch('/api/query', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          uast: uastOutput,
          query: queryToExecute
        })
      })

      const data = await response.json()

      if (data.error) {
        throw new Error(data.error)
      }

      setQueryOutput(formatJSON(data.results))

    } catch (error) {
      setQueryOutput(`Error: ${error.message}`)
    } finally {
      setIsQuerying(false)
    }
  }

  const formatJSON = (jsonString) => {
    try {
      const parsed = JSON.parse(jsonString)
      return JSON.stringify(parsed, null, 2)
    } catch (e) {
      return jsonString
    }
  }

  const handleQueryClick = (query) => {
    setQueryInput(query)
    debouncedQuery(query)
  }

  const LoadingSpinner = () => (
    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2" />
  )

  return (
    <div className="h-screen bg-background flex flex-col">
      {/* Header */}
      <div className="flex-shrink-0 p-4 border-b border-border">
        <div className="text-center">
          <h1 className="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent mb-1">
            UAST Mapping Development Service
          </h1>
          <p className="text-muted-foreground text-sm">
            Interactive development environment for UAST mappings
          </p>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col p-4 gap-4 min-h-0">
        {/* Language Selector and Custom Mappings Toggle */}
        <div className="flex-shrink-0">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <Label htmlFor="language" className="text-sm font-medium">Language:</Label>
              <Select value={language} onValueChange={setLanguage}>
                <SelectTrigger className="w-48">
                  <SelectValue placeholder="Select a language" />
                </SelectTrigger>
                <SelectContent>
                  {languages.map((lang) => (
                    <SelectItem key={lang.value} value={lang.value}>
                      {lang.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowCustomMappings(!showCustomMappings)}
              className="flex items-center gap-2"
            >
              <Settings className="h-4 w-4" />
              Custom Mappings
              {showCustomMappings ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
            </Button>
            
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              {(isParsing || isQuerying) && (
                <>
                  <LoadingSpinner />
                  <span>{isParsing ? 'Parsing...' : 'Querying...'}</span>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Custom Mappings Panel */}
        {showCustomMappings && (
          <Card className="flex-shrink-0">
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2 text-sm">
                  <Settings className="h-4 w-4" />
                  Custom UAST Mappings
                </CardTitle>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={addCustomMapping}
                    className="flex items-center gap-2"
                  >
                    <Plus className="h-4 w-4" />
                    Add Mapping
                  </Button>
                </div>
              </div>
              <CardDescription className="text-xs">
                Define custom UAST mappings that will override built-in parsers
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Example Mappings */}
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">Load Example:</span>
                {exampleCustomMappings.map((example, index) => (
                  <Button
                    key={index}
                    variant="outline"
                    size="sm"
                    onClick={() => loadExampleMapping(example)}
                    className="text-xs"
                  >
                    {example.name}
                  </Button>
                ))}
              </div>

              {/* Custom Mappings List */}
              <div className="space-y-3">
                {customMappings.map((mapping, index) => (
                  <Card key={index} className="border-dashed">
                    <CardContent className="pt-4">
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-4">
                          <Input
                            value={mapping.name}
                            onChange={(e) => updateCustomMapping(index, 'name', e.target.value)}
                            placeholder="Mapping name"
                            className="w-32 text-sm"
                          />
                          <Input
                            value={mapping.extensions.join(', ')}
                            onChange={(e) => updateCustomMapping(index, 'extensions', e.target.value.split(',').map(ext => ext.trim()))}
                            placeholder="Extensions (.ext1, .ext2)"
                            className="w-48 text-sm"
                          />
                        </div>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => removeCustomMapping(index)}
                          className="text-red-500 hover:text-red-700"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                      <textarea
                        value={mapping.uast}
                        onChange={(e) => updateCustomMapping(index, 'uast', e.target.value)}
                        placeholder="Enter UAST mapping DSL..."
                        className="w-full h-32 p-2 text-xs font-mono bg-background border rounded-md resize-none focus:outline-none focus:ring-2 focus:ring-ring"
                      />
                    </CardContent>
                  </Card>
                ))}
              </div>

              {customMappings.length === 0 && (
                <div className="text-center text-muted-foreground text-sm py-4">
                  No custom mappings defined. Add one to override built-in parsers.
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Main Grid */}
        <div className="flex-1 grid grid-cols-1 lg:grid-cols-2 gap-4 min-h-0">
          {/* Code Input */}
          <Card className="flex flex-col">
            <CardHeader className="flex-shrink-0 pb-3">
              <CardTitle className="flex items-center gap-2 text-sm">
                <Code className="h-4 w-4" />
                Code Input
              </CardTitle>
              <CardDescription className="text-xs">
                Type code to parse automatically
              </CardDescription>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col min-h-0">
              <div className="flex-1 min-h-0">
                <textarea
                  value={code}
                  onChange={(e) => handleCodeChange(e.target.value)}
                  className="w-full h-full p-3 text-sm font-mono bg-background border rounded-md resize-none focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder="Enter your code here..."
                />
              </div>
            </CardContent>
          </Card>

          {/* Shared Output Area */}
          <Card className="flex flex-col">
            <CardHeader className="flex-shrink-0 pb-3">
                          <CardTitle className="flex items-center gap-2 text-sm">
              <FileText className="h-4 w-4" />
              UAST Output & Query Results
            </CardTitle>
            <div className="flex items-center gap-2">
              <Button
                variant={showRawJson ? "default" : "outline"}
                size="sm"
                onClick={() => setShowRawJson(!showRawJson)}
                className="text-xs"
              >
                {showRawJson ? "Tree View" : "Raw JSON"}
              </Button>
            </div>
              <CardDescription className="text-xs">
                Parsed UAST structure and query results
              </CardDescription>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col min-h-0 space-y-3">
              {/* Query Input */}
              <div className="flex-shrink-0">
                <Input
                  value={queryInput}
                  onChange={(e) => handleQueryChange(e.target.value)}
                  placeholder="Enter UAST query (e.g., rfilter(.type == 'Import'))"
                  className="text-sm"
                />
              </div>

              {/* Output Display */}
              <div className="flex-1 min-h-0 relative">
                <div className="h-full overflow-auto">
                  {(() => {
                    // Show loading state
                    if (isParsing) {
                      return (
                        <div className="flex items-center justify-center h-full text-muted-foreground">
                          <LoadingSpinner />
                          <span className="ml-2">Parsing code...</span>
                        </div>
                      )
                    }
                    if (isQuerying) {
                      return (
                        <div className="flex items-center justify-center h-full text-muted-foreground">
                          <LoadingSpinner />
                          <span className="ml-2">Executing query...</span>
                        </div>
                      )
                    }
                    // Show query results if available and not empty
                    if (queryOutput !== 'No query results yet. Enter a query to see results.' && 
                        queryOutput !== 'No UAST data available. Parse some code first.' &&
                        queryOutput.trim()) {
                      return (
                        <div className="p-3">
                          {showRawJson ? (
                            <pre className="text-xs font-mono whitespace-pre-wrap">
                              {queryOutput}
                            </pre>
                          ) : (
                            <ReactJsonView
                              src={(() => {
                                try {
                                  return typeof queryOutput === 'string' ? JSON.parse(queryOutput) : queryOutput
                                } catch (e) {
                                  return { error: 'Invalid JSON', raw: queryOutput }
                                }
                              })()}
                              theme="monokai"
                              collapsed={2}
                              displayDataTypes={false}
                              displayObjectSize={true}
                              enableClipboard={true}
                              style={{ backgroundColor: 'transparent' }}
                            />
                          )}
                        </div>
                      )
                    }
                    // Show UAST output if available
                    if (uastOutput !== 'No UAST data yet. Start typing code to parse automatically.' &&
                        uastOutput !== 'No UAST data available. Parse some code first.') {
                      return (
                        <div className="p-3">
                          {showRawJson ? (
                            <pre className="text-xs font-mono whitespace-pre-wrap">
                              {uastOutput}
                            </pre>
                          ) : (
                            <ReactJsonView
                              src={(() => {
                                try {
                                  return typeof uastOutput === 'string' ? JSON.parse(uastOutput) : uastOutput
                                } catch (e) {
                                  return { error: 'Invalid JSON', raw: uastOutput }
                                }
                              })()}
                              theme="monokai"
                              collapsed={2}
                              displayDataTypes={false}
                              displayObjectSize={true}
                              enableClipboard={true}
                              style={{ backgroundColor: 'transparent' }}
                            />
                          )}
                        </div>
                      )
                    }
                    // Default message
                    return (
                      <div className="flex items-center justify-center h-full text-muted-foreground">
                        No UAST data yet. Start typing code to parse automatically.
                      </div>
                    )
                  })()}
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Floating Examples Panel */}
      {showExamples && (
        <div className="fixed bottom-4 left-4 right-4 z-50">
          <Card className="shadow-2xl">
            <CardHeader className="flex flex-row items-center justify-between pb-3">
              <CardTitle className="flex items-center gap-2 text-sm">
                <Zap className="h-4 w-4" />
                Example Queries
              </CardTitle>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setShowExamples(false)}
                className="h-6 w-6 p-0"
              >
                <X className="h-4 w-4" />
              </Button>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-5 gap-2">
                {exampleQueries.map((example, index) => (
                  <Button
                    key={index}
                    variant="outline"
                    size="sm"
                    className="text-xs h-auto p-2 flex flex-col items-start gap-1"
                    onClick={() => handleQueryClick(example.query)}
                  >
                    <span className="font-medium">{example.title}</span>
                    <code className="text-xs opacity-70 break-all">
                      {example.query}
                    </code>
                  </Button>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Floating Examples Button */}
      {!showExamples && (
        <div className="fixed bottom-4 right-4 z-40">
          <Button
            onClick={() => setShowExamples(true)}
            className="rounded-full w-12 h-12 shadow-lg"
          >
            <Zap className="h-5 w-5" />
          </Button>
        </div>
      )}


    </div>
  )
}

export default App 