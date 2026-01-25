# Installation & Deployment

## Installation

### From Source

```bash
git clone https://github.com/dmytrogajewski/hercules.git
cd hercules
make build

# Binaries are created in current directory
./herr --help
./uast --help
```

### Go Install

```bash
go install github.com/dmytrogajewski/hercules/cmd/herr@latest
go install github.com/dmytrogajewski/hercules/cmd/uast@latest
```

### Docker

```bash
# Build
docker build -t hercules .

# Run static analysis
docker run -v $(pwd):/code hercules uast parse /code/main.go | docker run -i hercules herr analyze

# Run history analysis
docker run -v $(pwd):/repo hercules herr history -a burndown /repo
```

## Binaries

The project produces two CLI tools:

| Binary | Purpose |
|--------|---------|
| `herr` | Code analysis (static and history) |
| `uast` | UAST parsing and utilities |

## Usage Examples

### Static Analysis

```bash
# Parse and analyze
uast parse main.go | herr analyze

# Specific analyzers
uast parse main.go | herr analyze -a complexity,halstead

# JSON output
uast parse main.go | herr analyze -f json
```

### History Analysis

```bash
# Analyze git repository
herr history -a burndown .
herr history -a burndown,couples,devs .
herr history -a devs --head .
```

### UAST Operations

```bash
# Parse source file
uast parse main.go

# Validate language mapping
uast validate go

# Start development server
uast server --port 8080
```

## Build Options

```bash
# Build all
make build

# Build with race detector
make build-race

# Run tests
make test

# Run linter
make lint
```
