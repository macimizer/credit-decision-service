APP_NAME := credit-decision-service

.PHONY: tidy fmt vet lint test test-unit test-integration bench build run docker-build

tidy:
	go mod tidy

fmt:
	gofmt -w $(shell find . -type f -name '*.go')

vet:
	go vet ./...

lint:
	golangci-lint run ./...

test: test-unit

test-unit:
	go test ./... -race -coverprofile=coverage.out

test-integration:
	TEST_DATABASE_DSN=$${TEST_DATABASE_DSN:-postgres://postgres:postgres@localhost:5432/credit_service?sslmode=disable} 	RUN_INTEGRATION_TESTS=true 	go test ./... -tags=integration -count=1

bench:
	go test ./... -run=^$$ -bench=. -benchmem

build:
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) ./cmd/api

run:
	go run ./cmd/api

docker-build:
	docker build -t $(APP_NAME):local .
