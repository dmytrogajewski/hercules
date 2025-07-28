#!/bin/bash

echo "🚀 UAST Mapping Development Service Demo"
echo "========================================"
echo ""

# Check if service is running
if ! curl -s http://localhost:8080/ > /dev/null; then
    echo "❌ Service is not running. Please start it with: make dev-service"
    exit 1
fi

echo "✅ Service is running on http://localhost:8080"
echo ""

# Test 1: Parse Go code
echo "📝 Test 1: Parsing Go code"
echo "---------------------------"
GO_CODE='package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Hello, World!")
}'

echo "Input code:"
echo "$GO_CODE"
echo ""

RESPONSE=$(curl -s -X POST http://localhost:8080/api/parse \
    -H "Content-Type: application/json" \
    -d "{\"code\": \"$GO_CODE\", \"language\": \"go\"}")

if echo "$RESPONSE" | jq -e '.error' > /dev/null; then
    echo "❌ Parse error: $(echo "$RESPONSE" | jq -r '.error')"
else
    echo "✅ Parse successful!"
    UAST=$(echo "$RESPONSE" | jq -r '.uast')
    echo "UAST structure generated (showing first 200 chars):"
    echo "$UAST" | head -c 200
    echo "..."
    echo ""
fi

# Test 2: Query for Import nodes
echo "🔍 Test 2: Querying for Import nodes"
echo "------------------------------------"
if [ -n "$UAST" ]; then
    QUERY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/query \
        -H "Content-Type: application/json" \
        -d "{\"uast\": \"$UAST\", \"query\": \"rfilter(.type == \\\"Import\\\")\"}")

    if echo "$QUERY_RESPONSE" | jq -e '.error' > /dev/null; then
        echo "❌ Query error: $(echo "$QUERY_RESPONSE" | jq -r '.error')"
    else
        echo "✅ Query successful!"
        RESULTS=$(echo "$QUERY_RESPONSE" | jq -r '.results')
        echo "Query results:"
        echo "$RESULTS"
        echo ""
    fi
fi

# Test 3: Query for Function nodes
echo "🔍 Test 3: Querying for Function nodes"
echo "--------------------------------------"
if [ -n "$UAST" ]; then
    QUERY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/query \
        -H "Content-Type: application/json" \
        -d "{\"uast\": \"$UAST\", \"query\": \"rfilter(.type == \\\"Function\\\")\"}")

    if echo "$QUERY_RESPONSE" | jq -e '.error' > /dev/null; then
        echo "❌ Query error: $(echo "$QUERY_RESPONSE" | jq -r '.error')"
    else
        echo "✅ Query successful!"
        RESULTS=$(echo "$QUERY_RESPONSE" | jq -r '.results')
        echo "Query results:"
        echo "$RESULTS"
        echo ""
    fi
fi

# Test 4: Parse Python code
echo "📝 Test 4: Parsing Python code"
echo "------------------------------"
PYTHON_CODE='import os
import sys
from typing import List

def main():
    print("Hello, World!")
    
if __name__ == "__main__":
    main()'

echo "Input code:"
echo "$PYTHON_CODE"
echo ""

RESPONSE=$(curl -s -X POST http://localhost:8080/api/parse \
    -H "Content-Type: application/json" \
    -d "{\"code\": \"$PYTHON_CODE\", \"language\": \"python\"}")

if echo "$RESPONSE" | jq -e '.error' > /dev/null; then
    echo "❌ Parse error: $(echo "$RESPONSE" | jq -r '.error')"
else
    echo "✅ Parse successful!"
    PYTHON_UAST=$(echo "$RESPONSE" | jq -r '.uast')
    echo "UAST structure generated (showing first 200 chars):"
    echo "$PYTHON_UAST" | head -c 200
    echo "..."
    echo ""
fi

# Test 5: Query Python for Pattern nodes (imports)
echo "🔍 Test 5: Querying Python for Pattern nodes (imports)"
echo "-----------------------------------------------------"
if [ -n "$PYTHON_UAST" ]; then
    QUERY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/query \
        -H "Content-Type: application/json" \
        -d "{\"uast\": \"$PYTHON_UAST\", \"query\": \"rfilter(.type == \\\"Pattern\\\")\"}")

    if echo "$QUERY_RESPONSE" | jq -e '.error' > /dev/null; then
        echo "❌ Query error: $(echo "$QUERY_RESPONSE" | jq -r '.error')"
    else
        echo "✅ Query successful!"
        RESULTS=$(echo "$QUERY_RESPONSE" | jq -r '.results')
        echo "Query results:"
        echo "$RESULTS"
        echo ""
    fi
fi

echo "🎉 Demo completed!"
echo ""
echo "🌐 Open http://localhost:8080 in your browser to use the interactive interface"
echo "📚 Check web/README.md for more information" 