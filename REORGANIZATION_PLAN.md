# Hercules Project Reorganization Plan

## Current Structure Analysis

The current Hercules project has a mixed structure that partially follows Go standards but needs reorganization. Here's the analysis:

### Current Structure Issues:
1. **Mixed responsibilities**: `internal/` contains both private and public packages
2. **Inconsistent naming**: Some directories don't follow Go conventions
3. **Missing standard directories**: No clear separation of concerns
4. **Build artifacts in root**: `bin/`, `build.log` should be in build directories
5. **Configuration scattered**: Config files in root instead of `configs/`

## Proposed New Structure (Following Go Standards)

```
hercules/
├── api/                    # OpenAPI/Swagger specs, JSON schema files, protocol definition files
│   └── proto/             # Protocol buffer files
├── assets/                # Other assets to go along with your repository (images, logo, etc)
├── build/                 # Packaging and CI/CD
│   ├── ci/               # CI configuration and scripts
│   ├── package/          # Packaging (i.e. Docker, Helm, RPM, etc)
│   ├── scripts/          # Build scripts
│   └── tools/            # Build tools
├── cmd/                   # Main applications of the project
│   ├── hercules/         # Main hercules binary
│   ├── uast/             # UAST parser binary
│   └── herr/             # Hercules analyzer binary
├── configs/               # Configuration file templates or default configs
│   └── config.yaml.example
├── deployments/           # IaaS, PaaS, system and container orchestration deployment configurations and templates
│   ├── docker/
│   ├── k8s/
│   └── helm/
├── docs/                  # Design and user documents
│   ├── api/
│   ├── design/
│   └── user/
├── examples/              # Examples for your applications and/or public libraries
├── githooks/              # Git hooks
├── internal/              # Private application and library code
│   ├── app/              # Application-specific code
│   ├── pkg/              # Private library code
│   └── server/           # Server-specific code
├── pkg/                   # Library code that's ok to use by external applications
│   ├── analyzers/        # Code analysis tools (separate package)
│   └── uast/             # UAST parsing and manipulation (separate package)
├── test/                  # Additional external test apps and test data
│   ├── data/             # Test data
│   ├── fixtures/         # Test fixtures
│   └── benchmarks/       # Benchmark tests
├── third_party/           # Forked and third-party code
├── tools/                 # Project support tools
├── web/                   # Web application specific components
├── website/               # Project website
├── .gitignore
├── .gitmodules
├── DCO
├── Dockerfile
├── LICENSE.md
├── MAINTAINERS
├── Makefile
├── README.md
├── go.mod
└── go.sum
```

## ✅ COMPLETED MIGRATION STEPS

### Phase 1: Create New Directory Structure ✅
- [x] Created standard directories: `api/`, `assets/`, `build/`, `configs/`, `deployments/`, `examples/`, `githooks/`, `test/`, `third_party/`, `tools/`, `web/`, `website/`
- [x] Created subdirectories: `build/{ci,package,scripts,tools}`, `deployments/{docker,k8s,helm}`, `test/{data,fixtures}`, `internal/{app,pkg,server}`

### Phase 2: Move and Reorganize Files ✅

#### 2.1 Move Configuration Files ✅
- [x] Moved `config.yaml.example` to `configs/`

#### 2.2 Move Deployment Files ✅
- [x] Moved `docker-compose.yml` to `deployments/docker/`
- [x] Moved `k8s/` to `deployments/`
- [x] Moved `helm/` to `deployments/`

#### 2.3 Move Build Artifacts ✅
- [x] Moved `bin/` to `build/`
- [x] Moved `build.log` to `build/`
- [x] Moved `protoc-gen-gogo` to `build/tools/`

#### 2.4 Reorganize Internal Structure ✅
- [x] Moved `internal/core` to `internal/app/`
- [x] Moved `internal/grpc` to `internal/server/`
- [x] Moved `internal/config` to `internal/pkg/`
- [x] Moved `internal/yaml` to `internal/pkg/`
- [x] Moved `internal/mathutil` to `internal/pkg/`
- [x] Moved `internal/levenshtein` to `internal/pkg/`
- [x] Moved `internal/toposort` to `internal/pkg/`
- [x] Moved `internal/rbtree` to `internal/pkg/`
- [x] Moved `internal/plumbing` to `internal/pkg/`
- [x] Moved `internal/extractor` to `internal/pkg/`
- [x] Moved `internal/importmodel` to `internal/pkg/`
- [x] Moved `internal/burndown` to `internal/pkg/`
- [x] Moved `internal/test` to `internal/pkg/`
- [x] Moved `internal/test_data` to `test/data/`
- [x] Moved `internal/pb` to `api/proto/`
- [x] Moved `internal/uastconvert` to `internal/pkg/`
- [x] Moved `internal/dummies.go` to `internal/pkg/`
- [x] Moved `internal/dummies_test.go` to `internal/pkg/`
- [x] Moved `internal/global_test.go` to `internal/pkg/`
- [x] Moved `internal/__init__.py` to `internal/pkg/`
- [x] Moved `leaves` to `internal/pkg/`

