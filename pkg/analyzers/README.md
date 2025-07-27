# Analyzers

This package provides comprehensive code analysis tools for Hercules, focusing on code complexity, quality metrics, and maintainability indicators. The architecture follows SOLID principles with modular design and extensive use of common components.

## Architecture

The package follows a clean, modular architecture with each analyzer being self-contained while leveraging shared common modules:

```
pkg/analyzers/
├── analyze/             # Core interfaces and factory
│   └── analyzer.go      # CodeAnalyzer and ResultAggregator interfaces
├── common/              # Shared modules and utilities
│   ├── aggregator.go    # Common aggregation logic
│   ├── data_collector.go # Data collection utilities
│   ├── data_extraction.go # UAST data extraction
│   ├── formatter.go     # Common formatting utilities
│   ├── metrics_processor.go # Metrics processing
│   ├── reporter.go      # Advanced reporting capabilities
│   ├── result_builder.go # Result construction utilities
│   └── uast_traversal.go # UAST traversal and filtering
├── cohesion/            # Code cohesion analysis
│   ├── cohesion.go      # Main analyzer
│   ├── aggregator.go    # Aggregation logic
│   └── cohesion_test.go
├── comments/            # Comment density and quality analysis
│   ├── comments.go      # Main analyzer
│   ├── aggregator.go    # Aggregation logic
│   ├── types.go         # Type definitions
│   └── comments_test.go
├── complexity/          # Cyclomatic and cognitive complexity
│   ├── complexity.go    # Main analyzer
│   ├── aggregator.go    # Aggregation logic
│   └── complexity_test.go
├── halstead/            # Halstead complexity measures
│   ├── halstead.go      # Main analyzer orchestration
│   ├── metrics.go       # Metrics calculation logic
│   ├── detector.go      # Operator/operand detection
│   ├── formatter.go     # Report formatting
│   ├── aggregator.go    # Result aggregation
│   └── halstead_test.go
└── README.md
```

## Core Components

### 1. Core Interfaces (`analyze/`)
- **`CodeAnalyzer`**: Defines the contract for all analyzers
- **`ResultAggregator`**: Defines aggregation contract
- **`Factory`**: Manages analyzer registration and execution

### 2. Common Modules (`common/`)
- **`Aggregator`**: Standardized result aggregation across files
- **`DataExtractor`**: UAST data extraction with custom extractors
- **`Formatter`**: Advanced formatting with progress bars and tables
- **`Reporter`**: Multi-format reporting (text, JSON, summary)
- **`ResultBuilder`**: Structured result construction
- **`UASTTraverser`**: Advanced UAST traversal with filtering
- **`MetricsProcessor`**: Standardized metrics processing

### 3. Analyzers
Each analyzer implements the `CodeAnalyzer` interface and leverages common modules:

- **`CohesionAnalyzer`**: Measures code cohesion and coupling
- **`CommentsAnalyzer`**: Analyzes comment density and quality
- **`ComplexityAnalyzer`**: Measures cyclomatic and cognitive complexity
- **`HalsteadAnalyzer`**: Calculates Halstead complexity measures

## Available Analyzers

### 1. Code Cohesion Analyzer (`cohesion/`)
- **Purpose:** Measures code cohesion and coupling between modules
- **Metrics:**
  - **Cohesion Score:** Measures how closely related functions are
  - **Coupling Score:** Measures dependencies between modules
  - **Modularity Index:** Overall modularity assessment
- **Thresholds:**
  - Cohesion: Green ≥ 0.7, Yellow 0.5-0.7, Red < 0.5
  - Coupling: Green ≤ 0.3, Yellow 0.3-0.5, Red > 0.5

### 2. Comment Density & Quality Analyzer (`comments/`)
- **Purpose:** Measures comment density and evaluates documentation quality
- **Metrics:**
  - **Comment Density:** Comment lines / Total lines
  - **Documentation Coverage:** Documented functions / Total functions
  - **Comment Quality:** Average quality score of all comments
  - **Function Documentation:** Per-function documentation analysis
