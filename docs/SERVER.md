# Hercules Server

Hercules can now run as an HTTP server to provide Git repository analysis capabilities via REST API.

## Quick Start

Start the server:

```bash
# Start with default configuration
./hercules server

# Start with custom configuration file
./hercules server -c config.yaml

# Start with environment variables
HERCULES_SERVER_PORT=9000 ./hercules server
```

## Configuration

The server uses [Viper](https://github.com/spf13/viper) for configuration management, supporting:

- **Configuration files** (YAML, JSON, TOML, HCL, INI, Java properties)
- **Environment variables** (prefixed with `HERCULES_`)
- **Command line flags**
- **Default values**

### Configuration File

Create a `config.yaml` file:

```yaml
# Server settings
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s

# Cache settings
cache:
  directory: "/tmp/hercules-cache"
  cleanup_interval: "1h"
  max_size: "10GB"

# Analysis settings
analysis:
  default_tick_size: 24
  default_granularity: 30
  default_sampling: 30
  max_concurrent_analyses: 10
  timeout: "30m"

# Logging settings
logging:
  level: "info"
  format: "json"
  output: "stdout"

# Repository settings
repository:
  clone_timeout: "10m"
  max_file_size: "1MB"
  allowed_protocols: ["https", "http", "ssh", "git"]
```

### Environment Variables

Override configuration with environment variables:

```bash
export HERCULES_SERVER_PORT=9000
export HERCULES_CACHE_DIRECTORY="/var/cache/hercules"
export HERCULES_ANALYSIS_MAX_CONCURRENT_ANALYSES=5
./hercules server
```

## API Endpoints

### Health Check
```
GET /health
```

Returns server health status:
```json
{
  "status": "healthy",
  "timestamp": "2025-07-09T01:48:09.49047031+03:00",
  "version": 10,
  "hash": "68bb211faaedeffb53e799ab89e2aa48d8cb0ad3",
  "config": {
    "server_port": 8080,
    "cache_dir": "/tmp/hercules-cache"
  }
}
```

### List Available Analyses
```
GET /api/v1/analyses
```

Returns available analysis types:
```json
{
  "analyses": [
    {
      "name": "burndown",
      "description": "Line burndown statistics for project, files and developers"
    },
    {
      "name": "couples",
      "description": "Coupling statistics for files and developers"
    },
    {
      "name": "devs",
      "description": "Developer activity statistics"
    },
    {
      "name": "commits-stat",
      "description": "Statistics for each commit"
    },
    {
      "name": "file-history",
      "description": "File history analysis"
    },
    {
      "name": "imports-per-dev",
      "description": "Import usage per developer"
    },
    {
      "name": "shotness",
      "description": "Structural hotness analysis"
    }
  ],
  "timestamp": "2025-07-09T01:48:14.59786901+03:00"
}
```

### Submit Analysis Request
```
POST /api/v1/analyze
```

Request body:
```json
{
  "repository": "https://github.com/dmytrogajewski/hercules.git",
  "analyses": ["burndown", "couples"],
  "options": {
    "tick-size": "24",
    "granularity": "30",
    "sampling": "30"
  }
}
```

Response:
```json
{
  "status": "accepted",
  "message": "Analysis started",
  "timestamp": "2025-07-09T01:48:19.943861855+03:00"
}
```

### Check Analysis Status
```
GET /api/v1/status/{job_id}
```

Response:
```json
{
  "status": "completed",
  "message": "Analysis completed",
  "timestamp": "2025-07-09T01:48:25.123456789+03:00"
}
```

## Analysis Options

### Burndown Analysis
- `tick-size`: Number of hours per tick (default: 24)
- `granularity`: Granularity in days (default: 30)
- `sampling`: Sampling rate (default: 30)

### Couples Analysis
- `tick-size`: Number of hours per tick (default: 24)

### Devs Analysis
- `tick-size`: Number of hours per tick (default: 24)

## Configuration Options

### Server Configuration
- `server.port`: HTTP server port (default: 8080)
- `server.host`: HTTP server host (default: "0.0.0.0")
- `server.read_timeout`: Request read timeout (default: 30s)
- `server.write_timeout`: Response write timeout (default: 30s)
- `server.idle_timeout`: Connection idle timeout (default: 60s)

### Cache Configuration
- `cache.directory`: Cache directory for repositories (default: "/tmp/hercules-cache")
- `cache.cleanup_interval`: Cache cleanup interval (default: 1h)
- `cache.max_size`: Maximum cache size (default: "10GB")

### Analysis Configuration
- `analysis.default_tick_size`: Default tick size for analyses (default: 24)
- `analysis.default_granularity`: Default granularity for burndown (default: 30)
- `analysis.default_sampling`: Default sampling rate (default: 30)
- `analysis.max_concurrent_analyses`: Maximum concurrent analyses (default: 10)
- `analysis.timeout`: Analysis timeout (default: 30m)

### Logging Configuration
- `logging.level`: Log level (default: "info")
- `logging.format`: Log format (default: "json")
- `logging.output`: Log output (default: "stdout")

### Repository Configuration
- `repository.clone_timeout`: Repository clone timeout (default: 10m)
- `repository.max_file_size`: Maximum file size to process (default: "1MB")
- `repository.allowed_protocols`: Allowed repository protocols (default: ["https", "http", "ssh", "git"])

## Examples

### Basic Usage
```bash
# Start server with default configuration
./hercules server

# Test health endpoint
curl http://localhost:8080/health

# List available analyses
curl http://localhost:8080/api/v1/analyses

# Submit analysis request
curl -X POST -H "Content-Type: application/json" \
  -d '{"repository":"https://github.com/dmytrogajewski/hercules.git","analyses":["burndown"]}' \
  http://localhost:8080/api/v1/analyze
```

### Custom Configuration
```bash
# Create custom config
cat > my-config.yaml << EOF
server:
  port: 9000
  host: "127.0.0.1"

cache:
  directory: "/var/cache/hercules"

analysis:
  max_concurrent_analyses: 5
  timeout: "1h"
EOF

# Start with custom config
./hercules server -c my-config.yaml
```

### Environment Variables
```bash
# Override specific settings
export HERCULES_SERVER_PORT=9000
export HERCULES_CACHE_DIRECTORY="/var/cache/hercules"
export HERCULES_ANALYSIS_MAX_CONCURRENT_ANALYSES=5

# Start server
./hercules server
```

## Architecture

The server uses a modular architecture with:

- **Configuration Management**: Viper for flexible configuration
- **HTTP Server**: Standard library with Gorilla Mux for routing
- **Job Management**: In-memory job tracking with mutex protection
- **Repository Caching**: Temporary directories for efficient cloning
- **Async Processing**: Background goroutines for analysis execution

## Future Enhancements

- **Job Tracking**: Implement proper job status tracking with unique IDs
- **Result Storage**: Store analysis results for later retrieval
- **Authentication**: Add authentication and authorization
- **Rate Limiting**: Implement request rate limiting
- **Result Streaming**: Stream analysis results as they become available
- **WebSocket Support**: Real-time analysis progress updates 