# FRD-1.2: Integer-Based Graph Processing

## 1. Overview
The current graph implementation in `pkg/toposort` uses `map[string]map[string]int` which is memory inefficient and slow due to string hashing and map overhead. This FRD proposes replacing it with an integer-based `[][]int` adjacency list representation, using a symbol table for string mapping.

## 2. Goals
- Reduce memory usage by at least 40% for large graphs (10k+ nodes).
- Improve topological sort performance by at least 1.5x.
- Maintain backward compatibility with the existing `Graph` API.

## 3. Architecture

### 3.1. Symbol Table (`pkg/toposort/symtab.go`)
Bidirectional mapping between strings and integer IDs.

```go
type SymbolTable struct {
    strToID map[string]int
    idToStr []string
}

func (s *SymbolTable) Intern(name string) int
func (s *SymbolTable) Resolve(id int) string
```

### 3.2. Integer Graph (`pkg/toposort/intgraph.go`)
Adjacency list representation.

```go
type IntGraph struct {
    nodes [][]int // Adjacency list: nodes[u] = [v1, v2, ...]
    inDegree []int
}

func (g *IntGraph) AddEdge(u, v int)
func (g *IntGraph) TopoSort() ([]int, bool) // returns sorted IDs and success flag
```

### 3.3. Graph Wrapper (`pkg/toposort/toposort.go`)
The existing `Graph` struct will be refactored to wrap `IntGraph` and `SymbolTable`, maintaining the current string-based API.

```go
type Graph struct {
    symbols *SymbolTable
    intGraph *IntGraph
}

func (g *Graph) AddNode(name string)
func (g *Graph) AddEdge(from, to string)
func (g *Graph) TopSort() ([]string, bool)
```

## 4. Implementation Plan

1.  **Create Symbol Table**: Implement `pkg/toposort/symtab.go` with tests.
2.  **Implement IntGraph**: Implement `pkg/toposort/intgraph.go` with adjacency list and Kahn's algorithm for topological sort.
3.  **Refactor Graph**: Update `pkg/toposort/toposort.go` to use `IntGraph` and `SymbolTable` internally.
4.  **Verification**: Run existing tests to ensure backward compatibility.
5.  **Benchmarks**: Compare memory and CPU performance.

## 5. Testing Strategy
- **Unit Tests**: Test `SymbolTable` and `IntGraph` in isolation.
- **Integration Tests**: Ensure `Graph` passes all existing tests.
- **Benchmarks**: Measure performance gain.

## 6. Acceptance Criteria
- All existing tests in `pkg/toposort` pass.
- Memory usage reduced.
- Topological sort performance improved.
