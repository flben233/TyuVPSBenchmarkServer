# Report Module API Documentation

## Overview
The report module provides REST API endpoints for managing VPS benchmark reports with SQLite database storage using GORM.

## Database
- **Location**: `./data/benchmark.db`
- **Auto-created**: Database and tables are automatically created on first run
- **ORM**: GORM with SQLite driver

## API Endpoints

### Data Endpoints (Public)

#### 1. List Reports
Get a paginated list of all reports.

**Endpoint**: `GET /report/data/list`

**Query Parameters**:
- `page` (optional, default: 1): Page number
- `page_size` (optional, default: 10): Number of items per page

**Response**:
```json
{
  "data": [
    {
      "name": "Report Title",
      "id": "abc123...",
      "date": "2025-11-29"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 10
}
```

#### 2. Get Report Details
Get full details of a specific report.

**Endpoint**: `GET /report/data/details`

**Query Parameters**:
- `id` (required): Report ID

**Response**:
```json
{
  "data": {
    "id": "abc123...",
    "title": "Report Title",
    "time": "2025-11-29",
    "link": "https://...",
    "spdtest": [...],
    "ecs": {...},
    "media": {...},
    "besttrace": [...],
    "itdog": {...},
    "disk": {...}
  }
}
```

#### 3. Search Reports
Search reports based on various criteria.

**Endpoint**: `GET /report/data/search` or `POST /report/data/search`

**Query Parameters** (for GET):
- `keyword` (optional): Search keyword (matches title)
- `page` (optional, default: 1): Page number
- `page_size` (optional, default: 10): Number of items per page

**Request Body** (for POST):
```json
{
  "name": "keyword",
  "media_unlocks": ["Netflix", "YouTube"],
  "ct_params": {
    "back_route": "CN2 GIA",
    "min_download": 100.0,
    "max_download": 1000.0,
    "min_upload": 50.0,
    "max_upload": 500.0,
    "latency": 50.0
  },
  "cu_params": {...},
  "cm_params": {...},
  "virtualization": "KVM",
  "ipv6_support": true,
  "disk_level": 3
}
```

**Response**: Same format as List Reports

### Admin Endpoints

#### 4. Add Report
Add a new benchmark report.

**Endpoint**: `POST /report/admin/add`

**Request Body** (JSON):
```json
{
  "html": "<html>...</html>"
}
```

Or send raw HTML directly in the request body.

**Response**:
```json
{
  "message": "Report added successfully",
  "report_id": "abc123..."
}
```

#### 5. Delete Report
Delete a report by ID.

**Endpoint**: `POST /report/admin/delete`

**Request Body** (JSON):
```json
{
  "id": "abc123..."
}
```

Or use query parameter: `?id=abc123...`

**Response**:
```json
{
  "message": "Report deleted successfully"
}
```

## Database Schema

### BenchmarkResult Table
| Field | Type | Description |
|-------|------|-------------|
| id | INTEGER | Primary key (auto-increment) |
| report_id | VARCHAR(255) | Unique report identifier |
| title | VARCHAR(500) | Report title |
| time | VARCHAR(100) | Report timestamp |
| link | VARCHAR(1000) | Report link |
| raw_html | TEXT | Original HTML content |
| spdtest | TEXT | JSON: Speed test results |
| ecs | TEXT | JSON: ECS (system info) results |
| media | TEXT | JSON: Media unlock results |
| besttrace | TEXT | JSON: Best trace results |
| itdog | TEXT | JSON: ITDog results |
| disk | TEXT | JSON: Disk benchmark results |
| created_at | DATETIME | Record creation time |
| updated_at | DATETIME | Record last update time |

## Data Models

### JSONField Structure
Complex nested data (speedtest, ecs, media, etc.) is stored as JSON text in SQLite using the custom `JSONField` type that handles serialization/deserialization automatically.

## Running the Server

1. Ensure `config.json` exists with proper configuration
2. Run the server: `go run main.go`
3. Database will be automatically created at `./data/benchmark.db`
4. Server will start on the configured port (default: 12345)

## Example Usage

### Add a Report
```bash
curl -X POST http://localhost:12345/report/admin/add \
  -H "Content-Type: application/json" \
  -d '{"html":"<html>...</html>"}'
```

### List Reports
```bash
curl http://localhost:12345/report/data/list?page=1&page_size=10
```

### Get Report Details
```bash
curl http://localhost:12345/report/data/details?id=abc123
```

### Search Reports
```bash
curl http://localhost:12345/report/data/search?keyword=test
```

### Delete Report
```bash
curl -X POST http://localhost:12345/report/admin/delete \
  -H "Content-Type: application/json" \
  -d '{"id":"abc123"}'
```

## Features

- ✅ Automatic database initialization on first run
- ✅ GORM-based ORM with SQLite
- ✅ Full CRUD operations
- ✅ Pagination support
- ✅ Advanced search with multiple criteria
- ✅ JSON field storage for complex nested data
- ✅ Automatic timestamp tracking (created_at, updated_at)
- ✅ Unique report ID generation
- ✅ Raw HTML storage for archival purposes