#### 2.5 Reorganize Public Packages ✅
- [x] **CORRECTED**: Kept `pkg/analyzers/` as separate package with:
  - `analyzer.go`
  - `complexity.go`
  - `complexity_test.go`
- [x] **CORRECTED**: Kept `pkg/uast/` as separate package with all UAST-related files

#### 2.6 Move Scripts and Tools ✅
- [x] Moved `scripts/` to `build/scripts/`
- [x] Moved `contrib/` to `examples/`

#### 2.7 Move Documentation ✅
- [x] Documentation remains in `docs/` (already properly organized)

#### 2.8 Move Third-party Dependencies ✅
- [x] Moved `go-sitter-forest/` to `third_party/`
- [x] Moved `grammars/` to `third_party/`

#### 2.9 Move Application Files ✅
- [x] Moved `core.go` to `cmd/hercules/`
- [x] Moved `version.go` to `internal/pkg/`
- [x] Moved `benchmarks/` to `test/`

### Phase 3: Update Build System and Scripts ✅

#### 3.1 Update Makefile ✅
- [x] Updated `GOBIN` path from `bin` to `build/bin`
- [x] Updated script paths from `scripts/` to `build/scripts/`
- [x] Updated benchmark output paths to `test/benchmarks/`
- [x] Updated protobuf paths from `internal/pb` to `api/proto`
- [x] Updated all script references in Makefile targets

#### 3.2 Update Build Scripts ✅
- [x] **Precompile Script** (`build/scripts/precompgen/precompile.go`): No changes needed (uses relative paths)
- [x] **UAST Maps Generator** (`build/scripts/uastmapsgen/gen_uastmaps.py`):
  - Updated `grammars_dir` from `'grammars'` to `'third_party/grammars'`
  - Updated `uast` binary path from `'./uast'` to `'./build/bin/uast'`

#### 3.3 Update Benchmark Scripts ✅
- [x] **Benchmark Runner** (`build/scripts/benchmark/benchmark_runner.py`):
  - Updated results directory from `"benchmarks"` to `"test/benchmarks"`
- [x] **Benchmark Plot** (`build/scripts/benchmark/benchmark_plot.py`):
  - Updated output directory from `'benchmark_plots'` to `'test/benchmark_plots'`
- [x] **Benchmark Report** (`build/scripts/benchmark/benchmark_report.py`):
  - Updated results file path from `"benchmarks"` to `"test/benchmarks"`
  - Updated `find_previous_runs()` to use `"test/benchmarks"`
- [x] **Benchmark Comparison** (`build/scripts/benchmark/benchmark_comparison.py`):
  - Updated results file path from `"benchmarks"` to `"test/benchmarks"`
  - Updated `find_previous_runs()` to use `"test/benchmarks"`
- [x] **Benchmark Multi-Comparison** (`build/scripts/benchmark/benchmark_comparison_multi.py`):
  - Updated results file path from `"benchmarks"` to `"test/benchmarks"`

### Phase 4: Update Import Paths ✅
- [x] Updated all import statements in Go files to reflect new structure
- [x] Updated `go.mod` if needed
- [x] Updated documentation references

### Phase 5: Clean Up ✅
- [x] Removed empty directories
- [x] Updated `.gitignore` for new structure
- [x] Test build process
- [x] Updated `README.md` with new structure

## 🎉 REORGANIZATION COMPLETE

All phases of the reorganization have been completed successfully! The project now follows the Standard Go Project Layout and all systems are working correctly.

### ✅ Verification Results
- **Build System**: ✅ All binaries build successfully
- **Benchmark System**: ✅ Comprehensive benchmark suite works with organized results
- **Import Paths**: ✅ All Go imports updated to reflect new structure
- **Documentation**: ✅ README updated with new project structure
- **Cleanup**: ✅ .gitignore updated, no empty directories

### 🚀 Benefits Achieved
1. **Standard Compliance**: Follows Go project layout standards
2. **Better Separation**: Clear distinction between public and private code
3. **Improved Maintainability**: Logical organization makes code easier to find
4. **Better Tooling**: Standard structure works better with Go tools
5. **Clearer Intent**: Directory names clearly indicate purpose
6. **Easier Onboarding**: New developers can understand structure quickly
7. **Proper Package Separation**: `analyzers` and `uast` packages remain separate as intended
8. **Updated Build System**: All scripts and Makefile targets work with new structure 