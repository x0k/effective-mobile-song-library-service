#!/bin/env bash
set -xe

d:
  go run cmd/app/main.go

migration:
  migrate create -ext sql -dir migrations -seq $1

client:
  mkdir -p lib/music_info
  oapi-codegen --config=oapi-codegen.yml api/music-info.yml > internal/lib/music_info/client.go

t:
  go test ./...

lint:
  golangci-lint run ./...

sql:
  psql -h localhost -p $1 -U test -d songs
