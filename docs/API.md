# Hercules API Documentation

## Overview

The Hercules HTTP API provides programmatic access to Git repository analysis capabilities. The API is RESTful and returns JSON responses.

**Base URL**: `http://localhost:8080` (default)

**Content-Type**: `application/json`

## Authentication

Currently, the API does not require authentication. Future versions may include API key or OAuth2 authentication.

## Rate Limiting

The API implements basic rate limiting:
- Maximum 10 concurrent analysis jobs (configurable)
- Request throttling based on server load

## Error Handling

All errors follow a consistent format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "timestamp": "2025-07-09T01:48:09.49047031+03:00"
}
```

### HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `202 Accepted` - Request accepted for processing
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Endpoints

### Health Check

Check server health and status.

```http
GET /health
```

**Response:**
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

**Response Fields:**
- `status` (string): Server status ("healthy", "unhealthy")
- `timestamp` (string): ISO 8601 timestamp
- `version` (integer): Hercules version number
- `hash` (string): Git commit hash of the build
- `config` (object): Current server configuration

### List Available Analyses

Get a list of all available analysis types.

```http
GET /api/v1/analyses
```

**Response:**
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

**Response Fields:**
- `analyses` (array): List of available analysis types
  - `name` (string): Analysis identifier
  - `description` (string): Human-readable description
- `timestamp` (string): ISO 8601 timestamp

### Submit Analysis Request

Submit a new analysis job.

```http
POST /api/v1/analyze
Content-Type: application/json
```

**Request Body:**
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

**Request Fields:**
- `repository` (string, required): Git repository URL or local path
- `analyses` (array, required): List of analysis types to run
- `options` (object, optional): Analysis-specific options
  - `tick-size` (string): Time granularity in hours
  - `granularity` (string): Burndown granularity in days
  - `sampling` (string): Burndown sampling frequency

**Response:**
```json
{
  "status": "accepted",
  "message": "Analysis started",
  "job_id": "job_1234567890",
  "timestamp": "2025-07-09T01:48:19.943861855+03:00"
}
```

**Response Fields:**
- `status` (string): Job status ("accepted", "failed")
- `message` (string): Human-readable message
- `job_id` (string): Unique job identifier
- `timestamp` (string): ISO 8601 timestamp

**Error Responses:**

```json
{
  "error": "Repository URL is required",
  "code": "MISSING_REPOSITORY",
  "timestamp": "2025-07-09T01:48:19.943861855+03:00"
}
```

```json
{
  "error": "Too many concurrent analyses",
  "code": "RATE_LIMIT_EXCEEDED",
  "timestamp": "2025-07-09T01:48:19.943861855+03:00"
}
```

### Check Analysis Status

Get the status of a running or completed analysis job.

```http
GET /api/v1/status/{job_id}
```

**Path Parameters:**
- `job_id` (string, required): Job identifier returned from submit request

**Response:**
```json
{
  "status": "completed",
  "message": "Analysis completed",
  "job_id": "job_1234567890",
  "start_time": "2025-07-09T01:48:19.943861855+03:00",
  "end_time": "2025-07-09T01:49:25.123456789+03:00",
  "results": {
    "burndown": {
      "name": "BurndownAnalysis",
      "description": "Line burndown statistics",
      "result_type": "*leaves.BurndownResult"
    }
  },
  "timestamp": "2025-07-09T01:49:25.123456789+03:00"
}
```

**Response Fields:**
- `status` (string): Job status ("running", "completed", "failed")
- `message` (string): Human-readable message
- `job_id` (string): Job identifier
- `start_time` (string): ISO 8601 timestamp when job started
- `end_time` (string): ISO 8601 timestamp when job completed (if completed)
- `results` (object): Analysis results (if completed)
- `timestamp` (string): ISO 8601 timestamp of the response

**Error Response:**
```json
{
  "error": "Job not found",
  "code": "JOB_NOT_FOUND",
  "timestamp": "2025-07-09T01:48:19.943861855+03:00"
}
```

## Analysis Types

### Burndown Analysis

Analyzes line burndown statistics for the entire repository.

**Options:**
- `tick-size` (string): Time granularity in hours (default: "24")
- `granularity` (string): Burndown granularity in days (default: "30")
- `sampling` (string): Burndown sampling frequency (default: "30")

### Couples Analysis

Analyzes coupling statistics between files and developers.

**Options:**
- `tick-size` (string): Time granularity in hours (default: "24")

### Developer Analysis

Analyzes developer activity statistics.

**Options:**
- `tick-size` (string): Time granularity in hours (default: "24")

### Commits Statistics

Analyzes statistics for each commit.

**Options:** None

### File History Analysis

Analyzes file history and evolution.

**Options:** None

### Imports Per Developer

Analyzes import usage patterns per developer.

**Options:**
- `tick-size` (string): Time granularity in hours (default: "24")

### Structural Hotness

Analyzes structural hotness of code elements.

**Options:** None

## Examples

### Complete Analysis Workflow

```bash
# 1. Submit analysis request
curl -X POST -H "Content-Type: application/json" \
  -d '{
    "repository": "https://github.com/dmytrogajewski/hercules.git",
    "analyses": ["burndown", "couples"],
    "options": {
      "tick-size": "24",
      "granularity": "30"
    }
  }' \
  http://localhost:8080/api/v1/analyze

