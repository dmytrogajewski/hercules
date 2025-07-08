# Visualization Modules: Functional Documentation

This document describes how the visualization modules should interact with the Go-based Hercules analysis engine, including data flow, protocols, and specific functionality. The visualization functionality can be reimplemented in any programming language.

## Architecture Overview

The system operated as a **two-stage pipeline**:

1. **Go Stage (Hercules)**: Performs Git repository analysis and generates structured data
2. **Visualization Stage**: Consumes the Go output for visualization, statistical analysis, and reporting

## Data Flow and Communication Protocol

### 1. Protocol Buffers as Data Exchange Format

The Go and visualization components communicated through **Protocol Buffers (protobuf)**:

- **Go side**: Serializes analysis results to protobuf format
- **Visualization side**: Deserializes protobuf data for processing and visualization
- **Shared schema**: `internal/pb/pb.proto` defines the data structures

### 2. Command-Line Pipeline Integration

```bash
# Go generates protobuf data
hercules --burndown --couples --devs --quiet --pb https://github.com/user/repo > analysis.pb

# Visualization component consumes and visualizes the data
{visualisation-cli} -f pb -m all -o output_dir analysis.pb
```

### 3. Real-time Pipeline (stdin/stdout)

```bash
# Direct pipeline: Go output → Visualization input
hercules --burndown --couples --devs --quiet https://github.com/user/repo | {visualisation-cli} -m all -o out
```

## Specific Module Interactions

### 1. Burndown Analysis

**Go Side (`leaves/burndown.go`)**:
- Analyzes code churn over time
- Tracks file-level and developer-level changes
- Generates sparse matrices for efficient storage
- Outputs: `BurndownAnalysisResults` protobuf

**Visualization Side Requirements**:
- Deserialize sparse matrices from protobuf
- Convert to dense matrices for analysis
- Generate time-series visualizations
- Create heatmaps showing code evolution patterns

**Data Structure**:
```protobuf
message BurndownAnalysisResults {
    int32 granularity = 1;
    int32 sampling = 2;
    BurndownSparseMatrix project = 3;
    repeated BurndownSparseMatrix files = 4;
    repeated BurndownSparseMatrix people = 5;
    CompressedSparseRowMatrix people_interaction = 6;
    repeated FilesOwnership files_ownership = 7;
    int64 tick_size = 8;
}
```

### 2. Developer Analysis

**Go Side (`leaves/devs.go`)**:
- Tracks developer activity over time
- Analyzes commits, lines added/removed/changed
- Groups by programming languages
- Outputs: `DevsAnalysisResults` protobuf

**Visualization Side Requirements**:
- Create developer activity timelines
- Generate productivity metrics
- Visualize language usage patterns
- Produce developer comparison charts

### 3. Couples Analysis

**Go Side (`leaves/couples.go`)**:
- Analyzes file and developer coupling
- Uses sparse matrix compression for large datasets
- Outputs: `CouplesAnalysisResults` protobuf

**Visualization Side Requirements**:
- Create network graphs of file relationships
- Visualize developer collaboration patterns
- Generate coupling heatmaps

### 4. Imports Analysis

**Go Side (`leaves/imports_printer.go`)**:
- Extracts import statements from source code
- Tracks dependency usage over time
- Outputs: `ImportsPerDeveloperResults` protobuf

**Visualization Side Requirements**:
- Analyze dependency evolution
- Create dependency graphs
- Track library adoption patterns

### 5. Sentiment Analysis

**Go Side (`leaves/comment_sentiment.go`)**:
- Uses TensorFlow for sentiment analysis
- Processes commit messages and comments
- Outputs: `SentimentAnalysisResults` protobuf

**Visualization Side Requirements**:
- Visualize sentiment trends over time
- Create sentiment heatmaps
- Correlate sentiment with code changes

## Visualization and Reporting Requirements

### 1. Plotting System Requirements

**Core Functionality**:
- Support multiple output formats (PNG, SVG, PDF)
- Configurable plot styles and themes
- Interactive plotting capabilities

