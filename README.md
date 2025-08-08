# Beacon - API Monitoring Service

Beacon is a scalable API monitoring service built with Go and Temporal for workflow orchestration.

## Features

- Service and endpoint management
- Automated endpoint health checks
- Incident tracking and resolution
- Webhook notifications for incidents
- Aggregated metrics with time windows
- PostgreSQL for data persistence
- Temporal for reliable workflow execution

## Architecture

- **CLI**: Command-line interface for CRUD operations
- **Worker**: Background service using Temporal for monitoring tasks
- **Database**: PostgreSQL for storing services, endpoints, pings, incidents, and webhooks

## Getting Started

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- PostgreSQL client (for migrations)

### Quick Start

1. Start the infrastructure:
```bash
make up
```

2. Run database migrations:
```bash
make migrate
```

3. Build the binaries:
```bash
make build
```

### Using the CLI

Create a service:
```bash
./bin/beacon services create --name "My API" --description "Production API"
```

Add an endpoint:
```bash
./bin/beacon endpoints create \
  --service-id <service-uuid> \
  --name "Health Check" \
  --url "https://api.example.com/health" \
  --interval 60
```

List services:
```bash
./bin/beacon services list
```

View incidents:
```bash
./bin/beacon incidents list --status open
```

### Running the Worker

The worker automatically starts with Docker Compose. To run locally:

```bash
make run-worker
```

## CLI Commands

### Services
- `services create` - Create a new service
- `services list` - List all services
- `services get <id>` - Get service details
- `services update <id>` - Update service
- `services delete <id>` - Delete service

### Endpoints
- `endpoints create` - Create endpoint
- `endpoints list` - List endpoints
- `endpoints get <id>` - Get endpoint details
- `endpoints update <id>` - Update endpoint
- `endpoints delete <id>` - Delete endpoint

### Pings
- `pings get <id>` - Get ping details
- `pings list --endpoint-id <id>` - List pings for endpoint

### Incidents
- `incidents list` - List incidents
- `incidents get <id>` - Get incident details
- `incidents resolve <id>` - Resolve incident

### Webhooks
- `webhooks create` - Create webhook
- `webhooks list` - List webhooks
- `webhooks update <id>` - Update webhook
- `webhooks delete <id>` - Delete webhook

## Development

Run tests:
```bash
make test
```

Clean up Docker volumes:
```bash
make clean
```

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string
- `TEMPORAL_HOST` - Temporal server address (default: localhost:7233)