# Response:
{
  "status": "accepted",
  "message": "Analysis started",
  "job_id": "job_1234567890",
  "timestamp": "2025-07-09T01:48:19.943861855+03:00"
}

# 2. Check status
curl http://localhost:8080/api/v1/status/job_1234567890

# Response:
{
  "status": "completed",
  "message": "Analysis completed",
  "job_id": "job_1234567890",
  "start_time": "2025-07-09T01:48:19.943861855+03:00",
  "end_time": "2025-07-09T01:49:25.123456789+03:00",
  "results": {
    "burndown": {
      "name": "BurndownAnalysis",
      "description": "Line burndown statistics",
      "result_type": "*leaves.BurndownResult"
    },
    "couples": {
      "name": "CouplesAnalysis",
      "description": "Coupling statistics",
      "result_type": "*leaves.CouplesResult"
    }
  },
  "timestamp": "2025-07-09T01:49:25.123456789+03:00"
}
```

### Error Handling

```bash
# Invalid repository URL
curl -X POST -H "Content-Type: application/json" \
  -d '{"repository": "", "analyses": ["burndown"]}' \
  http://localhost:8080/api/v1/analyze

# Response:
{
  "error": "Repository URL is required",
  "code": "MISSING_REPOSITORY",
  "timestamp": "2025-07-09T01:48:19.943861855+03:00"
}
```

## SDK Examples

### Python

```python
import requests
import time

class HerculesClient:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
    
    def submit_analysis(self, repository, analyses, options=None):
        payload = {
            "repository": repository,
            "analyses": analyses
        }
        if options:
            payload["options"] = options
        
        response = requests.post(f"{self.base_url}/api/v1/analyze", json=payload)
        response.raise_for_status()
        return response.json()
    
    def get_status(self, job_id):
        response = requests.get(f"{self.base_url}/api/v1/status/{job_id}")
        response.raise_for_status()
        return response.json()
    
    def wait_for_completion(self, job_id, poll_interval=5):
        while True:
            status = self.get_status(job_id)
            if status["status"] in ["completed", "failed"]:
                return status
            time.sleep(poll_interval)

# Usage
client = HerculesClient()
job = client.submit_analysis(
    repository="https://github.com/dmytrogajewski/hercules.git",
    analyses=["burndown", "couples"]
)
result = client.wait_for_completion(job["job_id"])
print(f"Analysis completed: {result}")
```

### JavaScript

```javascript
class HerculesClient {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
    }
    
    async submitAnalysis(repository, analyses, options = {}) {
        const payload = {
            repository,
            analyses,
            ...(Object.keys(options).length > 0 && { options })
        };
        
        const response = await fetch(`${this.baseUrl}/api/v1/analyze`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return response.json();
    }
    
    async getStatus(jobId) {
        const response = await fetch(`${this.baseUrl}/api/v1/status/${jobId}`);
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return response.json();
    }
    
    async waitForCompletion(jobId, pollInterval = 5000) {
        while (true) {
            const status = await this.getStatus(jobId);
            if (status.status === 'completed' || status.status === 'failed') {
                return status;
            }
            await new Promise(resolve => setTimeout(resolve, pollInterval));
        }
    }
}

// Usage
const client = new HerculesClient();
client.submitAnalysis(
    'https://github.com/dmytrogajewski/hercules.git',
    ['burndown', 'couples']
).then(job => {
    return client.waitForCompletion(job.job_id);
}).then(result => {
    console.log('Analysis completed:', result);
}).catch(error => {
    console.error('Error:', error);
});
```

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Concurrent Jobs**: Maximum 10 concurrent analysis jobs (configurable)
- **Request Throttling**: Based on server load and available resources

When rate limits are exceeded, the API returns a `429 Too Many Requests` status code.

## Best Practices

1. **Always check job status**: Don't assume analysis is complete immediately
2. **Handle errors gracefully**: Implement proper error handling for all API calls
3. **Use appropriate timeouts**: Set reasonable timeouts for long-running analyses
4. **Cache results**: Store analysis results to avoid re-computation
5. **Monitor job status**: Implement polling to track analysis progress
6. **Handle rate limits**: Implement exponential backoff for rate limit errors

## Versioning

The API version is included in the URL path (`/api/v1/`). Future versions will maintain backward compatibility where possible.

## Support

For API support and questions:
- **Issues**: [GitHub Issues](https://github.com/dmytrogajewski/hercules/issues)
- **Discussions**: [GitHub Discussions](https://github.com/dmytrogajewski/hercules/discussions)
- **Documentation**: [Wiki](https://github.com/dmytrogajewski/hercules/wiki) 