# Performance Optimization Roadmap

## Overview

This roadmap provides a detailed implementation plan for the performance optimizations outlined in `OPTIMIZATION_PLAN.md`. Each feature is decomposed into actionable tasks with clear Definition of Ready (DoR), Definition of Done (DoD), and testability criteria.

**Current State Analysis:**
- Single-Pass UAST Visitor: NOT IMPLEMENTED (0%)
- Parallel Analyzers: NOT IMPLEMENTED (0%)
- Parallel Diffing: PARTIALLY IMPLEMENTED (~40% - Imports analyzer has the pattern)
- Graph Int-Mapping: NOT IMPLEMENTED (0%)
- RBTree Sharding: NOT IMPLEMENTED (0%)

---

## Phase 1: Algorithmic Refinements

### Feature 1.1: Single-Pass UAST Visitor

**Priority:** P0 (Highest)  
**Estimated Impact:** 2-5x speedup for static analysis  
**Target:** `pkg/analyzers/common/`

#### Description

Refactor the UAST analysis to use a Visitor/Listener pattern where the tree is traversed once, and analyzers register hooks for specific node types. Currently, each analyzer performs multiple independent traversals, resulting in O(A × P × N) complexity where A=analyzers, P=passes per analyzer, N=nodes.

#### Definition of Ready (DoR)

- [ ] UAST node types catalog exists in `pkg/uast/pkg/node/types.go`
- [ ] All current analyzers (complexity, halstead, cohesion, comments) are identified
- [ ] Node type dependencies for each analyzer are documented
- [ ] Performance baseline measurements collected

#### Tasks

| # | Task | Description | Files |
|---|------|-------------|-------|
| 1.1.1 | Define Visitor Interface | Create `NodeVisitor` interface with `OnEnter(node)`, `OnExit(node)` methods | `pkg/analyzers/analyze/visitor.go` |
| 1.1.2 | Create MultiAnalyzerTraverser | Single traverser that manages multiple visitors, dispatches events | `pkg/analyzers/analyze/multi_traverser.go` |
| 1.1.3 | Implement Hook Registry | Map node types to registered visitors using `map[string][]NodeVisitor` | `pkg/analyzers/analyze/multi_traverser.go` |
| 1.1.4 | Migrate Complexity Analyzer | Implement `NodeVisitor` for complexity, register hooks for Function, If, For, Switch | `pkg/analyzers/complexity/visitor.go` |
| 1.1.5 | Migrate Halstead Analyzer | Implement `NodeVisitor` for halstead, register hooks for operators/operands | `pkg/analyzers/halstead/visitor.go` |
| 1.1.6 | Migrate Cohesion Analyzer | Implement `NodeVisitor` for cohesion metrics | `pkg/analyzers/cohesion/visitor.go` |
| 1.1.7 | Migrate Comments Analyzer | Implement `NodeVisitor` for comment analysis | `pkg/analyzers/comments/visitor.go` |
| 1.1.8 | Update Factory.RunAnalyzers | Use MultiAnalyzerTraverser instead of individual traversals | `pkg/analyzers/analyze/analyzer.go` |

#### Definition of Done (DoD)

- [x] `MultiAnalyzerTraverser` traverses UAST exactly once per analysis run
- [x] All 4 analyzers migrated to visitor pattern
- [x] No analyzer calls `FindNodesByType` or `FindNodesByRoles` directly (Visitor implementations use internal filtering or helpers, effectively replacing direct traversal for analysis)
- [ ] Benchmark shows >= 2x improvement on large UAST (1000+ nodes)
- [x] All existing tests pass
- [x] New unit tests for visitor registration and dispatch

#### Test Strategy

```go
// Test single traversal
func TestMultiAnalyzerTraverser_SinglePass(t *testing.T) {
    counter := &traversalCounter{}
    traverser := NewMultiAnalyzerTraverser()
    traverser.RegisterVisitor(counter)
    traverser.Traverse(largeUAST)
    assert.Equal(t, nodeCount, counter.visits) // Exactly once per node
}

// Benchmark comparison
func BenchmarkOldVsNew(b *testing.B) {
    // Compare old multi-pass vs new single-pass
}
```

---

### Feature 1.2: Integer-Based Graph Processing

**Priority:** P2  
**Estimated Impact:** 2x graph speedup, significant memory reduction  
**Target:** `pkg/toposort/`

