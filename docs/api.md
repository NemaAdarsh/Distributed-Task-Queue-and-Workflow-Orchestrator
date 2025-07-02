# FlowCtl API Documentation

## Base URL

All API endpoints are prefixed with `/api/v1`

## Authentication

Currently, the API does not require authentication. In production deployments, implement JWT or API key authentication.

## Content Type

All requests and responses use `application/json` content type.

## Error Handling

The API returns standard HTTP status codes:

- `200 OK` - Successful request
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

Error responses include a JSON object with an error message:

```json
{
  "error": "Description of the error"
}
```

## Endpoints

### Workflows

#### Create Workflow

Creates a new workflow and queues it for execution.

**POST** `/api/v1/workflows`

**Request Body:**

```json
{
  "name": "string (required)",
  "description": "string (optional)",
  "config": {
    "max_concurrency": "integer (optional, default: 10)",
    "timeout": "string (optional, default: 1h)",
    "retry_policy": {
      "max_attempts": "integer (optional, default: 3)",
      "initial_delay": "string (optional, default: 1s)", 
      "max_delay": "string (optional, default: 5m)",
      "backoff_factor": "float (optional, default: 2.0)"
    }
  },
  "tasks": [
    {
      "name": "string (required)",
      "type": "string (required)",
      "payload": "object (optional)",
      "max_retries": "integer (optional, default: 3)",
      "priority": "integer (optional, default: 1)",
      "dependencies": "array of strings (optional)"
    }
  ]
}
```

**Response:**

```json
{
  "id": "uuid",
  "name": "string",
  "description": "string",
  "status": "pending",
  "tasks": [...],
  "config": {...},
  "created_at": "ISO 8601 timestamp",
  "updated_at": "ISO 8601 timestamp"
}
```

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Data Processing",
    "description": "Process daily data files",
    "tasks": [
      {
        "name": "extract_data",
        "type": "etl",
        "payload": {
          "source_url": "s3://bucket/data.csv",
          "target_url": "postgres://db/staging"
        },
        "priority": 1
      },
      {
        "name": "validate_data", 
        "type": "generic",
        "dependencies": ["extract_data"],
        "payload": {
          "command": "validate_schema"
        },
        "priority": 2
      }
    ]
  }'
