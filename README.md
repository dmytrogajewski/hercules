# Hercules

> **Disclaimer:**
> This project originates from [src-d/hercules](https://github.com/src-d/hercules), but follows a **completely different concept and architecture**. It is not a drop-in replacement and is maintained independently.

[![CI](https://github.com/dmytrogajewski/hercules/actions/workflows/ci.yml/badge.svg)](https://github.com/dmytrogajewski/hercules/actions/workflows/ci.yml)
[![Docker Image](https://img.shields.io/badge/docker-ready-blue)](https://github.com/dmytrogajewski/hercules/pkgs/container/hercules)
[![Go Reference](https://pkg.go.dev/badge/github.com/dmytrogajewski/hercules.svg)](https://pkg.go.dev/github.com/dmytrogajewski/hercules)

Hercules is a scalable, cloud-native Git repository analysis platform. It provides advanced analytics, REST/gRPC APIs, and supports high-scale deployments with S3-compatible caching, Kubernetes, and modern DevOps workflows.

---

## 🚀 Features

- **S3 & Multi-Backend Cache**: Pluggable cache (S3, MinIO, local, memory) for horizontal scaling and cost efficiency
- **REST & gRPC APIs**: Run analyses via HTTP or gRPC
- **Modern CI/CD**: GitHub Actions, multi-arch Docker, security scanning, release automation
- **Cloud-Native**: Distroless Docker, Kubernetes, Helm, autoscaling, health checks
- **Extensible**: Add new analyses, plug in custom cache backends
- **Production Ready**: IAM/secret support, resource limits, observability, best practices

---

## 🏗️ Quick Start

### Docker (Distroless)
```sh
docker build -t hercules:latest .
docker run -p 8080:8080 hercules:latest
```

### Docker Compose (with MinIO S3 cache)
```sh
docker-compose up -d
# Hercules: http://localhost:8080
# MinIO Console: http://localhost:9001 (minioadmin/minioadmin)
```

### Kubernetes
```sh
kubectl apply -f k8s/deployment.yaml
kubectl get pods -l app=hercules
```

### Helm
```sh
helm repo add hercules https://dmytrogajewski.github.io/hercules
helm install hercules hercules/hercules
```

---

## 🖥️ Command-Line Usage

Hercules can be used as a powerful CLI tool for direct repository analysis, automation, and scripting.

### Basic Analysis

```sh
# Analyze a repository for burndown statistics
hercules --burndown https://github.com/dmytrogajewski/hercules.git

# Multiple analyses in one run
hercules --burndown --couples --devs https://github.com/dmytrogajewski/hercules.git
```

### Custom Options

```sh
# Custom tick size, granularity, and sampling
hercules --burndown --tick-size 12 --granularity 15 --sampling 10 https://github.com/dmytrogajewski/hercules.git
```

### Output Formats

```sh
# Output as JSON
hercules --burndown --json https://github.com/dmytrogajewski/hercules.git > result.json

# Output as YAML
hercules --burndown --yaml https://github.com/dmytrogajewski/hercules.git > result.yaml

# Output as Protobuf
hercules --burndown --pb https://github.com/dmytrogajewski/hercules.git > result.pb
```

### Caching (Local/S3)

```sh
# Use local cache directory
hercules --burndown --cache /tmp/hercules-cache https://github.com/dmytrogajewski/hercules.git

# Use S3 cache (set via config or env)
HERCULES_CACHE_BACKEND=s3 HERCULES_CACHE_S3_BUCKET=my-bucket hercules --burndown https://github.com/dmytrogajewski/hercules.git
```

### Automation Example

```sh
# Analyze all repos in a directory
for repo in ~/code/*/.git; do
  hercules --burndown --devs "$(dirname "$repo")" > "$(basename "$(dirname "$repo")").json"
done
```

### Using Config Files, Env Vars, and Flags

- **Config file:** `hercules --config config.yaml --burndown ...`
- **Environment:** `HERCULES_ANALYSIS_TIMEOUT=60m hercules --burndown ...`
- **Flags:** `hercules --burndown --tick-size 12 ...`

### Supported Analyses & Options

| Flag              | Description                                 | Options (flags/env/config)         |
|-------------------|---------------------------------------------|------------------------------------|
| `--burndown`      | Line burndown statistics                    | `--tick-size`, `--granularity`, `--sampling` |
| `--couples`       | File/developer coupling                     | `--tick-size`                      |
| `--devs`          | Developer activity                          | `--tick-size`                      |
| `--commits-stat`  | Commit statistics                           |                                    |
| `--file-history`  | File history analysis                       |                                    |
| `--imports-per-dev` | Import usage per developer                 |                                    |
| `--shotness`      | Structural hotness                          |                                    |

### CLI Help

```sh
hercules --help
```

---

## ⚡️ Example API Usage

```sh
curl -X POST http://localhost:8080/api/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "repository": "https://github.com/dmytrogajewski/hercules",
    "analyses": ["burndown"],
    "options": {}
  }'
```

---

## 🔒 S3/MinIO/Cloud Cache

- **Any S3-compatible backend**: AWS S3, MinIO, DigitalOcean Spaces, Wasabi, Backblaze, etc.
- **Configurable via YAML, env, or CLI**
- **IAM role or secret support**

Example config:
```yaml
cache:
  enabled: true
  backend: "s3"
  s3_bucket: "hercules-cache"
  s3_region: "us-east-1"
  s3_endpoint: "http://minio:9000" # for MinIO
  s3_prefix: "hercules"
  ttl: "168h"
  aws_access_key_id: "minioadmin"
  aws_secret_access_key: "minioadmin"
```

---

## 🛠️ Architecture

- **Stateless server**: All state in cache (S3/local/memory)
- **Async job queue**: Scalable analysis jobs
- **Pluggable pipeline**: Add new analyses easily
- **Modern Go**: Go 1.24+, AWS SDK v2, Gorilla Mux, Viper
- **Distroless container**: Minimal, secure, non-root

## 📁 Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

```
hercules/
├── api/                    # Protocol definitions and schemas
│   └── proto/             # Protocol buffer files
├── build/                 # Build and CI/CD artifacts
│   ├── bin/              # Compiled binaries
│   ├── scripts/          # Build and maintenance scripts
│   └── tools/            # Build tools
├── cmd/                   # Main applications
│   ├── herr/             # Hercules analyzer binary
│   └── uast/             # UAST parser binary
├── configs/               # Configuration templates
├── deployments/           # Deployment configurations
│   ├── docker/           # Docker configurations
│   ├── k8s/              # Kubernetes manifests
│   └── helm/             # Helm charts
├── docs/                  # Documentation
├── examples/              # Examples and plugins
├── internal/              # Private application code
│   ├── app/              # Application-specific code
│   ├── pkg/              # Private libraries
│   └── server/           # Server implementations
├── pkg/                   # Public libraries
│   ├── analyzers/        # Code analysis tools
│   └── uast/             # UAST parsing and manipulation
├── test/                  # Test data and benchmarks
│   ├── data/             # Test data
│   ├── fixtures/         # Test fixtures
│   └── benchmarks/       # Benchmark results
└── third_party/           # Third-party dependencies
    ├── grammars/         # Language grammars
    └── go-sitter-forest/ # Tree-sitter grammars
```

**Key packages:**
- `pkg/analyzers/`: Public code analysis APIs
- `pkg/uast/`: Universal Abstract Syntax Tree parsing and manipulation
- `internal/app/core/`: Core application logic and pipeline
- `internal/pkg/`: Private utility libraries
- `cmd/herr/`: Main Hercules analyzer binary
- `cmd/uast/`: UAST parser binary

---

## 📦 Deployment

- [Dockerfile](./Dockerfile) (distroless, non-root)
- [docker-compose.yml](./docker-compose.yml) (with MinIO)
- [k8s/deployment.yaml](./k8s/deployment.yaml) (Kubernetes example)
- [helm/hercules/](./helm/hercules/) (Helm chart)
- [docs/DEPLOYMENT.md](./docs/DEPLOYMENT.md) (full guide)

---

## 🧪 CI/CD

- [GitHub Actions](.github/workflows/ci.yml):
  - Lint, test, coverage, security scan
  - Integration test with MinIO
  - Multi-arch Docker build & push
  - Release binaries for all major OS/arch on tag

---

## 📚 Documentation

- [Configuration](docs/CONFIGURATION.md)
- [API Reference](docs/API.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Recipes & Integrations](docs/RECIPES.md)

---

## 🏛️ Project History

This project was originally forked from [src-d/hercules](https://github.com/src-d/hercules), but has since been **completely re-architected** for modern, cloud-native, and high-scale use cases. It is not API or feature compatible with the original src-d/hercules.

---

## 📝 License

[Apache 2.0](./LICENSE.md)

## Embedded UAST Provider

Hercules supports an **Embedded UAST Provider** as a drop-in alternative to Babelfish for UAST-based analyses. This allows you to run structural code analyses offline, in CI, or in restricted environments without a running Babelfish server.

**Key points:**
- The embedded provider uses built-in parsers (currently Go's standard library) to generate UASTs for supported languages.
- Enable it with the CLI flag:
  ```sh
  ./hercules --shotness --uast-provider=embedded <repo>
  ```
- If a file's language is unsupported, Hercules will skip it or warn, but will not fail the analysis.
- The default provider is still Babelfish. You can switch back at any time with `--uast-provider=babelfish`.

**Supported languages:**
- Go (via GoEmbeddedProvider)

**Planned (via Tree-sitter):**
- Java, Kotlin, Swift, JavaScript/TypeScript/React/Angular, Rust, PHP, Python

**Roadmap:**
- See `docs/UAST_PROVIDER_ROADMAP.md` for progress and planned language support.

**Example usage:**
```sh
./hercules --shotness --uast-provider=embedded <repo>
```

If you want to contribute support for more languages, see the roadmap and open a PR!

