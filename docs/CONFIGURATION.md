# Configuration Guide

Hercules can be configured using YAML configuration files, environment variables, or command-line flags. This guide covers all configuration options.

## Configuration File

Create a `config.yaml` file in your working directory or specify a custom path with the `--config` flag.

### Basic Configuration

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"

grpc:
  enabled: false
  port: 9090
  host: "0.0.0.0"

cache:
  enabled: true
  backend: "local"
  directory: "/tmp/hercules-cache"
  ttl: "24h"

analysis:
  default_tick_size: 24
  default_granularity: 30
  default_sampling: 30
  max_concurrent_analyses: 10
  timeout: "30m"

logging:
  level: "info"
  format: "json"
  output: "stdout"

repository:
  clone_timeout: "10m"
  max_file_size: "1MB"
  allowed_protocols: ["https", "http", "ssh", "git"]
```

## Cache Configuration

Hercules supports multiple cache backends for storing analysis results and improving performance.

### Local Filesystem Cache

The default cache backend uses the local filesystem:

```yaml
cache:
  enabled: true
  backend: "local"
  directory: "/tmp/hercules-cache"
  ttl: "24h"
```

### S3 Cache Backend

For production environments with high scale, S3 caching provides shared storage across multiple instances:

```yaml
cache:
  enabled: true
  backend: "s3"
  s3_bucket: "hercules-cache"
  s3_region: "us-east-1"
  s3_prefix: "hercules"
  ttl: "168h"  # 1 week
```

#### S3 Configuration Options

| Option | Description | Required |
|--------|-------------|----------|
| `s3_bucket` | S3 bucket name | Yes |
| `s3_region` | AWS region | Yes |
| `s3_endpoint` | Custom endpoint (for MinIO, etc.) | No |
| `s3_prefix` | Key prefix for organization | No |
| `aws_access_key_id` | AWS access key | No* |
| `aws_secret_access_key` | AWS secret key | No* |

*AWS credentials are optional if using IAM roles or AWS CLI configuration.

#### MinIO Configuration

For self-hosted S3-compatible storage:

```yaml
cache:
  enabled: true
  backend: "s3"
  s3_bucket: "hercules-cache"
  s3_region: "us-east-1"
  s3_endpoint: "http://minio:9000"
  s3_prefix: "hercules"
  ttl: "24h"
  aws_access_key_id: "minioadmin"
  aws_secret_access_key: "minioadmin"
```

### Memory Cache

For development or testing with fast, temporary storage:

```yaml
cache:
  enabled: true
  backend: "memory"
  ttl: "1h"
```

## Environment Variables

All configuration options can be set using environment variables with the `HERCULES_` prefix:

```bash
export HERCULES_SERVER_PORT=8080
export HERCULES_CACHE_ENABLED=true
export HERCULES_CACHE_BACKEND=s3
export HERCULES_CACHE_S3_BUCKET=hercules-cache
export HERCULES_CACHE_S3_REGION=us-east-1
```

## Command-Line Flags

Server-specific options can be set via command-line flags:

```bash
hercules server --port 8080 --config /path/to/config.yaml
```

## Configuration Validation

Hercules validates configuration on startup and will fail with descriptive error messages for invalid settings.

### Common Issues

1. **Invalid port numbers**: Must be between 1-65535
2. **Missing S3 bucket**: Required when using S3 backend
3. **Invalid TTL format**: Must be parseable duration (e.g., "24h", "1h30m")
4. **Permission errors**: Ensure cache directory is writable

## Production Recommendations

### For High-Scale Deployments (50k+ repos)

1. **Use S3 caching**: Enables horizontal scaling
2. **Set appropriate TTL**: Balance freshness vs performance
3. **Configure IAM roles**: Avoid hardcoded credentials
4. **Use regional endpoints**: Minimize latency

```yaml
cache:
  enabled: true
  backend: "s3"
  s3_bucket: "hercules-cache-prod"
  s3_region: "us-west-2"
  s3_prefix: "hercules/v1"
  ttl: "168h"  # 1 week
```

### For Development

```yaml
cache:
  enabled: true
  backend: "local"
  directory: "/tmp/hercules-cache"
  ttl: "1h"
```

### For Testing

```yaml
cache:
  enabled: true
  backend: "memory"
  ttl: "10m"
``` 