# Development Guide

This guide covers everything you need to know to contribute to Hercules development.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Building](#building)
- [Testing](#testing)
- [Code Style](#code-style)
- [Adding New Features](#adding-new-features)
- [Debugging](#debugging)
- [Performance](#performance)
- [Release Process](#release-process)

## Prerequisites

### Required Software

- **Go 1.21+**: [Download](https://golang.org/dl/)
- **Git**: [Download](https://git-scm.com/downloads)
- **Make**: Usually pre-installed on Unix systems
- **Docker** (optional): [Download](https://www.docker.com/products/docker-desktop)

### Recommended Tools

- **VS Code** with Go extension
- **GoLand** or **Vim/Emacs** with Go support
- **Delve** debugger: `go install github.com/go-delve/delve/cmd/dlv@latest`

## Development Setup

### 1. Clone the Repository

```bash
git clone https://github.com/dmytrogajewski/hercules.git
cd hercules
```

### 2. Install Dependencies

```bash
# Download Go modules
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

### 3. Build the Project

```bash
# Build binary
make

# Build with specific tags
make build TAGS="unit_tests"

# Install (optional)
sudo make install
```

### 4. Verify Setup

```bash
# Run tests
make test

# Run linter
make lint

# Check build
./hercules --help
```

## Project Structure

```
hercules/
├── cmd/hercules/           # Main application entry point
│   ├── root.go            # Root command and CLI setup
│   └── server.go          # HTTP server implementation
├── internal/              # Internal packages (not importable)
│   ├── config/           # Configuration management
│   ├── core/             # Core analysis engine
│   ├── plumbing/         # Git plumbing utilities
│   ├── burndown/         # Burndown analysis
│   ├── mathutil/         # Math utilities
│   ├── pb/               # Protocol buffer definitions
│   └── ...
├── leaves/               # Analysis implementations
│   ├── burndown.go       # Burndown analysis
│   ├── couples.go        # Couples analysis
│   ├── devs.go          # Developer analysis
│   └── ...
├── doc/                 # Documentation
├── docs/                # Additional documentation
├── contrib/             # Contrib plugins
├── Makefile             # Build configuration
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── README.md            # Project README
```

### Key Components

#### `cmd/hercules/`
Contains the main application entry point and CLI commands.

#### `internal/`
Internal packages that are not meant to be imported by external code.

- **`config/`**: Configuration management using Viper
- **`core/`**: Core analysis engine and pipeline
- **`plumbing/`**: Git utilities and repository handling
- **`pb/`**: Protocol buffer definitions for data serialization

#### `leaves/`
Analysis implementations that can be deployed in the pipeline.

## Building

### Basic Build

```bash
# Build binary
make

# Build with specific architecture
GOOS=linux GOARCH=amd64 make

# Build with debug information
make build DEBUG=1
```

### Build Tags

```bash
# Build with unit tests only
make build TAGS="unit_tests"

# Build with all features
make build TAGS="all"

# Build for production
make build TAGS="production"
```

### Cross-Compilation

```bash
# Build for multiple platforms
make build-all

# Build for specific platform
GOOS=windows GOARCH=amd64 make build
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with verbose output
make testv

# Run specific package tests
go test ./internal/config

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration
```

### Test Structure

```bash
# Unit tests
go test ./internal/config
go test ./internal/core
go test ./leaves

# Integration tests
go test ./cmd/hercules -tags=integration

# Benchmark tests
go test -bench=. ./internal/core
```

### Writing Tests

#### Unit Tests

```go
// internal/config/config_test.go
package config

import (
    "testing"
    "time"
)

func TestLoadConfigDefaults(t *testing.T) {
    cfg, err := LoadConfig("")
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }
    
    if cfg.Server.Port != 8080 {
        t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
    }
}
```

#### Integration Tests

```go
// cmd/hercules/server_integration_test.go
//go:build integration

package main

import (
    "testing"
    "net/http"
    "net/http/httptest"
)

func TestServerIntegration(t *testing.T) {
    // Test server endpoints
}
```

### Test Best Practices

1. **Use descriptive test names**: `TestLoadConfigWithValidFile`
2. **Test both success and failure cases**
3. **Use table-driven tests for multiple scenarios**
4. **Mock external dependencies**
5. **Test edge cases and error conditions**

## Code Style

### Go Conventions

Follow [Effective Go](https://golang.org/doc/effective_go.html) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

#### Naming

```go
// Good
func LoadConfig(path string) (*Config, error)
var maxConcurrentAnalyses int

// Bad
func load_config(path string) (*config, error)
var MaxConcurrentAnalyses int
```

#### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Bad
if err != nil {
    return err
}
```

#### Comments

```go
// LoadConfig loads configuration from file and environment variables.
// It supports multiple configuration sources with sensible defaults.
func LoadConfig(configPath string) (*Config, error) {
    // Implementation...
}
```

### Linting

```bash
# Run linter
make lint

# Run linter with specific rules
golangci-lint run --enable=goimports,unused

# Fix imports
goimports -w .
```

### Pre-commit Hooks

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
set -e

# Run tests
make test

# Run linter
make lint

# Check formatting
goimports -d .
```

## Adding New Features

### 1. Create a Feature Branch

```bash
git checkout -b feature/new-analysis
```

### 2. Implement the Feature

#### Adding a New Analysis

1. **Create the analysis in `leaves/`**:

```go
// leaves/new_analysis.go
package leaves

import (
    "github.com/dmytrogajewski/hercules/internal/core"
)

type NewAnalysis struct {
    core.NoopMerger
    // Analysis-specific fields
}

func (na *NewAnalysis) Name() string {
    return "NewAnalysis"
}

func (na *NewAnalysis) Provides() []string {
    return []string{"new_analysis"}
}

func (na *NewAnalysis) Requires() []string {
    return []string{}
}

func (na *NewAnalysis) Consume(deps map[string]interface{}) (map[string]interface{}, error) {
    // Implementation
    return map[string]interface{}{"new_analysis": result}, nil
}
```

2. **Add command-line support in `cmd/hercules/root.go`**:

```go
// Add flag
newAnalysisCmd.Flags().Bool("new-analysis", false, "Enable new analysis")

// Add to pipeline
if newAnalysis {
    pipeline.DeployItem(&leaves.NewAnalysis{})
}
```

3. **Add server support in `cmd/hercules/server.go`**:

```go
case "new-analysis":
    item := pipeline.DeployItem(&leaves.NewAnalysis{}).(hercules.LeafPipelineItem)
    deployed = append(deployed, item)
```

#### Adding Configuration Options

1. **Add to `internal/config/config.go`**:

```go
type AnalysisConfig struct {
    // Existing fields...
    NewAnalysisOption string `mapstructure:"new_analysis_option"`
}
```

2. **Add defaults in `setDefaults`**:

```go
v.SetDefault("analysis.new_analysis_option", "default_value")
```

3. **Add validation**:

```go
if config.Analysis.NewAnalysisOption == "" {
    return fmt.Errorf("new analysis option cannot be empty")
}
```

### 3. Write Tests

```go
// leaves/new_analysis_test.go
func TestNewAnalysis(t *testing.T) {
    analysis := &NewAnalysis{}
    
    // Test Name()
    if analysis.Name() != "NewAnalysis" {
        t.Errorf("Expected Name() to return 'NewAnalysis', got %s", analysis.Name())
    }
    
    // Test Provides()
    provides := analysis.Provides()
    expected := []string{"new_analysis"}
    if !reflect.DeepEqual(provides, expected) {
        t.Errorf("Expected Provides() to return %v, got %v", expected, provides)
    }
}
```

### 4. Update Documentation

1. **Update README.md** with new features
2. **Update API documentation** if adding endpoints
3. **Update configuration documentation** if adding options

### 5. Submit Pull Request

```bash
# Commit changes
git add .
git commit -m "feat: add new analysis feature"

# Push branch
git push origin feature/new-analysis
```

## Debugging

### Using Delve

```bash
# Debug the server
dlv debug ./cmd/hercules -- server

# Debug with specific configuration
dlv debug ./cmd/hercules -- server -c config.yaml

# Set breakpoints
(dlv) break main.main
(dlv) break internal/config.LoadConfig
```

### Using VS Code

Create `.vscode/launch.json`:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Server",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/hercules",
            "args": ["server", "-c", "config.yaml"]
        }
    ]
}
```

### Logging

```go
import "log"

// Debug logging
log.Printf("Loading configuration from: %s", configPath)

// Error logging
log.Printf("Failed to load config: %v", err)
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.

# Memory profiling
go test -memprofile=mem.prof -bench=.

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Performance

### Benchmarking

```go
// internal/core/pipeline_test.go
func BenchmarkPipeline(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Benchmark code
    }
}
```

### Profiling

```bash
# Run with profiling
./hercules --burndown https://github.com/dmytrogajewski/hercules.git 2> profile.out

# Analyze profile
go tool pprof profile.out
```

### Optimization Tips

1. **Use buffered channels** for high-throughput operations
2. **Pool objects** to reduce GC pressure
3. **Use sync.Pool** for frequently allocated objects
4. **Profile before optimizing**
5. **Measure the impact** of optimizations

## Release Process

### 1. Version Bumping

```bash
# Update version in version.go
git tag v1.0.0
git push origin v1.0.0
```

### 2. Building Releases

```bash
# Build for all platforms
make release

# Build specific platform
GOOS=linux GOARCH=amd64 make build
```

### 3. Creating Release Notes

1. **Summarize changes** since last release
2. **List new features** and improvements
3. **Document breaking changes**
4. **Include migration guide** if needed

### 4. Publishing

1. **Create GitHub release** with release notes
2. **Upload binaries** for all platforms
3. **Update documentation** if needed
4. **Announce on social media** and mailing lists

## Contributing Guidelines

### Before Contributing

1. **Read the documentation**
2. **Check existing issues** and pull requests
3. **Discuss major changes** in an issue first
4. **Follow the code style** and conventions

### Pull Request Process

1. **Create a feature branch** from `main`
2. **Write tests** for new functionality
3. **Update documentation** as needed
4. **Ensure all tests pass**
5. **Submit pull request** with clear description

### Code Review

1. **Review for correctness** and completeness
2. **Check for security issues**
3. **Verify performance impact**
4. **Ensure documentation is updated**
5. **Test the changes** locally

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/dmytrogajewski/hercules/issues)
- **Discussions**: [GitHub Discussions](https://github.com/dmytrogajewski/hercules/discussions)
- **Documentation**: [Wiki](https://github.com/dmytrogajewski/hercules/wiki)
- **Chat**: [Gitter](https://gitter.im/dmytrogajewski/hercules)

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Testing](https://golang.org/pkg/testing/)
- [Go Modules](https://golang.org/ref/mod) 