#### Description

Replace string-based `map[string]map[string]int` graph representation with integer-based `[][]int` adjacency list. Implement symbol table for string-to-int mapping at graph boundaries.

#### Definition of Ready (DoR)

- [ ] Current graph API surface documented
- [ ] All call sites of `pkg/toposort` identified
- [ ] Memory baseline measured for typical graph sizes

#### Tasks

| # | Task | Description | Files |
|---|------|-------------|-------|
| 1.2.1 | Create Symbol Table | Bidirectional `string <-> int` mapping with `Intern(string) int` and `Resolve(int) string` | `pkg/toposort/symtab.go` |
| 1.2.2 | Implement IntGraph | New struct using `[][]int` adjacency list, integer-only internal operations | `pkg/toposort/intgraph.go` |
| 1.2.3 | Add Wrapper API | Maintain backward-compatible `Graph` API that wraps `IntGraph` + `SymbolTable` | `pkg/toposort/toposort.go` |
| 1.2.4 | Migrate Kahn's Algorithm | Update topological sort to work on integer indices | `pkg/toposort/intgraph.go` |
| 1.2.5 | Add Cycle Detection | Integer-based cycle detection with symbol resolution for error messages | `pkg/toposort/intgraph.go` |
| 1.2.6 | Benchmark & Optimize | Profile and optimize hot paths, consider pre-allocation | `pkg/toposort/toposort_test.go` |

#### Definition of Done (DoD)

- [ ] `IntGraph` passes all existing `Graph` tests
- [ ] Memory usage reduced by >= 40% for graphs with 10k+ nodes
- [ ] Topological sort >= 1.5x faster on large graphs
- [ ] API backward compatible (no breaking changes to callers)
- [ ] Symbol table has O(1) average lookup

#### Test Strategy

```go
func TestSymbolTable_Bidirectional(t *testing.T) {
    st := NewSymbolTable()
    id := st.Intern("node_name")
    assert.Equal(t, "node_name", st.Resolve(id))
}

func BenchmarkGraph_StringVsInt(b *testing.B) {
    // Compare memory and CPU for 10k node graph
}
```

---

## Phase 2: Concurrency & Parallelism

### Feature 2.1: Parallel Static Analyzers

**Priority:** P1  
**Estimated Impact:** Linear speedup with CPU cores  
**Target:** `pkg/analyzers/analyze/`

#### Description

Modify `Factory.RunAnalyzers` to execute registered analyzers in parallel goroutines. UAST nodes are immutable during analysis, making this safe for concurrent access.

#### Definition of Ready (DoR)

- [ ] Feature 1.1 (Single-Pass Visitor) completed OR decision to parallelize without it
- [ ] Confirmed UAST immutability during analysis
- [ ] Error handling strategy defined (fail-fast vs collect-all)

#### Tasks

| # | Task | Description | Files |
|---|------|-------------|-------|
| 2.1.1 | Add Concurrency Config | Add `MaxParallelAnalyzers` to config with sensible default (runtime.NumCPU()) | `pkg/config/` or inline |
| 2.1.2 | Implement Worker Pool | Create goroutine pool for analyzer execution with sync.WaitGroup | `pkg/analyzers/analyze/analyzer.go` |
| 2.1.3 | Thread-Safe Result Collection | Use sync.Map or channel for collecting analyzer reports | `pkg/analyzers/analyze/analyzer.go` |
| 2.1.4 | Error Aggregation | Collect errors from all analyzers, return combined error | `pkg/analyzers/analyze/analyzer.go` |
| 2.1.5 | Context Cancellation | Support context.Context for graceful cancellation | `pkg/analyzers/analyze/analyzer.go` |

#### Definition of Done (DoD)

- [ ] `RunAnalyzers` executes analyzers concurrently
- [ ] Speedup of >= 2x when running 4+ analyzers on multi-core system
- [ ] No race conditions (verified with `-race` flag)
- [ ] Errors from any analyzer properly propagated
- [ ] Context cancellation stops pending analyzers

#### Test Strategy

```go
func TestRunAnalyzers_Parallel(t *testing.T) {
    // Verify all analyzers run and results collected
}

func TestRunAnalyzers_RaceCondition(t *testing.T) {
    // Run with -race, verify no data races
}

func BenchmarkRunAnalyzers_Parallel(b *testing.B) {
    // Compare sequential vs parallel
}
```

