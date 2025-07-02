# FlowCtl - Distributed Task Queue and Workflow Orchestrator

FlowCtl is a distributed task queue and workflow orchestrator built for managing complex workflows such as ETL pipelines, ML training pipelines, and CI/CD tasks. It provides at-least-once delivery guarantees, fault tolerance, and horizontal scalability.

## Architecture

### Core Components

- **Scheduler**: Go-based core engine that manages workflow execution and task scheduling
- **Workers**: Language-agnostic workers that execute tasks via REST/gRPC
- **State Store**: PostgreSQL for persistent metadata storage
- **Task Queue**: Redis for high-performance task queuing and coordination
- **Web Dashboard**: React-based UI for workflow visualization and monitoring

### Key Features

- **Workflow Definition**: YAML-based DSL for defining complex workflows
- **Task Dependencies**: Support for task dependencies and DAG execution
- **Retry Logic**: Configurable retry policies with exponential backoff
- **Dead Letter Queues**: Failed tasks are moved to dead letter queues after max retries
- **Worker Management**: Automatic worker registration and health monitoring
- **Real-time Monitoring**: Web dashboard with metrics and task visualization
- **Horizontal Scaling**: Support for multiple workers and schedulers

## Getting Started

### Prerequisites

- Go 1.21 or later
- PostgreSQL 12 or later
- Redis 6 or later
- Node.js 16 or later (for web dashboard)

### Installation

1. Clone the repository:
```bash
git clone 
cd flowctl
```

2. Install Go dependencies:
```bash
go mod download
```

3. Set up PostgreSQL database:
```sql
CREATE DATABASE flowctl;
CREATE USER flowctl_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE flowctl TO flowctl_user;
```

4. Start Redis server:
```bash
redis-server
```

5. Build the applications:
```bash
go build -o bin/scheduler cmd/scheduler/main.go
go build -o bin/worker cmd/worker/main.go
```

6. Build the web dashboard:
```bash
cd web/dashboard
npm install
npm run build
cd ../..
```

### Running the System

1. Start the scheduler:
```bash
./bin/scheduler \
  -postgres="postgres://flowctl_user:your_password@localhost/flowctl?sslmode=disable" \
  -redis="localhost:6379" \
  -api=":8080"
```

2. Start workers:
```bash
# ETL worker
./bin/worker -types="etl" -addr="localhost:9001"

# ML training worker  
./bin/worker -types="ml_training" -addr="localhost:9002"

# CI worker
./bin/worker -types="ci" -addr="localhost:9003"

# Generic worker
./bin/worker -types="generic" -addr="localhost:9004"
```

3. Access the web dashboard at http://localhost:8080

## Workflow Definition

Workflows are defined using YAML files with the following structure:

```yaml
name: "My Workflow"
description: "Description of what this workflow does"

config:
  max_concurrency: 10
  timeout: "1h"
  retry_policy:
    max_attempts: 3
    initial_delay: "10s"
    max_delay: "5m"
    backoff_factor: 2.0

tasks:
  - name: "task1"
    type: "etl"
    payload:
      source_url: "s3://bucket/data"
      target_url: "postgres://db/table"
    priority: 1
    
  - name: "task2"
    type: "generic"
    depends_on: ["task1"]
    payload:
      command: "process_data"
    priority: 2
    max_retries: 5
```

### Task Types

FlowCtl supports several built-in task types:

- **etl**: Extract, Transform, Load operations
- **ml_training**: Machine learning model training
- **ci**: Continuous integration tasks
- **generic**: General purpose command execution

### Configuration Options

- `max_concurrency`: Maximum number of tasks to run concurrently
- `timeout`: Maximum workflow execution time
- `retry_policy`: Retry configuration for failed tasks
- `priority`: Task execution priority (higher numbers execute first)
- `depends_on`: List of task dependencies

## API Reference

### Create Workflow

```http
POST /api/v1/workflows
Content-Type: application/json

{
  "name": "Test Workflow",
  "description": "A test workflow",
  "tasks": [
    {
      "name": "task1",
      "type": "generic",
      "payload": {
        "command": "echo hello"
      }
    }
  ]
}
```

### Get Workflow

```http
GET /api/v1/workflows/{id}
```

### Cancel Workflow

```http
PUT /api/v1/workflows/{id}/cancel
```

