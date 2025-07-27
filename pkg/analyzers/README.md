# Analyzers

This package provides code analysis tools for Hercules, focusing on code complexity and related metrics. Each analyzer is self-contained with its own analysis, aggregation, and formatting logic.

## Architecture

The package follows a clean architecture with each analyzer being fully self-contained in its own folder:

```
pkg/analyzers/
├── analyzer.go          # Core interfaces and factory
├── service.go           # High-level service interface with adapters
├── complexity/          # Cyclomatic complexity analyzer
│   ├── complexity.go    # Complexity analyzer (with aggregator and formatter)
│   └── complexity_test.go
├── halstead/            # Halstead complexity measures
│   ├── halstead.go      # Halstead analyzer (with aggregator and formatter)
│   └── halstead_test.go
├── comment_density/     # Comment density analysis
│   ├── comment_density.go    # Comment density analyzer (with aggregator and formatter)
│   └── comment_density_test.go
└── README.md
```

## Core Components

### 1. Analyzers
Each analyzer is in its own folder and implements the `CodeAnalyzer` interface with:
- **Analysis logic**: Core analysis algorithms
- **Aggregation logic**: Result aggregation across files
- **Formatting logic**: Custom output formatting

- **ComplexityAnalyzer** (`complexity/`): Measures cyclomatic complexity
- **HalsteadAnalyzer** (`halstead/`): Measures Halstead complexity metrics
- **CommentDensityAnalyzer** (`comment_density/`): Measures comment density and documentation quality

### 2. Service
High-level service that orchestrates analysis workflows:

- **Service**: Provides `Analyze()` and `AnalyzeAndFormat()` methods
- **Adapters**: Handle interface compatibility between analyzers
- **Delegation**: Each analyzer handles its own aggregation and formatting

## Available Analyzers

### 1. Cyclomatic Complexity Analyzer
- **Folder:** `complexity/`
- **Purpose:** Measures the cyclomatic complexity of functions and methods in the codebase.
- **Formula:**

  > **Cyclomatic Complexity (CC) = E - N + 2P**
  >
  > Where:
  > - E = number of edges in the control flow graph
  > - N = number of nodes in the control flow graph
  > - P = number of connected components (typically 1 for a single function)

  In practice, this analyzer counts:
  - Each function starts with a base complexity of 1.
  - Each decision point (if, for, while, case, catch, etc.) adds 1.
  - Each logical operator (`&&`, `||`, etc.) adds 1.

- **Thresholds:**
  - Green: ≤ 1
  - Yellow: 2–5
  - Red: > 5

### 2. Cognitive Complexity (part of `complexity/`)
- **Purpose:** Measures how difficult code is to understand, considering nesting and flow.
- **Formula:**
  - Inherits cyclomatic complexity points.
  - Adds points for:
    - Function calls
    - Nested expressions
    - Additional control flow constructs

- **Thresholds:**
  - Green: ≤ 1
  - Yellow: 2–7
  - Red: > 7

### 3. Nesting Depth (part of `complexity/`)
- **Purpose:** Measures the maximum nesting level of control structures in a function.
- **Formula:**
  - Increments depth for each nested block (if, loop, switch, try, etc.)
  - Reports the maximum depth encountered

- **Thresholds:**
  - Green: ≤ 1
  - Yellow: 2–3
  - Red: > 3

### 4. Halstead Complexity Measures (`halstead/`)
- **Purpose:** Measures software complexity based on the number of distinct operators and operands.
- **Formulas:**

  > **Program Vocabulary:** η = η₁ + η₂
  > **Program Length:** N = N₁ + N₂
  > **Estimated Length:** N̂ = η₁×log₂(η₁) + η₂×log₂(η₂)
  > **Volume:** V = N × log₂(η)
  > **Difficulty:** D = (η₁/2) × (N₂/η₂)
  > **Effort:** E = D × V
  > **Time to Program:** T = E/18 (seconds)
  > **Delivered Bugs:** B = V/3000

  Where:
  - η₁ = number of distinct operators
  - η₂ = number of distinct operands
  - N₁ = total number of operators
  - N₂ = total number of operands

- **Thresholds:**
  - Volume: Green ≤ 100, Yellow 101–1000, Red > 1000
  - Difficulty: Green ≤ 5, Yellow 6–15, Red > 15
  - Effort: Green ≤ 1000, Yellow 1001–10000, Red > 10000

