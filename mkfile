#!/bin/env bash
set -xe

d:
  go run main.go

db:
  sqlc generate

migration:
  migrate create -ext sql -dir db/migrations -seq $1