- **Quality Assessment:**
  - **Detailed:** Comprehensive comments with multiple sentences
  - **Adequate:** Well-structured comments with sufficient explanation
  - **Short:** Brief but meaningful comments
  - **Minimal:** Very brief comments (single words)
  - **Very Short:** Extremely brief comments (< 5 characters)
- **Thresholds:**
  - Comment Density: Green ≥ 25%, Yellow 15-25%, Red < 15%
  - Documentation Score: Green ≥ 0.8, Yellow 0.6-0.8, Red < 0.6
  - Comment Quality: Green ≥ 0.7, Yellow 0.5-0.7, Red < 0.5

### 3. Cyclomatic Complexity Analyzer (`complexity/`)
- **Purpose:** Measures the cyclomatic complexity of functions and methods
- **Metrics:**
  - **Cyclomatic Complexity:** E - N + 2P (edges - nodes + 2*components)
  - **Cognitive Complexity:** How difficult code is to understand
  - **Nesting Depth:** Maximum nesting level of control structures
- **Formula:**
  - Each function starts with base complexity of 1
  - Each decision point (if, for, while, case, catch) adds 1
  - Each logical operator (`&&`, `||`) adds 1
- **Thresholds:**
  - Cyclomatic: Green ≤ 1, Yellow 2-5, Red > 5
  - Cognitive: Green ≤ 1, Yellow 2-7, Red > 7
  - Nesting: Green ≤ 1, Yellow 2-3, Red > 3

### 4. Halstead Complexity Measures (`halstead/`)
- **Purpose:** Measures software complexity based on operators and operands
- **Metrics:**
  - **Program Vocabulary:** η = η₁ + η₂
  - **Program Length:** N = N₁ + N₂
  - **Estimated Length:** N̂ = η₁×log₂(η₁) + η₂×log₂(η₂)
  - **Volume:** V = N × log₂(η)
  - **Difficulty:** D = (η₁/2) × (N₂/η₂)
  - **Effort:** E = D × V
  - **Time to Program:** T = E/18 (seconds)
  - **Delivered Bugs:** B = V/3000
- **Thresholds:**
  - Volume: Green ≤ 100, Yellow 101-1000, Red > 1000
  - Difficulty: Green ≤ 5, Yellow 6-15, Red > 15
  - Effort: Green ≤ 1000, Yellow 1001-10000, Red > 10000

## Common Modules Reference

### 1. Aggregator (`common/aggregator.go`)
Provides standardized result aggregation across multiple files:

```go
// Create aggregator with custom configuration
aggregator := common.NewAggregatorWithCustomEmptyResult(
    "analyzer_name",
    []string{"metric1", "metric2"}, // numeric keys to sum
    []string{"count1", "count2"},   // count keys to sum
    "collection_key",               // collection to aggregate
    messageBuilder,                 // custom message builder
    emptyResultBuilder,             // custom empty result builder
)

// Aggregate results
aggregator.Aggregate(results)
finalResult := aggregator.GetResult()
```

### 2. Data Extractor (`common/data_extraction.go`)
Handles UAST data extraction with custom extractors:

```go
// Configure data extractor
config := common.ExtractionConfig{
    DefaultExtractors: true,
    NameExtractors: map[string]common.NameExtractor{
        "function_name": common.ExtractFunctionName,
        "custom_name":   customExtractor,
    },
}
extractor := common.NewDataExtractor(config)

// Extract data from nodes
name, ok := extractor.ExtractName(node, "function_name")
```

### 3. Formatter (`common/formatter.go`)
Provides advanced formatting capabilities:

```go
// Configure formatter
config := common.FormatConfig{
    ShowProgressBars: true,
    ShowTables:       true,
    ShowDetails:      true,
    SkipHeader:       false,
}
formatter := common.NewFormatter(config)

// Format report
formatted := formatter.FormatReport(report)
```

### 4. Reporter (`common/reporter.go`)
Handles multi-format reporting:

