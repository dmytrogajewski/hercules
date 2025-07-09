# Deployment Guide

This guide covers deploying Hercules using Docker, Docker Compose, Kubernetes, and Helm.

## Quick Start with Docker

### Build and Run

```bash
# Build the image
docker build -t hercules .

# Run with local cache
docker run -p 8080:8080 -p 9090:9090 hercules server

# Run with custom config
docker run -p 8080:8080 -v $(pwd)/config.yaml:/etc/hercules/config.yaml hercules server --config /etc/hercules/config.yaml
```

### Using Docker Compose

```bash
# Start with MinIO for S3 cache testing
docker-compose up -d

# Access services
# Hercules: http://localhost:8080
# MinIO Console: http://localhost:9001 (minioadmin/minioadmin)
```

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured
- Helm 3.x (optional)

### Direct kubectl Deployment

```bash
# Apply the deployment
kubectl apply -f k8s/deployment.yaml

# Check status
kubectl get pods -l app=hercules
kubectl logs -l app=hercules

# Access the service
kubectl port-forward svc/hercules-service 8080:80
```

### Using Helm

```bash
# Add the chart
helm repo add hercules https://dmytrogajewski.github.io/hercules
helm repo update

# Install with default values
helm install hercules hercules/hercules

# Install with custom values
helm install hercules hercules/hercules \
  --set replicaCount=5 \
  --set config.cache.s3_bucket=my-hercules-cache \
  --set config.cache.s3_region=us-west-2

# Upgrade existing installation
helm upgrade hercules hercules/hercules \
  --set image.tag=v1.2.0
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HERCULES_SERVER_PORT` | HTTP server port | 8080 |
| `HERCULES_CACHE_ENABLED` | Enable caching | true |
| `HERCULES_CACHE_BACKEND` | Cache backend (local/s3/memory) | local |
| `HERCULES_CACHE_S3_BUCKET` | S3 bucket name | - |
| `HERCULES_CACHE_S3_REGION` | AWS region | us-east-1 |
| `HERCULES_CACHE_S3_ENDPOINT` | Custom S3 endpoint | - |
| `HERCULES_CACHE_S3_PREFIX` | S3 key prefix | hercules |
| `HERCULES_CACHE_TTL` | Cache TTL | 24h |
| `HERCULES_ANALYSIS_MAX_CONCURRENT_ANALYSES` | Max concurrent jobs | 10 |
| `HERCULES_ANALYSIS_TIMEOUT` | Analysis timeout | 30m |

### S3 Configuration Examples

#### AWS S3 with IAM Roles
```yaml
cache:
  enabled: true
  backend: "s3"
  s3_bucket: "hercules-cache"
  s3_region: "us-east-1"
  s3_prefix: "hercules"
  ttl: "168h"
```

#### MinIO (Self-hosted)
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

#### DigitalOcean Spaces
```yaml
cache:
  enabled: true
  backend: "s3"
  s3_bucket: "hercules-cache"
  s3_region: "nyc3"
  s3_endpoint: "https://nyc3.digitaloceanspaces.com"
  s3_prefix: "hercules"
  ttl: "168h"
```

## Production Deployment

### High-Scale Setup (50k+ repos)

```yaml
# values-production.yaml
replicaCount: 10

resources:
  limits:
    cpu: 2000m
    memory: 4Gi
  requests:
    cpu: 500m
    memory: 1Gi

autoscaling:
  enabled: true
  minReplicas: 5
  maxReplicas: 20
  targetCPUUtilizationPercentage: 60
  targetMemoryUtilizationPercentage: 70

config:
  cache:
    enabled: true
    backend: "s3"
    s3_bucket: "hercules-cache-prod"
    s3_region: "us-west-2"
    s3_prefix: "hercules/v1"
    ttl: "168h"
  
  analysis:
    max_concurrent_analyses: 3
    timeout: "60m"
  
  server:
    read_timeout: "60s"
    write_timeout: "60s"
    idle_timeout: "120s"
```

