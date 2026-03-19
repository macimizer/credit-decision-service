# Credit Decision & Management Service

Production-oriented backend technical test implementation in Go for the **Tu Crédito** assignment.

## What is implemented

- REST API for **clients**, **banks**, and **credits**
- PostgreSQL persistence with embedded SQL migrations
- Redis caching for frequent reads
- Redis Streams domain events:
  - `CreditCreated`
  - `CreditApproved`
- Concurrent credit decision flow using:
  - goroutines
  - channels
  - worker pools
- Structured logging with `log/slog`
- Prometheus metrics
- Health endpoints
- `pprof` profiling endpoints
- Unit tests, repository tests, integration tests, and benchmarks
- Dockerfile, Docker Compose, Makefile, CI pipeline

## Project structure

```text
.
├── cmd/api
├── docs
├── internal
│   ├── config
│   ├── domain
│   ├── infrastructure
│   ├── observability
│   ├── platform
│   ├── ports
│   ├── service
│   └── transport
├── tests/integration
├── .github/workflows/ci.yml
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## Architecture overview

See [docs/architecture.md](docs/architecture.md).

## Technology stack and tested versions

### Core runtime

- **Go**: 1.23.2
- **PostgreSQL**: 16.4
- **Redis**: 7.4.2
- **Docker Compose**: v2+

### Main libraries

- `github.com/go-chi/chi/v5` v5.2.1
- `github.com/redis/go-redis/v9` v9.7.1
- `github.com/prometheus/client_golang` v1.21.1
- `github.com/jackc/pgx/v5` v5.7.2
- `github.com/stretchr/testify` v1.10.0

## Functional scope mapped to the assignment

### Block 0 – Backend fundamentals

Implemented:
- well-structured project layout (`cmd`, `internal`, `docs`, `tests`)
- REST API for clients, banks, credits
- context-aware request handling
- structured logging
- decoupled repositories
- SQL migrations
- Dockerfile + Docker Compose

### Block A – Concurrency & performance

Implemented:
- goroutines, channels, and explicit worker pool
- concurrent validation and scoring during credit creation
- `pprof` endpoints under `/debug/pprof`
- benchmark suite with `go test -bench`
- Redis cache for frequent read endpoints

### Block B – Distributed architecture

Implemented:
- domain events published to Redis Streams
- clear separation between handler/service/domain/infrastructure
- metrics and health endpoints

### Block C – Testing & CI/CD

Implemented:
- unit tests for decision engine and credit service
- repository tests
- integration tests against a real PostgreSQL database
- GitHub Actions pipeline for lint/test/build

### Block D – Optional AdTech track

Included as a lightweight bonus:
- extensible rule-based decision engine interface
- clear extension point for routing and prioritization logic

## API endpoints

### Health and observability

- `GET /healthz`
- `GET /readyz`
- `GET /metrics`
- `GET /debug/pprof/`

### Clients

- `POST /api/v1/clients`
- `GET /api/v1/clients`
- `GET /api/v1/clients/{clientID}`

### Banks

- `POST /api/v1/banks`
- `GET /api/v1/banks`
- `GET /api/v1/banks/{bankID}`

### Credits

- `POST /api/v1/credits`
- `GET /api/v1/credits`
- `GET /api/v1/credits/{creditID}`

## Step-by-step Linux setup

The instructions below assume **Ubuntu 22.04+** or any modern Linux distribution with equivalent package names.

### 1) Install required tools

#### Option A — local Go runtime + local Docker

```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl git make
```

Install Go 1.23.2:

```bash
curl -LO https://go.dev/dl/go1.23.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz
rm go1.23.2.linux-amd64.tar.gz
```

Add Go to your shell profile:

```bash
echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc
echo 'export PATH=$HOME/go/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
```

Verify:

```bash
go version
```

Install Docker Engine + Docker Compose plugin, following your distro’s official Docker instructions.
Verify:

```bash
docker --version
docker compose version
```

#### Option B — Docker-only runtime

You can also run the service using Docker Compose without installing Go locally. You only need Docker Engine + Docker Compose plugin.

### 2) Unzip the project and enter the directory

```bash
unzip credit-decision-service.zip
cd credit-decision-service
```

### 3) Create the environment file

```bash
cp .env.example .env
```

Default values already work for Docker Compose.

### 4) Run the full stack with Docker Compose

```bash
docker compose up --build
```

This starts:
- API on `http://localhost:8080`
- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`

### 5) Verify the service

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

Expected responses:

```json
{"status":"ok"}
{"status":"ready"}
```

## Step-by-step local Go execution on Linux

If you prefer to run the API directly with Go while PostgreSQL and Redis run in Docker:

### 1) Start only infrastructure containers

```bash
docker compose up -d postgres redis
```

### 2) Export environment variables

```bash
export APP_NAME=credit-decision-service
export HTTP_ADDR=:8080
export LOG_LEVEL=INFO
export POSTGRES_DSN='postgres://postgres:postgres@localhost:5432/credit_service?sslmode=disable'
export REDIS_ADDR='localhost:6379'
export REDIS_PASSWORD=''
export REDIS_DB=0
export CACHE_TTL=5m
export EVENT_STREAM_NAME='credit-events'
export WORKER_COUNT=4
export AUTO_MIGRATE=true
```

### 3) Download dependencies and run the service

```bash
go mod tidy
go run ./cmd/api
```

## Build commands

```bash
make tidy
make fmt
make vet
make test
make bench
make build
```

Binary output:

```bash
./bin/credit-decision-service
```

## Test commands

### Unit tests

```bash
make test-unit
```

### Integration tests

Start PostgreSQL first, then run:

```bash
export TEST_DATABASE_DSN='postgres://postgres:postgres@localhost:5432/credit_service?sslmode=disable'
export RUN_INTEGRATION_TESTS=true
make test-integration
```

## Example API flow

### 1) Create a client

```bash
curl -X POST http://localhost:8080/api/v1/clients   -H 'Content-Type: application/json'   -d '{
    "full_name": "Ana Gomez",
    "email": "ana.gomez@example.com",
    "birth_date": "1992-07-10",
    "country": "CO"
  }'