```go
// Configure reporter
config := common.ReportConfig{
    Format:      "text", // text, json, summary
    ShowDetails: true,
    ShowTables:  true,
}
reporter := common.NewReporter(config)

// Generate report
err := reporter.GenerateReport(report, writer)
```

### 5. Result Builder (`common/result_builder.go`)
Constructs structured analysis results:

```go
builder := common.NewResultBuilder()

// Build collection result
result := builder.BuildCollectionResult(
    "analyzer_name",
    "collection_key",
    collectionData,
    metrics,
    message,
)

// Build custom empty result
emptyResult := builder.BuildCustomEmptyResult(map[string]interface{}{
    "metric1": 0,
    "metric2": 0.0,
    "message": "No data found",
})
```

### 6. UAST Traverser (`common/uast_traversal.go`)
Provides advanced UAST traversal with filtering:

```go
// Configure traversal
config := common.TraversalConfig{
    Filters: []common.NodeFilter{
        {
            Types:    []string{node.UASTFunction, node.UASTMethod},
            Roles:    []string{node.RoleFunction, node.RoleDeclaration},
            MinLines: 1,
        },
    },
    MaxDepth:    10,
    IncludeRoot: false,
}
traverser := common.NewUASTTraverser(config)

// Find nodes by type
functions := traverser.FindNodesByType(root, []string{node.UASTFunction})

// Find nodes by roles
declarations := traverser.FindNodesByRoles(root, []string{node.RoleDeclaration})
```

## Implementing a New Analyzer

### Step 1: Create Analyzer Structure
Create a new directory for your analyzer:

```bash
mkdir pkg/analyzers/myanalyzer
cd pkg/analyzers/myanalyzer
```

### Step 2: Implement Core Interface
Create the main analyzer file (`myanalyzer.go`):

```go
package myanalyzer

import (
    "fmt"
    "io"

    "github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
    "github.com/dmytrogajewski/hercules/pkg/analyzers/common"
    "github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

// MyAnalyzer provides custom analysis
type MyAnalyzer struct {
    traverser *common.UASTTraverser
    extractor *common.DataExtractor
    formatter *common.Reporter
}

// NewMyAnalyzer creates a new analyzer instance
func NewMyAnalyzer() *MyAnalyzer {
    // Configure UAST traverser
    traversalConfig := common.TraversalConfig{
        Filters: []common.NodeFilter{
            {
                Types: []string{node.UASTFunction},
                Roles: []string{node.RoleFunction},
            },
        },
        MaxDepth: 10,
    }

    // Configure data extractor
    extractionConfig := common.ExtractionConfig{
        DefaultExtractors: true,
        NameExtractors: map[string]common.NameExtractor{
            "function_name": common.ExtractFunctionName,
        },
    }

    return &MyAnalyzer{
        traverser: common.NewUASTTraverser(traversalConfig),
        extractor: common.NewDataExtractor(extractionConfig),
        formatter: common.NewReporter(common.ReportConfig{
            Format:      "text",
            ShowDetails: true,
        }),
    }
}

// Name returns the analyzer name
func (m *MyAnalyzer) Name() string {
    return "myanalyzer"
}

// Analyze performs the analysis
func (m *MyAnalyzer) Analyze(root *node.Node) (analyze.Report, error) {
    if root == nil {
        return nil, fmt.Errorf("root node is nil")
    }

    // Find relevant nodes
    functions := m.traverser.FindNodesByType(root, []string{node.UASTFunction})
    
    if len(functions) == 0 {
        return m.buildEmptyResult("No functions found"), nil
    }

    // Perform analysis
    results := m.analyzeFunctions(functions)
    
    // Build result using common result builder
    return m.buildResult(results), nil
}

// Thresholds returns metric thresholds
func (m *MyAnalyzer) Thresholds() analyze.Thresholds {
    return analyze.Thresholds{
        "metric1": {
            "green":  10,
            "yellow": 50,
            "red":    100,
        },
    }
}

// CreateAggregator returns a new aggregator
func (m *MyAnalyzer) CreateAggregator() analyze.ResultAggregator {
    return NewMyAggregator()
}

// FormatReport formats human-readable output
func (m *MyAnalyzer) FormatReport(report analyze.Report, w io.Writer) error {
    return m.formatter.GenerateReport(report, w)
}

// FormatReportJSON formats JSON output
func (m *MyAnalyzer) FormatReportJSON(report analyze.Report, w io.Writer) error {
    config := common.ReportConfig{Format: "json"}
    reporter := common.NewReporter(config)
    return reporter.GenerateReport(report, w)
}

// Helper methods
func (m *MyAnalyzer) analyzeFunctions(functions []*node.Node) []map[string]interface{} {
    // Implementation here
    return nil
}

func (m *MyAnalyzer) buildResult(results []map[string]interface{}) analyze.Report {
    metrics := map[string]interface{}{
        "total_functions": len(results),
        "metric1":         42.0,
    }

    return common.NewResultBuilder().BuildCollectionResult(
        "myanalyzer",
        "functions",
        results,
        metrics,
        "Analysis completed successfully",
    )
}

func (m *MyAnalyzer) buildEmptyResult(message string) analyze.Report {
    return common.NewResultBuilder().BuildCustomEmptyResult(map[string]interface{}{
        "total_functions": 0,
        "metric1":         0.0,
        "message":         message,
    })
}
```

