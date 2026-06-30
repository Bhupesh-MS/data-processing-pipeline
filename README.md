# Data Processing Pipeline

A robust, concurrent Go data processing pipeline that ingests data from CSV/JSON/API sources, processes records via generic validation, transformation, and aggregation rules, and exports the finalized records. 

## Features
- **Highly Concurrent**: Multi-stage pipeline using Goroutines and Channels.
- **Resilient**: Captures validation and execution errors at every stage in a persistent SQLite database instead of halting.
- **Observable**: A fully-featured REST API allowing job creation, cancellation, status tracking, and error querying.
- **Configurable**: Define your data sources, rules, aggregations, and SQLite/CSV/JSON export targets purely via JSON.

## Components
- `cmd/server`: HTTP server entrypoint.
- `internal/api/routes`: Route definitions and middleware.
- `internal/api/controllers`: HTTP request handlers.
- `internal/api/services`: Core business logic and job orchestration.
- `internal/api/repositories`: Data access layer interfaces and models.
- `internal/pipeline`: Orchestrator, channels, worker pool execution, and context management.
- `internal/ingestion`: Connectors for JSON APIs, and CSV/JSON files.
- `internal/validation` & `internal/transformation`: Worker nodes applying configurable business rules.
- `internal/aggregation`: Fan-in aggregation logic.
- `internal/export`: Configurable multi-target export (SQLite, CSV, JSON).
- `internal/storage`: SQLite repository implementation (`storage.SQLiteStore`).

## Getting Started

1. **Install dependencies:**
   ```bash
   go mod tidy
   ```

2. **Run the server:**
   ```bash
   go run ./cmd/server
   ```
   (Optional: override `SERVER_ADDR` or `DB_PATH` environment variables).

3. **Run tests (Unit & Integration):**
   ```bash
   # Run all tests
   go test -v ./...
   
   # Check test coverage (>90% for API layers)
   go test -coverprofile=coverage.out ./...
   go tool cover -func=coverage.out
   
   # View coverage in browser
   go tool cover -html=coverage.out
   ```

## Example API Requests

### 1. Check Health
```bash
curl http://localhost:8080/health
```

### 2. Create a Pipeline Job
This example fetches a JSON list, validates it, transforms it, aggregates the scores, and exports to a SQLite table.
```bash
curl -X POST http://localhost:8080/api/v1/pipelines \
  -H "Content-Type: application/json" \
  -d '{
    "sources": [
        {"type": "json", "url": "https://jsonplaceholder.typicode.com/users"}
    ],
    "validations": [
        {"field": "id", "type": "int", "required": true}
    ],
    "transformations": [
        {"field": "name", "action": "uppercase"}
    ],
    "exports": [
        {"type": "sqlite", "table": "users_export"},
        {"type": "json", "path": "output.json"}
    ],
    "worker_count": 5
  }'
```

### 3. Track Progress & Metrics
Replace `<job-id>` with the ID returned by the POST request.
```bash
curl http://localhost:8080/api/v1/pipelines/<job-id>/progress
```

### 4. Fetch Failed Records / Errors
```bash
curl http://localhost:8080/api/v1/pipelines/<job-id>/errors
```

### 5. Cancel Running Job
```bash
curl -X PATCH http://localhost:8080/api/v1/pipelines/<job-id>/cancel
```