**Key Features**:
- Time-series plotting with customizable ticks
- Heatmap generation for matrix data
- Network graph visualization
- Statistical chart generation

**Implementation Options**:
- **JavaScript/TypeScript**: D3.js, Chart.js, Plotly.js
- **Python**: matplotlib, seaborn, plotly
- **R**: ggplot2, plotly, base R graphics
- **C++**: Qt Charts, OpenGL, custom rendering
- **Java**: JFreeChart, XChart, JavaFX
- **Go**: gonum/plot, go-echarts, custom SVG generation

### 2. CLI Interface Requirements

**Command Structure**:
```bash
{visualisation-cli} [OPTIONS] [INPUT_FILE]
```

**Key Options**:
- `-m, --mode`: Analysis mode (burndown, devs, couples, etc.)
- `-f, --format`: Input format (pb, yaml, json)
- `-o, --output`: Output directory
- `--backend`: Visualization backend
- `--disable-projector`: Disable advanced features

**Implementation Options**:
- **Any language**: Use standard CLI libraries (cobra for Go, argparse for Python, etc.)

### 3. Data Readers Requirements

**Supported Formats**:
- Protocol Buffers (primary)
- YAML
- JSON
- Direct stdin streaming

**Functionality**:
- Automatic format detection
- Streaming data processing
- Memory-efficient large file handling
- Error handling and validation

**Implementation Options**:
- **Protocol Buffers**: Available for all major languages
- **YAML/JSON**: Standard libraries in most languages
- **Streaming**: Use language-native I/O libraries

## Statistical Analysis Capabilities

### 1. Time-Series Analysis
- Trend detection in code evolution
- Seasonal pattern identification
- Anomaly detection in developer activity

**Implementation Libraries**:
- **Python**: pandas, numpy, scipy
- **R**: ts, forecast, zoo
- **JavaScript**: d3, statistics.js
- **Go**: gonum/stat, custom algorithms
- **C++**: Eigen, custom implementations

### 2. Clustering and Classification
- Developer behavior clustering
- File type classification
- Project similarity analysis

**Implementation Libraries**:
- **Python**: scikit-learn, hdbscan
- **R**: cluster, mclust
- **JavaScript**: ml-matrix, clustering algorithms
- **Go**: gonum/cluster, custom implementations
- **C++**: mlpack, custom algorithms

### 3. Network Analysis
- Dependency graph analysis
- Developer collaboration networks
- File coupling networks

**Implementation Libraries**:
- **Python**: networkx, igraph
- **R**: igraph, sna
- **JavaScript**: d3-force, graphology
- **Go**: gonum/graph, custom implementations
- **C++**: Boost.Graph, custom algorithms

## Integration Points with Go

### 1. Protobuf Schema Generation
```makefile
# Generate protobuf bindings for your language
protoc --[language]_out [output_dir] --proto_path=internal/pb internal/pb/pb.proto
```

### 2. Docker Integration
Package your visualization component with Go:
```dockerfile
COPY [your-visualization-component] /root/src
RUN [install-your-component]
```

### 3. CI/CD Integration
Complete pipeline example:
```yaml
- hercules --burndown --couples --devs --quiet https://github.com/dmytrogajewski/hercules | [your-visualization-tool] -m all -o out
```

## Data Processing Patterns

### 1. Sparse Matrix Handling
- Go generates sparse matrices for memory efficiency
- Visualization component converts to dense matrices for analysis
- Handles large repositories efficiently

**Implementation Strategy**:
```pseudocode
// Convert sparse matrix to dense
function sparseToDense(sparseMatrix):
    denseMatrix = new Matrix(sparseMatrix.rows, sparseMatrix.columns)
    for each row in sparseMatrix.rows:
        for each column in row.columns:
            denseMatrix[row.index][column.index] = column.value
    return denseMatrix
```

### 2. Time-Series Processing
- Go provides tick-based time series data
- Visualization component interpolates and smooths for visualization
- Supports multiple time granularities

**Implementation Strategy**:
```pseudocode
// Process time series data
function processTimeSeries(ticks, data):
    timeSeries = []
    for i = 0 to ticks.length:
        timeSeries.push({
            timestamp: ticks[i],
            value: data[i]
        })
    return smooth(timeSeries)
```