### Step 3: Implement Aggregator
Create the aggregator file (`aggregator.go`):

```go
package myanalyzer

import (
    "github.com/dmytrogajewski/hercules/pkg/analyzers/analyze"
    "github.com/dmytrogajewski/hercules/pkg/analyzers/common"
)

// MyAggregator aggregates analysis results
type MyAggregator struct {
    *common.Aggregator
}

// NewMyAggregator creates a new aggregator
func NewMyAggregator() *MyAggregator {
    numericKeys := []string{"metric1", "metric2"}
    countKeys := []string{"total_functions"}
    messageBuilder := buildMyMessage
    emptyResultBuilder := buildEmptyMyResult

    return &MyAggregator{
        Aggregator: common.NewAggregatorWithCustomEmptyResult(
            "myanalyzer",
            numericKeys,
            countKeys,
            "functions",
            messageBuilder,
            emptyResultBuilder,
        ),
    }
}

// Helper functions
func buildMyMessage(metrics map[string]interface{}) string {
    return "Custom aggregation message"
}

func buildEmptyMyResult() map[string]interface{} {
    return map[string]interface{}{
        "total_functions": 0,
        "metric1":         0.0,
        "message":         "No data to aggregate",
    }
}
```

### Step 4: Add Tests
Create comprehensive tests (`myanalyzer_test.go`):

```go
package myanalyzer

import (
    "testing"
    "github.com/dmytrogajewski/hercules/pkg/uast/pkg/node"
)

func TestMyAnalyzer_Basic(t *testing.T) {
    analyzer := NewMyAnalyzer()
    
    // Test basic functionality
    if analyzer.Name() != "myanalyzer" {
        t.Errorf("Expected name 'myanalyzer', got '%s'", analyzer.Name())
    }
}

func TestMyAnalyzer_Analysis(t *testing.T) {
    analyzer := NewMyAnalyzer()
    
    // Create test UAST
    root := &node.Node{
        Type: node.UASTRoot,
        Children: []*node.Node{
            {
                Type: node.UASTFunction,
                Token: "testFunction",
            },
        },
    }
    
    // Run analysis
    report, err := analyzer.Analyze(root)
    if err != nil {
        t.Fatalf("Analysis failed: %v", err)
    }
    
    // Verify results
    if report == nil {
        t.Fatal("Expected non-nil report")
    }
}
```

### Step 5: Register Analyzer
Add your analyzer to the factory in the main application:

```go
// In your main application
analyzers := []analyze.CodeAnalyzer{
    cohesion.NewCohesionAnalyzer(),
    comments.NewCommentsAnalyzer(),
    complexity.NewComplexityAnalyzer(),
    halstead.NewHalsteadAnalyzer(),
    myanalyzer.NewMyAnalyzer(), // Add your analyzer
}

factory := analyze.NewFactory(analyzers)
```