### 5. Comment Density Analysis (`comment_density/`)
- **Purpose:** Measures the ratio of comments to code and evaluates documentation quality.
- **Metrics:**

  > **Comment Density = Comment Lines / Total Lines**
  > **Documentation Score = (Documented Functions / Total Functions) × 0.7 + min(Density × 4, 1.0) × 0.3**
  > **Comment Quality = Average quality score of all comments**

  The analyzer evaluates:
  - Overall comment density as a percentage of total lines
  - Function documentation coverage (functions with preceding comments)
  - Comment quality based on length and content patterns
  - Different comment types (line, block, hash comments)

- **Quality Assessment:**
  - **Detailed:** Comprehensive comments with multiple sentences
  - **Adequate:** Well-structured comments with sufficient explanation
  - **Short:** Brief but meaningful comments
  - **Minimal:** Very brief comments (single words)
  - **Very Short:** Extremely brief comments (< 5 characters)

- **Thresholds:**
  - Comment Density: Green ≥ 25%, Yellow 15-25%, Red < 5%
  - Documentation Score: Green ≥ 0.8, Yellow 0.6-0.8, Red < 0.3
  - Comment Quality: Green ≥ 0.7, Yellow 0.5-0.7, Red < 0.2

## Usage

### Basic Usage
```go
import "github.com/dmytrogajewski/hercules/pkg/analyzers"

// Create service
service := analyzers.NewService()

// Run analysis and format results
err := service.AnalyzeAndFormat(input, []string{}, "json", writer)
```

### CLI Integration
The analyzers are used by the `herr` command-line tool:

```bash
# Analyze with all analyzers
uast parse main.go | herr analyze

# Analyze with specific analyzers
uast parse main.go | herr analyze --analyzers complexity,halstead,comment_density

# JSON output
uast parse main.go | herr analyze --format json
```

### Adding New Analyzers
1. **Create a new folder** for your analyzer (e.g., `pkg/analyzers/myanalyzer/`)
2. **Implement the `CodeAnalyzer` interface** with:
   - `Name()`: Returns analyzer name
   - `Analyze()`: Performs analysis on UAST
   - `Thresholds()`: Returns metric thresholds
   - `CreateAggregator()`: Returns aggregator instance
   - `FormatReport()`: Formats human-readable output
   - `FormatReportJSON()`: Formats JSON output

3. **Add aggregator logic** within the analyzer:
   - Implement `ResultAggregator` interface
   - Handle result aggregation across files

4. **Add formatting logic** within the analyzer:
   - Custom text formatting with thresholds
   - JSON formatting

5. **Register in service** (if needed):
   - Add to `NewService()` method
   - Create adapter if interface mismatch

## Design Principles

### 1. Self-Contained Analyzers
- **Each analyzer is complete**: Contains analysis, aggregation, and formatting
- **No external dependencies**: All logic is within the analyzer
- **Clear boundaries**: Each analyzer is independent

### 2. Interface-Driven Design
- **`CodeAnalyzer` interface**: Defines contract for all analyzers
- **`ResultAggregator` interface**: Defines aggregation contract
- **Polymorphic behavior**: Service works with any analyzer implementation

### 3. Extensibility
- **Easy to add analyzers**: Just implement the interface
- **No service changes**: New analyzers work automatically
- **Custom formatting**: Each analyzer controls its own output

### 4. Separation of Concerns
- **Analyzers**: Focus on their specific analysis domain
- **Service**: Orchestrates the workflow
- **CLI**: Handles user interaction and I/O

## Example Output

```
{
  "complexity": {
    "function_count": 3,
    "functions": {
      "foo": 2,
      "bar": 1,
      "baz": 4
    },
    "total_complexity": 7
  },
  "comment_density": {
    "total_lines": 150,
    "comment_lines": 30,
    "comment_density": 0.2,
    "documentation_score": 0.75,
    "comment_quality": 0.8,
    "total_functions": 5,
    "documented_functions": 4
  }
}
```

## References
- [Cyclomatic Complexity - Wikipedia](https://en.wikipedia.org/wiki/Cyclomatic_complexity)
- [Cognitive Complexity - SonarSource](https://www.sonarsource.com/docs/CognitiveComplexity.pdf)
- [Halstead Complexity Measures - Wikipedia](https://en.wikipedia.org/wiki/Halstead_complexity_measures) 