### Security Considerations

#### Using IAM Roles (Recommended)
```yaml
# No credentials needed - uses IAM role
config:
  cache:
    enabled: true
    backend: "s3"
    s3_bucket: "hercules-cache"
    s3_region: "us-east-1"
```

#### Using Secrets
```bash
# Create secret
kubectl create secret generic hercules-secrets \
  --from-literal=aws-access-key-id=AKIA... \
  --from-literal=aws-secret-access-key=...

# Reference in deployment
env:
  - name: HERCULES_CACHE_AWS_ACCESS_KEY_ID
    valueFrom:
      secretKeyRef:
        name: hercules-secrets
        key: aws-access-key-id
```

### Monitoring and Health Checks

The deployment includes:
- **Liveness Probe**: `/health` endpoint
- **Readiness Probe**: `/health` endpoint
- **Resource Limits**: CPU and memory constraints
- **Horizontal Pod Autoscaler**: Automatic scaling

### Network Configuration

#### Ingress Example (NGINX)
```yaml
ingress:
  enabled: true
  className: "nginx"
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  hosts:
    - host: hercules.your-domain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - hosts:
        - hercules.your-domain.com
      secretName: hercules-tls
```

#### Load Balancer
```yaml
service:
  type: LoadBalancer
  port: 80
  grpcPort: 9090
```

## Troubleshooting

### Common Issues

#### 1. S3 Connection Errors
```bash
# Check S3 credentials and permissions
kubectl logs -l app=hercules | grep -i s3

# Test S3 access
kubectl exec -it deployment/hercules -- aws s3 ls s3://hercules-cache
```

#### 2. High Memory Usage
```yaml
# Increase memory limits
resources:
  limits:
    memory: 4Gi
  requests:
    memory: 1Gi
```

#### 3. Analysis Timeouts
```yaml
# Increase timeout and reduce concurrency
config:
  analysis:
    timeout: "60m"
    max_concurrent_analyses: 2
```

#### 4. Cache Performance
```yaml
# Use local cache for development
config:
  cache:
    backend: "local"
    directory: "/tmp/hercules-cache"
    ttl: "1h"
```

### Debug Commands

```bash
# Check pod status
kubectl get pods -l app=hercules

# View logs
kubectl logs -l app=hercules -f

# Exec into container
kubectl exec -it deployment/hercules -- /bin/sh

# Check config
kubectl exec -it deployment/hercules -- cat /etc/hercules/config.yaml

# Test health endpoint
kubectl exec -it deployment/hercules -- wget -qO- http://localhost:8080/health
```

## Performance Tuning

### For Large Repositories

```yaml
config:
  analysis:
    timeout: "120m"
    max_concurrent_analyses: 1
  
  repository:
    clone_timeout: "30m"
    max_file_size: "10MB"
  
  cache:
    ttl: "720h"  # 30 days
```

### For High Concurrency

```yaml
config:
  analysis:
    max_concurrent_analyses: 10
    timeout: "30m"
  
  server:
    read_timeout: "30s"
    write_timeout: "30s"
```

## Backup and Recovery

### S3 Cache Backup
```bash
# Backup cache data
aws s3 sync s3://hercules-cache s3://hercules-cache-backup

# Restore cache data
aws s3 sync s3://hercules-cache-backup s3://hercules-cache
```

### Configuration Backup
```bash
# Export config
kubectl get configmap hercules-config -o yaml > config-backup.yaml

# Restore config
kubectl apply -f config-backup.yaml
```

## Cost Optimization

### S3 Storage Classes
```yaml
# Use Intelligent Tiering for cost savings
# Configure in AWS S3 bucket lifecycle rules
```

### Resource Optimization
```yaml
# Right-size resources based on usage
resources:
  requests:
    cpu: 250m
    memory: 512Mi
  limits:
    cpu: 1000m
    memory: 2Gi
```

This deployment setup provides a scalable, production-ready Hercules installation with S3 caching support! 🚀 