---

### Feature 2.2: Parallel Diff Computation

**Priority:** P1  
**Estimated Impact:** Significant for large commits  
**Target:** `pkg/analyzers/burndown/`

#### Description

Parallelize the CPU-heavy `diffmatchpatch` operations in Burndown analysis. Follow the proven pattern from `pkg/analyzers/imports/history.go` which uses worker pools with configurable goroutines.

#### Definition of Ready (DoR)

- [ ] Imports analyzer parallel pattern reviewed
- [ ] Burndown diff hot spots profiled
- [ ] File change batching strategy defined

#### Tasks

| # | Task | Description | Files |
|---|------|-------------|-------|
| 2.2.1 | Add Goroutines Config | Add `Burndown.Goroutines` config option (default: runtime.NumCPU()) | `pkg/analyzers/burndown/` |
| 2.2.2 | Create DiffJob Struct | Define job struct with file info, blob data, result channel | `pkg/analyzers/burndown/history.go` |
| 2.2.3 | Implement Worker Pool | Workers consume from jobs channel, compute diffs in parallel | `pkg/analyzers/burndown/history.go` |
| 2.2.4 | Collect Diff Results | Gather computed diffs, maintain order if needed | `pkg/analyzers/burndown/history.go` |
| 2.2.5 | Sequential Tree Update | Apply collected diffs to RBTree sequentially (preserves correctness) | `pkg/analyzers/burndown/history.go` |
| 2.2.6 | Progress Reporting | Optional callback for progress updates on large commits | `pkg/analyzers/burndown/history.go` |

#### Definition of Done (DoD)

- [ ] Diff computation parallelized across available cores
- [ ] RBTree updates remain sequential (correctness preserved)
- [ ] >= 2x speedup on commits with 100+ file changes
- [ ] Memory usage does not increase significantly
- [ ] Configurable worker count

#### Test Strategy

```go
func TestParallelDiff_Correctness(t *testing.T) {
    // Compare parallel vs sequential results, must match exactly
}

func TestParallelDiff_RaceCondition(t *testing.T) {
    // Run with -race flag
}

func BenchmarkDiff_LargeCommit(b *testing.B) {
    // Commit with 500 file changes
}
```

---

## Phase 3: Architectural Overhaul

### Feature 3.1: Sharded RBTree Allocator

**Priority:** P3  
**Estimated Impact:** Enables full parallel history processing  
**Target:** `pkg/rbtree/`

#### Description

Partition files by hash into K buckets, each with its own `Allocator` and `RBTree`. This allows K concurrent workers to process file histories independently. Cross-shard operations (renames) handled via message passing.

#### Definition of Ready (DoR)

- [ ] Feature 2.2 (Parallel Diff) completed
- [ ] File-to-shard hashing strategy defined
- [ ] Rename handling design documented
- [ ] Memory overhead of sharding estimated

#### Tasks

| # | Task | Description | Files |
|---|------|-------------|-------|
| 3.1.1 | Design Sharding Interface | Define `ShardedAllocator` interface with `GetShard(fileHash)` | `pkg/rbtree/sharded.go` |
| 3.1.2 | Implement ShardedAllocator | K independent `Allocator` instances with consistent hashing | `pkg/rbtree/sharded.go` |
| 3.1.3 | File-to-Shard Mapping | Hash function to deterministically assign files to shards | `pkg/rbtree/sharded.go` |
| 3.1.4 | Per-Shard Worker Pool | Each shard processed by dedicated worker, no lock contention | `pkg/analyzers/burndown/history.go` |
| 3.1.5 | Cross-Shard Rename Handling | Queue rename events, process after shard work completes | `pkg/analyzers/burndown/history.go` |
| 3.1.6 | Merge Shard Results | Combine per-shard burndown data into final result | `pkg/analyzers/burndown/history.go` |
| 3.1.7 | Update Hibernation | Hibernate/Boot all shards, maintain format compatibility | `pkg/rbtree/sharded.go` |

#### Definition of Done (DoD)

- [ ] Files processed in parallel across K shards
- [ ] No lock contention during normal file operations
- [ ] Renames across shards handled correctly
- [ ] Hibernation/Boot works for sharded allocator
- [ ] Linear speedup with shard count (up to core count)
- [ ] Memory overhead < 20% vs single allocator

