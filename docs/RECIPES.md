# Hercules Integration Recipes

This document provides practical recipes for integrating Hercules with popular developer tools and platforms, including GitHub, GitLab, Grafana, and more. Use these examples to automate analysis, visualize results, and enhance your DevOps workflows.

---

## 1. Integrating Hercules with GitHub Actions

**Goal:** Run Hercules analysis on every push and upload results as an artifact.

```yaml
name: Hercules Analysis
on: [push]
jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install Hercules
        run: |
          curl -L https://github.com/dmytrogajewski/hercules/releases/latest/download/hercules-linux-amd64 -o hercules
          chmod +x hercules
      - name: Run Hercules
        run: |
          ./hercules --burndown --pb . > analysis.pb
      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: hercules-analysis
          path: analysis.pb
```

---

## 2. Integrating Hercules with GitLab CI/CD

**Goal:** Run Hercules in a GitLab pipeline and save results as a job artifact.

```yaml
stages:
  - analyze

hercules_analysis:
  stage: analyze
  image: golang:1.21
  script:
    - curl -L https://github.com/dmytrogajewski/hercules/releases/latest/download/hercules-linux-amd64 -o hercules
    - chmod +x hercules
    - ./hercules --burndown --pb . > analysis.pb
  artifacts:
    paths:
      - analysis.pb
```

---

## 3. Visualizing Hercules Data in Grafana

**Goal:** Send Hercules metrics to Prometheus and visualize in Grafana dashboards.

### Step 1: Export Hercules Data to Prometheus Format
- Use a custom script or Hercules output in JSON, then transform to Prometheus metrics using a sidecar or exporter.

**Example exporter script (Python):**
```python
import json
from prometheus_client import start_http_server, Gauge
import time

with open('hercules_output.json') as f:
    data = json.load(f)

burndown_gauge = Gauge('hercules_burndown_lines', 'Burndown lines', ['file'])
for file, stats in data['burndown']['files'].items():
    burndown_gauge.labels(file=file).set(stats['lines'])

start_http_server(8000)
while True:
    time.sleep(10)
```

### Step 2: Add Prometheus as a Data Source in Grafana
- Configure Prometheus to scrape the exporter endpoint.
- Create Grafana dashboards using the `hercules_burndown_lines` metric.

---

## 4. Automating Hercules Analysis with Webhooks

**Goal:** Trigger Hercules analysis on repository events (push, merge) using webhooks.

- Set up a webhook in GitHub/GitLab to POST to a CI/CD system or a custom server running Hercules.
- Example: Use a small HTTP server to receive webhook events and run Hercules.

```python
from flask import Flask, request
import subprocess

app = Flask(__name__)

@app.route('/webhook', methods=['POST'])
def webhook():
    subprocess.Popen(['./hercules', '--burndown', '--pb', '.'])
    return '', 204

if __name__ == '__main__':
    app.run(port=9000)
```

---

## 5. Integrating Hercules with Dashboards and Reporting Tools

- Export Hercules results as JSON or CSV for import into BI tools (Tableau, PowerBI, etc).
- Use the HTTP or gRPC API to automate analysis and fetch results for dashboards.

**Example: Fetching results via HTTP API**
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"repository":"https://github.com/dmytrogajewski/hercules.git","analyses":["burndown"]}' \
  http://localhost:8080/api/v1/analyze
```

---

## 6. Hercules as a Microservice

- Deploy Hercules as a server (HTTP/gRPC) in your infrastructure.
- Integrate with CI/CD, dashboards, or other microservices via API.
- Use Docker for easy deployment:

```bash
docker run -p 8080:8080 dmytrogajewski/hercules hercules server
```

---

## 7. Tips for Large-Scale and Automation

- Use caching (`/tmp/hercules-cache`) for faster repeated analyses.
- Schedule periodic analyses with cron or CI/CD schedules.
- Monitor server health with `/health` endpoint or gRPC Health API.
- Use environment variables or config files for flexible deployment.

---

For more advanced recipes or to contribute your own, see the [docs/DEVELOPMENT.md](DEVELOPMENT.md) and [docs/API.md](API.md) files. 