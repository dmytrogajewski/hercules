package extractor

import (
	"testing"

	"github.com/dmytrogajewski/hercules/internal/importmodel"
	"github.com/stretchr/testify/assert"
)

func TestExtractGoImports(t *testing.T) {
	code := `
package main

import "fmt"
import "strings"

import (
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Hello")
}
`
	file, err := Extract("test.go", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "golang", file.Lang)
	assert.Contains(t, file.Imports, "fmt")
	assert.Contains(t, file.Imports, "strings")
	assert.Contains(t, file.Imports, "os")
	assert.Contains(t, file.Imports, "path/filepath")
}

func TestExtractPythonImports(t *testing.T) {
	code := `
import os
import sys
from typing import List, Dict
from pathlib import Path

def main():
    print("Hello")
`
	file, err := Extract("test.py", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "python", file.Lang)
	assert.Contains(t, file.Imports, "os")
	assert.Contains(t, file.Imports, "sys")
	assert.Contains(t, file.Imports, "typing")
	assert.Contains(t, file.Imports, "pathlib")
}

func TestExtractJavaScriptImports(t *testing.T) {
	code := `
import React from 'react';
import { useState, useEffect } from 'react';
import './styles.css';

function App() {
    return <div>Hello</div>;
}
`
	file, err := Extract("test.js", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "javascript", file.Lang)
	assert.Contains(t, file.Imports, "react")
	assert.Contains(t, file.Imports, "./styles.css")
}

func TestExtractJavaImports(t *testing.T) {
	code := `
import java.util.List;
import java.util.Map;
import org.springframework.stereotype.Component;

@Component
public class Test {
    // ...
}
`
	file, err := Extract("Test.java", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "java", file.Lang)
	assert.Contains(t, file.Imports, "java.util.List")
	assert.Contains(t, file.Imports, "java.util.Map")
	assert.Contains(t, file.Imports, "org.springframework.stereotype.Component")
}

func TestExtractCppImports(t *testing.T) {
	code := `
#include <iostream>
#include <vector>
#include "myheader.h"

int main() {
    return 0;
}
`
	file, err := Extract("test.cpp", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "cpp", file.Lang)
	assert.Contains(t, file.Imports, "iostream")
	assert.Contains(t, file.Imports, "vector")
	assert.Contains(t, file.Imports, "myheader.h")
}

func TestExtractCSharpImports(t *testing.T) {
	code := `
using System;
using System.Collections.Generic;
using Microsoft.AspNetCore.Mvc;

namespace TestApp {
    public class Test {
        // ...
    }
}
`
	file, err := Extract("Test.cs", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "csharp", file.Lang)
	assert.Contains(t, file.Imports, "System")
	assert.Contains(t, file.Imports, "System.Collections.Generic")
	assert.Contains(t, file.Imports, "Microsoft.AspNetCore.Mvc")
}

func TestExtractRustImports(t *testing.T) {
	code := `
use std::collections::HashMap;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize)]
struct Test {
    // ...
}
`
	file, err := Extract("test.rs", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "rust", file.Lang)
	assert.Contains(t, file.Imports, "std::collections::HashMap")
	assert.Contains(t, file.Imports, "serde::{Deserialize, Serialize}")
}

func TestExtractRubyImports(t *testing.T) {
	code := `
require 'json'
require 'net/http'
load 'config.rb'

class Test
  # ...
end
`
	file, err := Extract("test.rb", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "ruby", file.Lang)
	assert.Contains(t, file.Imports, "json")
	assert.Contains(t, file.Imports, "net/http")
	assert.Contains(t, file.Imports, "config.rb")
}

func TestExtractPhpImports(t *testing.T) {
	code := `<?php
use Symfony\Component\HttpFoundation\Request;
use App\Services\UserService;

class Test {
    // ...
}
`
	file, err := Extract("test.php", []byte(code))
	assert.NoError(t, err)
	assert.Equal(t, "php", file.Lang)
	assert.Contains(t, file.Imports, "Symfony\\Component\\HttpFoundation\\Request")
	assert.Contains(t, file.Imports, "App\\Services\\UserService")
}

func TestFileTypeVisible(t *testing.T) {
	var _ importmodel.File
}
