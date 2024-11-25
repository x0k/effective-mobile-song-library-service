#!/bin/env bash
set -xe

d:
  go run cmd/app/main.go

db:
  sqlc generate

migration:
  migrate create -ext sql -dir db/migrations -seq $1

client:
  mkdir -p lib/music_info
  oapi-codegen --config=oapi-codegen.yaml api/music-info.yaml > internal/lib/music_info/client.go

t:
  go test ./...

lint:
  golangci-lint run ./...
