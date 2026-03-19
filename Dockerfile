# syntax=docker/dockerfile:1.7
FROM golang:1.23.2-alpine3.20 AS builder

WORKDIR /src
RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/credit-decision-service ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/credit-decision-service /app/credit-decision-service
EXPOSE 8080
ENTRYPOINT ["/app/credit-decision-service"]
