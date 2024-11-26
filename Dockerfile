FROM golang:1.23.3-alpine3.20 AS builder

RUN apk add --no-cache build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o app ./cmd/app/main.go

FROM alpine:3.20.0

WORKDIR /app

COPY --from=builder /app/app .
COPY ./migrations ./migrations
ENV PG_MIGRATIONS_URI=file://migrations

ENTRYPOINT [ "./app" ]