### Get Metrics

```http
GET /api/v1/metrics
```

## Worker Implementation

Workers can be implemented in any language that supports HTTP or gRPC. Here's a simple Python worker example:

```python
import requests
import time
import json
from redis import Redis

class Worker:
    def __init__(self, worker_id, redis_host, task_types):
        self.worker_id = worker_id
        self.redis = Redis(host=redis_host)
        self.task_types = task_types
        
    def start(self):
        while True:
            for task_type in self.task_types:
                task = self.dequeue_task(task_type)
                if task:
                    self.execute_task(task)
            time.sleep(1)
    
    def dequeue_task(self, task_type):
        queue_key = f"queue:{task_type}"
        task_data = self.redis.brpop(queue_key, timeout=30)
        if task_data:
            return json.loads(task_data[1])
        return None
    
    def execute_task(self, task):
        # Implement task execution logic
        pass
```

## Monitoring and Observability

### Web Dashboard

The web dashboard provides:

- Real-time workflow and task status
- Worker health monitoring
- System metrics and performance graphs
- Workflow visualization with dependency graphs

### Metrics

FlowCtl exposes metrics for:

- Workflow completion rates
- Task execution times
- Worker utilization
- Queue depths
- Error rates

### Health Checks

Health check endpoints:

- `/api/v1/health` - Overall system health
- `/api/v1/metrics` - Detailed metrics

## Configuration

### Environment Variables

- `POSTGRES_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `API_PORT`: API server port (default: 8080)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

### Command Line Options

Scheduler options:
- `-postgres`: PostgreSQL connection string
- `-redis`: Redis address
- `-api`: API server address

Worker options:
- `-redis`: Redis address
- `-types`: Comma-separated task types
- `-addr`: Worker address

## Deployment

### Docker

Build Docker images:

```bash
# Scheduler
docker build -t flowctl-scheduler -f Dockerfile.scheduler .

# Worker
docker build -t flowctl-worker -f Dockerfile.worker .
```

### Kubernetes

Deploy using Kubernetes manifests:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: flowctl-scheduler
spec:
  replicas: 2
  selector:
    matchLabels:
      app: flowctl-scheduler
  template:
    metadata:
      labels:
        app: flowctl-scheduler
    spec:
      containers:
      - name: scheduler
        image: flowctl-scheduler:latest
        env:
        - name: POSTGRES_URL
          value: "postgres://user:pass@postgres:5432/flowctl"
        - name: REDIS_URL  
          value: "redis:6379"
```

## Development

### Building from Source

```bash
# Build scheduler
go build -o bin/scheduler cmd/scheduler/main.go

# Build worker
go build -o bin/worker cmd/worker/main.go

# Build web dashboard
cd web/dashboard && npm run build
```

### Running Tests

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## Troubleshooting

### Common Issues

**Scheduler won't start**
- Check PostgreSQL and Redis connectivity
- Verify database permissions
- Check firewall settings

**Tasks not executing**
- Verify workers are registered and healthy
- Check Redis queue for pending tasks
- Review worker logs for errors

**Web dashboard not loading**
- Ensure dashboard was built (`npm run build`)
- Check API server is running
- Verify port 8080 is accessible

### Logs

Enable debug logging:
```bash
./bin/scheduler -postgres="..." -redis="..." -log-level=debug
```

Check Redis queues:
```bash
redis-cli LLEN queue:etl
redis-cli LRANGE queue:etl 0 -1
```

## Performance Tuning

### Scheduler Optimization

- Increase `max_concurrency` for CPU-bound workflows
- Tune PostgreSQL connection pool size
- Use Redis clustering for high throughput

### Worker Optimization

- Scale workers horizontally based on task types
- Use dedicated workers for resource-intensive tasks
- Monitor worker memory and CPU usage

### Database Optimization

- Add indexes on frequently queried columns
- Use connection pooling
- Consider read replicas for reporting

## Security

### Authentication

FlowCtl supports:
- API key authentication
- JWT tokens
- Role-based access control

### Network Security

- Use TLS for all connections
- Implement network segmentation
- Restrict database access

### Data Protection

- Encrypt sensitive payload data
- Use secrets management for credentials
- Implement audit logging

## License

This project is licensed under the MIT License. See LICENSE file for details.