### 3. Multi-dimensional Analysis
- Go tracks multiple dimensions (files, developers, languages)
- Visualization component correlates and visualizes relationships
- Supports drill-down analysis

## Performance Characteristics

### 1. Memory Efficiency
- Sparse matrix compression in Go
- Streaming data processing in visualization component
- Efficient protobuf serialization

**Implementation Guidelines**:
- Use streaming I/O for large files
- Implement lazy loading for large datasets
- Use memory-mapped files for very large datasets

### 2. Scalability
- Handles repositories with millions of commits
- Supports distributed analysis
- Configurable memory limits

**Implementation Guidelines**:
- Implement pagination for large datasets
- Use worker pools for parallel processing
- Implement caching strategies

### 3. Caching
- Go caches analysis results
- Visualization component caches processed visualizations
- Supports incremental analysis

## Error Handling and Validation

### 1. Data Validation
- Protobuf schema validation
- Range checking for matrix operations
- Type safety between Go and visualization component

**Implementation Guidelines**:
```pseudocode
function validateProtobuf(data):
    try:
        result = protobuf.parse(data)
        validateMatrixRanges(result.matrices)
        return result
    catch ValidationError as e:
        log.error("Invalid protobuf data: " + e.message)
        return null
```

### 2. Error Recovery
- Graceful handling of malformed data
- Fallback visualization modes
- Detailed error reporting

## Configuration and Customization

### 1. Analysis Parameters
- Configurable time windows
- Adjustable granularity settings
- Custom filtering options

### 2. Visualization Options
- Multiple plot styles
- Custom color schemes
- Configurable output formats

### 3. Plugin System
- Extensible analysis modes
- Custom visualization plugins
- Third-party integration support

## Implementation Roadmap

### Phase 1: Core Infrastructure
1. **Protobuf Integration**: Implement protobuf deserialization
2. **Data Structures**: Create language-native data structures
3. **Basic CLI**: Implement command-line interface
4. **File I/O**: Support reading from files and stdin

### Phase 2: Basic Visualizations
1. **Time Series**: Implement basic line charts
2. **Heatmaps**: Create matrix visualization
3. **Bar Charts**: Developer activity visualization
4. **Scatter Plots**: Correlation analysis

### Phase 3: Advanced Features
1. **Network Graphs**: Dependency and coupling visualization
2. **Interactive Plots**: Zoom, pan, hover capabilities
3. **Statistical Analysis**: Clustering, trend detection
4. **Export Formats**: Multiple output formats

### Phase 4: Performance Optimization
1. **Memory Management**: Efficient data handling
2. **Caching**: Result caching and reuse
3. **Parallel Processing**: Multi-threaded analysis
4. **Streaming**: Real-time data processing

## Language-Specific Considerations

### JavaScript/TypeScript
- **Pros**: Rich visualization libraries, web deployment
- **Cons**: Memory limitations for large datasets
- **Libraries**: D3.js, Chart.js, Plotly.js, Node.js for CLI

### Python
- **Pros**: Excellent scientific computing ecosystem
- **Cons**: Performance limitations for real-time processing
- **Libraries**: matplotlib, seaborn, plotly, pandas

### R
- **Pros**: Statistical analysis excellence
- **Cons**: Limited CLI and deployment options
- **Libraries**: ggplot2, plotly, igraph, dplyr

### Go
- **Pros**: Performance, easy deployment
- **Cons**: Limited visualization libraries
- **Libraries**: gonum/plot, go-echarts, custom SVG

### C++
- **Pros**: Maximum performance, memory efficiency
- **Cons**: Complex development, limited libraries
- **Libraries**: Qt Charts, OpenGL, custom rendering

---

This architecture provides a powerful, scalable system for Git repository analysis with clear separation of concerns: Go handles the heavy computational work of Git analysis, while the visualization component provides rich visualization and statistical analysis capabilities. The modular design allows for easy reimplementation in any programming language while maintaining the same data flow and analysis capabilities. 