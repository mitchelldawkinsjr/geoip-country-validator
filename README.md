# GeoIP Service

HTTP API service that validates IP addresses against allowed countries using MaxMind's GeoLite2 database.

## Features

- **HTTP API**: RESTful endpoint for IP country validation
- **Production Ready**: Structured logging, health checks, graceful shutdown
- **Docker Support**: Multi-stage builds with security best practices
- **Configurable**: Environment-based configuration

## AI Tool Usage

This solution was developed using Claude (Anthropic) as a coding assistant. The AI was used for:

- **Architecture planning** and design decisions
- **Code generation** for Go source files
- **Documentation** writing and formatting
- **Best practices** application for Go development

The AI provided comprehensive solutions that follow Go idioms and production-ready patterns. I reviewed all the code, optimized any issues I found, and ensured everything is in working order. All code was thoroughly tested for correctness and functionality.

## Quick Start

### Prerequisites

1. **Go 1.24+** installed
2. **MaxMind GeoLite2 Database**: Download `GeoLite2-Country.mmdb` from [MaxMind](https://dev.maxmind.com/geoip/geoip2/geolite2/)
   - Sign up for a free account
   - Download the GeoLite2 Country database
   - Place `GeoLite2-Country.mmdb` in the project root

### Running Locally

```bash
# Run directly
go run ./cmd/geoip-server

# Or build first
go build -o geoip-server ./cmd/geoip-server
./geoip-server
```

### Docker

```bash
# Build and run
docker build -t geoip-service:latest .
mkdir -p ./data && cp GeoLite2-Country.mmdb ./data/
docker run -d --name geoip-service -p 8080:8080 -v $(pwd)/data:/app/data:ro geoip-service:latest
```

## API Usage

### `POST /v1/check`

Validates if an IP address is from an allowed country.

**Request:**
```json
{
  "ip_address": "81.2.69.142",
  "allowed_countries": ["US", "GB", "CA"]
}
```

**Response:**
```json
{
  "allowed": true,
  "country": "GB"
}
```

### Health Check

- `GET /health` - Service health status

### Testing

```bash
# Run tests
go test ./...

# Test the API
curl -X POST http://localhost:8080/v1/check \
  -H "Content-Type: application/json" \
  -d '{"ip_address":"8.8.8.8","allowed_countries":["US"]}'
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `GEOIP_DB_PATH` | `./GeoLite2-Country.mmdb` | Path to MaxMind database file |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

Example:
```bash
PORT=3000 LOG_LEVEL=debug go run ./cmd/geoip-server
```

## Architecture

```
├── cmd/geoip-server/     # Application entry point
├── internal/
│   ├── api/              # HTTP handlers and routing
│   ├── config/           # Configuration management
│   └── geoip/            # GeoIP lookup logic
├── Dockerfile            # Docker container definition
└── README.md            # This file
```

### Key Design Decisions

- **Separation of Concerns**: Clear boundaries between HTTP, business logic, and configuration layers
- **Dependency Injection**: Services injected into handlers for testability
- **Structured Logging**: JSON-formatted logs for observability
- **Graceful Shutdown**: Proper signal handling for zero-downtime deployments
- **Security**: Input validation, error handling, non-root container user

## Development

```bash
# Format code
go fmt ./...

# Run tests with verbose output
go test -v ./...

# Build for production
go build -ldflags="-w -s" -o geoip-server ./cmd/geoip-server
```