```

#### Get Workflow

Retrieves a specific workflow by ID.

**GET** `/api/v1/workflows/{id}`

**Parameters:**
- `id` (path) - Workflow ID

**Response:**

```json
{
  "id": "uuid",
  "name": "string",
  "description": "string", 
  "status": "pending|running|completed|failed|cancelled",
  "tasks": [
    {
      "id": "uuid",
      "workflow_id": "uuid",
      "name": "string",
      "type": "string",
      "status": "pending|running|completed|failed|retrying|cancelled",
      "payload": "object",
      "result": "object",
      "error": "string",
      "retry_count": "integer",
      "max_retries": "integer",
      "priority": "integer",
      "dependencies": "array of strings",
      "created_at": "ISO 8601 timestamp",
      "updated_at": "ISO 8601 timestamp",
      "started_at": "ISO 8601 timestamp",
      "completed_at": "ISO 8601 timestamp"
    }
  ],
  "config": "object",
  "created_at": "ISO 8601 timestamp",
  "updated_at": "ISO 8601 timestamp",
  "started_at": "ISO 8601 timestamp",
  "completed_at": "ISO 8601 timestamp"
}
```

#### List Workflows

Retrieves a paginated list of workflows.

**GET** `/api/v1/workflows`

**Query Parameters:**
- `page` (optional) - Page number (default: 1)
- `limit` (optional) - Number of items per page (default: 10)
- `status` (optional) - Filter by status

**Response:**

```json
{
  "workflows": [
    {
      "id": "uuid",
      "name": "string",
      "description": "string",
      "status": "string",
      "tasks": "array",
      "created_at": "ISO 8601 timestamp",
      "updated_at": "ISO 8601 timestamp"
    }
  ],
  "total": "integer",
  "page": "integer", 
  "limit": "integer"
}
```

#### Cancel Workflow

Cancels a running workflow.

**PUT** `/api/v1/workflows/{id}/cancel`

**Parameters:**
- `id` (path) - Workflow ID

**Response:**

```json
{
  "message": "Workflow cancelled"
}
```

### Tasks

#### Get Task

Retrieves a specific task by ID.

**GET** `/api/v1/tasks/{id}`

**Parameters:**
- `id` (path) - Task ID

**Response:**

```json
{
  "id": "uuid",
  "workflow_id": "uuid", 
  "name": "string",
  "type": "string",
  "status": "pending|running|completed|failed|retrying|cancelled",
  "payload": "object",
  "result": "object",
  "error": "string",
  "retry_count": "integer",
  "max_retries": "integer",
  "priority": "integer",
  "dependencies": "array of strings",
  "created_at": "ISO 8601 timestamp",
  "updated_at": "ISO 8601 timestamp",
  "started_at": "ISO 8601 timestamp",
  "completed_at": "ISO 8601 timestamp"
}
```

#### Get Workflow Tasks

Retrieves all tasks for a specific workflow.

**GET** `/api/v1/workflows/{id}/tasks`

**Parameters:**
- `id` (path) - Workflow ID

**Response:**

```json
{
  "tasks": [
    {
      "id": "uuid",
      "name": "string",
      "type": "string",
      "status": "string",
      "priority": "integer",
      "dependencies": "array",
      "created_at": "ISO 8601 timestamp",
      "started_at": "ISO 8601 timestamp",
      "completed_at": "ISO 8601 timestamp"
    }
  ]
}
```

### System

#### Health Check

Returns the health status of the system.

**GET** `/api/v1/health`

**Response:**

```json
{
  "status": "healthy|unhealthy",
  "timestamp": "ISO 8601 timestamp"
}
```

#### Get Metrics

Returns system metrics and statistics.

**GET** `/api/v1/metrics`

**Response:**

```json
{
  "workflows": {
    "total": "integer",
    "running": "integer", 
    "completed": "integer",
    "failed": "integer"
  },
  "tasks": {
    "total": "integer",
    "pending": "integer",
    "running": "integer",
    "completed": "integer", 
    "failed": "integer"
  },
  "workers": {
    "active": "integer",
    "idle": "integer"
  }
}
```

## Task Types

FlowCtl supports the following built-in task types:

### ETL Tasks (`etl`)

Used for Extract, Transform, Load operations.

**Common Payload Fields:**
- `source_url`: Source data location
- `target_url`: Target data location  
- `format`: Data format (json, csv, parquet, etc.)
- `transformation`: Transformation type

**Example:**
```json
{
  "type": "etl",
  "payload": {
    "source_url": "s3://bucket/data.json",
    "target_url": "postgres://db/table", 
    "format": "json",
    "transformation": "normalize"
  }
}
```

### ML Training Tasks (`ml_training`)

Used for machine learning model training and validation.

**Common Payload Fields:**
- `model_name`: Name of the model
- `dataset_url`: Training dataset location
- `algorithm`: ML algorithm to use
- `hyperparameters`: Model hyperparameters

**Example:**
```json
{
  "type": "ml_training",
  "payload": {
    "model_name": "fraud_detector",
    "dataset_url": "s3://ml-data/training.csv",
    "algorithm": "xgboost",
    "hyperparameters": {
      "max_depth": 6,
      "learning_rate": 0.1
    }
  }
}
```

### CI Tasks (`ci`)

Used for continuous integration operations.

**Common Payload Fields:**
- `repo_url`: Repository URL
- `command`: Command to execute
- `branch`: Git branch
- `environment`: Target environment

**Example:**
```json
{
  "type": "ci", 
  "payload": {
    "repo_url": "https://github.com/user/repo.git",
    "command": "npm test",
    "branch": "main",
    "environment": "test"
  }
}
```

### Generic Tasks (`generic`)

Used for general command execution.

**Common Payload Fields:**
- `command`: Command to execute
- `args`: Command arguments
- `environment`: Environment variables

**Example:**
```json
{
  "type": "generic",
  "payload": {
    "command": "python script.py",
    "args": ["--input", "data.csv"],
    "environment": {
      "PYTHONPATH": "/opt/scripts"
    }
  }
}
```

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Rate**: 100 requests per minute per IP
- **Headers**: Rate limit information is returned in response headers:
  - `X-RateLimit-Limit`: Request limit per window
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `X-RateLimit-Reset`: Unix timestamp when the window resets

When rate limit is exceeded, the API returns HTTP 429 with:

```json
{
  "error": "Rate limit exceeded. Try again later."
}
```

## Pagination

List endpoints support pagination using query parameters:

- `page`: Page number (1-based, default: 1)
- `limit`: Items per page (default: 10, max: 100)

Response includes pagination metadata:

```json
{
  "data": [...],
  "total": 150,
  "page": 1,
  "limit": 10,
  "total_pages": 15
}
```

## Error Codes

| Status Code | Description |
|-------------|-------------|
| 400 | Bad Request - Invalid input data |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource does not exist |
| 409 | Conflict - Resource already exists |
| 422 | Unprocessable Entity - Validation failed |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |
| 502 | Bad Gateway - Upstream service error |
| 503 | Service Unavailable - Service temporarily unavailable |

## SDKs and Client Libraries

### Go Client

```go
import "github.com/flowctl/go-client"

client := flowctl.NewClient("http://localhost:8080")
workflow, err := client.CreateWorkflow(ctx, workflowRequest)
```

### Python Client

```python
from flowctl import Client

client = Client("http://localhost:8080")
workflow = client.create_workflow(workflow_data)
```

### JavaScript Client

```javascript
import { FlowCtlClient } from '@flowctl/client';

const client = new FlowCtlClient('http://localhost:8080');
const workflow = await client.createWorkflow(workflowData);
```

## Webhooks

FlowCtl supports webhooks for real-time notifications of workflow and task events.

### Configuration

Configure webhooks via environment variables or configuration file:

```yaml
webhooks:
  - url: "https://example.com/webhook"
    events: ["workflow.completed", "workflow.failed"]
    secret: "webhook_secret"
```

### Events

- `workflow.created`
- `workflow.started`
- `workflow.completed`
- `workflow.failed`
- `workflow.cancelled`
- `task.started`
- `task.completed`
- `task.failed`
- `task.retrying`

### Payload

```json
{
  "event": "workflow.completed",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "workflow_id": "uuid",
    "workflow_name": "string",
    "status": "completed"
  }
}
```