## Best Practices

### 1. Use Common Modules
- **Leverage existing components** instead of reimplementing
- **Use `common.ResultBuilder`** for consistent result structures
- **Use `common.Aggregator`** for standardized aggregation
- **Use `common.UASTTraverser`** for efficient node finding

### 2. Follow SOLID Principles
- **Single Responsibility:** Each module has one clear purpose
- **Open/Closed:** Extend through composition, not modification
- **Liskov Substitution:** Implementers can be used interchangeably
- **Interface Segregation:** Keep interfaces focused and small
- **Dependency Inversion:** Depend on abstractions, not concretions

### 3. Error Handling
- **Validate inputs** early in the analysis process
- **Return meaningful errors** with context
- **Handle edge cases** gracefully (empty files, no functions, etc.)

### 4. Performance Considerations
- **Use efficient UAST traversal** with appropriate filters
- **Minimize memory allocations** in hot paths
- **Cache expensive computations** when possible

### 5. Testing
- **Write comprehensive unit tests** for all components
- **Test edge cases** and error conditions
- **Use table-driven tests** for multiple scenarios
- **Mock dependencies** for isolated testing

## Usage Examples

### Basic Usage
```go
import "github.com/dmytrogajewski/hercules/pkg/analyzers"

// Create factory with all analyzers
analyzers := []analyze.CodeAnalyzer{
    cohesion.NewCohesionAnalyzer(),
    comments.NewCommentsAnalyzer(),
    complexity.NewComplexityAnalyzer(),
    halstead.NewHalsteadAnalyzer(),
}

factory := analyze.NewFactory(analyzers)

// Run specific analyzer
report, err := factory.RunAnalyzer("halstead", root)
if err != nil {
    log.Fatal(err)
}

// Run multiple analyzers
reports, err := factory.RunAnalyzers(root, []string{"complexity", "halstead"})
if err != nil {
    log.Fatal(err)
}
```

### CLI Integration
```bash
# Analyze with all analyzers
uast parse main.go | herr analyze

# Analyze with specific analyzers
uast parse main.go | herr analyze --analyzers complexity,halstead,comments

# JSON output
uast parse main.go | herr analyze --format json

# Summary output
uast parse main.go | herr analyze --format summary
```

## Example Output

```json
{
  "halstead": {
    "analyzer_name": "halstead",
    "total_functions": 15,
    "volume": 5473.51,
    "difficulty": 153.89,
    "effort": 842312.37,
    "time_to_program": 46795.13,
    "delivered_bugs": 1.82,
    "functions": [
      {
        "name": "calculateMetrics",
        "volume": 312.78,
        "difficulty": 28.25,
        "effort": 8836.05,
        "volume_assessment": "🟡 Medium",
        "difficulty_assessment": "🔴 Complex"
      }
    ],
    "message": "High Halstead complexity detected - consider refactoring complex functions"
  },
  "complexity": {
    "analyzer_name": "complexity",
    "total_functions": 15,
    "average_complexity": 1.41,
    "cognitive_complexity": 74,
    "functions": [
      {
        "name": "processData",
        "cyclomatic_complexity": 4,
        "cognitive_complexity": 6,
        "nesting_depth": 3,
        "complexity_assessment": "🟡 Moderate"
      }
    ]
  }
}
```

## References
- [Cyclomatic Complexity - Wikipedia](https://en.wikipedia.org/wiki/Cyclomatic_complexity)
- [Cognitive Complexity - SonarSource](https://www.sonarsource.com/docs/CognitiveComplexity.pdf)
- [Halstead Complexity Measures - Wikipedia](https://en.wikipedia.org/wiki/Halstead_complexity_measures)
- [Code Cohesion - Wikipedia](https://en.wikipedia.org/wiki/Cohesion_(computer_science))
- [SOLID Principles - Wikipedia](https://en.wikipedia.org/wiki/SOLID) 