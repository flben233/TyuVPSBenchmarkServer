# Quick Start Guide

## Prerequisites
- Go 1.24 or higher
- Windows, Linux, or macOS

## Installation

### 1. Build the Server
```bash
go build -o benchmark_server.exe
```

### 2. Configure
Ensure `config.json` exists with the following structure:
```json
{
  "port": 12345,
  "baseUrl": ""
}
```

### 3. Run the Server
```bash
# Windows
.\benchmark_server.exe

# Linux/macOS
./benchmark_server
```

The server will:
- Create `./data` directory automatically
- Initialize SQLite database at `./data/benchmark.db`
- Create all necessary tables
- Start listening on the configured port

## Quick Test

### Add a Report
```bash
curl -X POST http://localhost:12345/report/admin/add \
  -H "Content-Type: application/json" \
  -d '{"html":"<html><head><meta name=\"title\" content=\"Test Report\"><meta name=\"time\" content=\"2025-11-29\"><meta name=\"link\" content=\"https://example.com\"></head><body>Test</body></html>"}'
```

### List Reports
```bash
curl http://localhost:12345/report/data/list
```

### Get Report Details
```bash
# Replace {id} with actual report ID from list response
curl http://localhost:12345/report/data/details?id={id}
```

### Search Reports
```bash
curl http://localhost:12345/report/data/search?keyword=Test
```

### Delete Report
```bash
# Replace {id} with actual report ID
curl -X POST http://localhost:12345/report/admin/delete \
  -H "Content-Type: application/json" \
  -d '{"id":"{id}"}'
```

## Using the Test Client

Rename the test client example:
```bash
# Windows
Move-Item test_client.go.example test_client_main.go

# Edit test_client_main.go to verify/adjust baseURL if needed
```

Run the test client (requires server to be running):
```bash
go run test_client_main.go
```

The test client will:
1. Add a sample report
2. List all reports
3. Get the report details
4. Search for reports
5. Delete the test report

## Directory Structure After First Run
```
TyuVPSBenchmarkServer/
├── data/
│   └── benchmark.db       # SQLite database (auto-created)
├── config.json            # Configuration file
├── benchmark_server.exe   # Compiled executable
└── ...                    # Source files
```

## Troubleshooting

### Port Already in Use
If you see "bind: address already in use", change the port in `config.json`:
```json
{
  "port": 8080,
  "baseUrl": ""
}
```

### Database Locked
If you see "database is locked" errors:
1. Close any SQLite database browsers
2. Restart the server

### Permission Denied
If you can't create the `./data` directory:
```bash
# Create manually with proper permissions
mkdir data
chmod 755 data
```

## API Documentation
See [REPORT_API.md](REPORT_API.md) for complete API documentation.

## Implementation Details
See [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) for implementation details.