```

### 2) Create a bank

```bash
curl -X POST http://localhost:8080/api/v1/banks   -H 'Content-Type: application/json'   -d '{
    "name": "Banco Central Demo",
    "type": "GOVERNMENT"
  }'
```

### 3) Create a credit

Replace `<CLIENT_ID>` and `<BANK_ID>` with the values returned above.

```bash
curl -X POST http://localhost:8080/api/v1/credits   -H 'Content-Type: application/json'   -d '{
    "client_id": "<CLIENT_ID>",
    "bank_id": "<BANK_ID>",
    "min_payment": 100.00,
    "max_payment": 450.00,
    "term_months": 24,
    "credit_type": "AUTO"
  }'
```

## Redis Streams inspection

To inspect emitted events:

```bash
docker exec -it credit-decision-service-redis redis-cli
XREAD COUNT 10 STREAMS credit-events 0-0
```

## Observability

### Metrics

```bash
curl http://localhost:8080/metrics
```

### pprof CPU profile

```bash
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=20
```

## Performance notes and bottlenecks

Potential bottlenecks considered:

1. **Database round trips during decisioning**
   - mitigated by concurrent repository fetches for client and bank
2. **Repeated reads for reference entities**
   - mitigated with Redis caching for clients and banks
3. **Blocking event publication**
   - mitigated with asynchronous event publishing after persistence
4. **Handler latency visibility gaps**
   - mitigated with request metrics and pprof endpoints

## Technical decisions

### Why Chi

- lightweight and idiomatic
- easy middleware composition
- production-friendly route structure without unnecessary abstraction

### Why `database/sql` with pgx driver

- explicit SQL control
- simple operational behavior
- avoids ORM overhead for a latency-sensitive backend

### Why Redis Streams

- lightweight event-driven integration point
- simple local setup compared to Kafka for a technical challenge
- easy path to replace through the `EventPublisher` interface

### Why embedded migrations

- zero external migration binary required
- simpler reviewer experience
- keeps project bootstrapping predictable in CI and Docker

## AI usage disclosure

This assignment explicitly asks for transparent AI usage when applicable.

- AI assistance was used to accelerate drafting and packaging.
- Architecture, trade-offs, naming, layering, concurrency model, and implementation choices were deliberately reviewed and shaped to follow a senior backend style.
- All code paths, README instructions, and deliverables should still be reviewed by the submitting engineer before handing in the test.

## What I would build next in a real production iteration

- transactional outbox pattern for event publication durability
- idempotency keys for credit creation
- richer rate limiting and authentication
- pagination and filtering
- OpenTelemetry traces
- config-backed rule engine for routing and monetization extensions