#### Test Strategy

```go
func TestShardedAllocator_FileDistribution(t *testing.T) {
    // Verify even distribution across shards
}

func TestShardedAllocator_CrossShardRename(t *testing.T) {
    // File renamed from shard A to shard B
}

func BenchmarkBurndown_Sharded(b *testing.B) {
    // Large repo with parallel shard processing
}
```

---

### Feature 3.2: SIMD Optimizations (Optional/Future)

**Priority:** P4 (Future)  
**Estimated Impact:** 2-4x for specific hot paths  
**Target:** `diffmatchpatch`, `levenshtein`

#### Description

Use Go Assembly or CGO to leverage AVX2/NEON instructions for vector comparisons in text diffing algorithms.

#### Definition of Ready (DoR)

- [ ] Hot paths profiled and identified
- [ ] Platform support matrix defined (x86_64, ARM64)
- [ ] CGO build complexity evaluated

#### Tasks

| # | Task | Description |
|---|------|-------------|
| 3.2.1 | Profile Hot Loops | Identify innermost loops in diff algorithms |
| 3.2.2 | Implement AVX2 Version | x86_64 SIMD implementation |
| 3.2.3 | Implement NEON Version | ARM64 SIMD implementation |
| 3.2.4 | Runtime CPU Detection | Select implementation based on CPU features |
| 3.2.5 | Fallback to Pure Go | Ensure pure Go fallback for unsupported platforms |

#### Definition of Done (DoD)

- [ ] SIMD versions >= 2x faster than pure Go
- [ ] Automatic fallback on unsupported CPUs
- [ ] Cross-compile still works
- [ ] No CGO if possible (pure Go assembly preferred)

---

## Implementation Schedule

```
Phase 1 (Algorithmic)
├── Feature 1.1: Single-Pass Visitor [P0] ─────────────┐
└── Feature 1.2: Int-Based Graph [P2] ─────────────────┤
                                                       │
Phase 2 (Concurrency)                                  │ (can start in parallel)
├── Feature 2.1: Parallel Analyzers [P1] ◄─────────────┘
│   └── (depends on 1.1 for maximum benefit)
└── Feature 2.2: Parallel Diffing [P1] ────────────────┐
                                                       │
Phase 3 (Architecture)                                 │
├── Feature 3.1: Sharded Allocator [P3] ◄──────────────┘
│   └── (depends on 2.2 for full benefit)
└── Feature 3.2: SIMD [P4] (optional/future)
```

## Dependencies

```
1.1 Single-Pass Visitor
 └──► 2.1 Parallel Analyzers (optional, but maximizes benefit)

2.2 Parallel Diffing
 └──► 3.1 Sharded Allocator (enables full parallelism)
```

## Metrics & Success Criteria

| Metric | Baseline | Target | How to Measure |
|--------|----------|--------|----------------|
| Static Analysis Time | 1x | 0.2-0.5x | `go test -bench` on 1000+ node UAST |
| Multi-Analyzer Time | 1x | 0.25x (4 cores) | `time` for 4 analyzers |
| Large Commit Diff | 1x | 0.25x (4 cores) | Commit with 100+ files |
| Graph Memory | 1x | 0.6x | `runtime.MemStats` for 10k node graph |
| Full Burndown | 1x | 0.125x (8 shards) | Large repo end-to-end |

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Visitor pattern changes analyzer semantics | Medium | High | Extensive test coverage, golden file comparison |
| Race conditions in parallel code | Medium | High | `-race` flag in CI, careful synchronization |
| Sharding breaks rename detection | Low | Medium | Dedicated cross-shard rename queue |
| SIMD complexity not worth it | Medium | Low | Keep as optional Phase 4 |

---

## Appendix: Existing Parallel Pattern Reference

The Imports analyzer (`pkg/analyzers/imports/history.go:166-195`) provides a proven pattern:

```go
wg := sync.WaitGroup{}
wg.Add(h.Goroutines)

jobs := make(chan *object.Change, len(changes))

for i := 0; i < h.Goroutines; i++ {
    go func() {
        defer wg.Done()
        for change := range jobs {
            // Process change in parallel
        }
    }()
}

for _, change := range changes {
    jobs <- change
}
close(jobs)
wg.Wait()
```

This pattern should be replicated for Burndown diffing and other parallelizable operations.
