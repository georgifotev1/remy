SHELL := /bin/bash

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)
COMPOSE_DEV := docker compose -f deploy/compose.dev.yaml

-include .env
APP_ENV ?= development
DB_URL ?= postgres://remy:remy@localhost:5432/remy?sslmode=disable
export

.PHONY: help up down migrate migrate-down web-install web-build run dev test test-int lint build tidy

help:
	@printf '%s\n' \
		'up          start local Postgres' \
		'down        stop local Postgres (data is kept)' \
		'migrate     apply goose migrations' \
		'migrate-down roll back the last migration' \
		'web-install install web dependencies' \
		'web-build   build the Vue app into web/dist' \
		'run         build web and run the API on :8080' \
		'dev         run the API and Vite dev server together' \
		'test        unit tests with the race detector' \
		'test-int    integration tests (build tag integration)' \
		'lint        run golangci-lint' \
		'build       build all commands into bin/' \
		'tidy        go mod tidy'

up:
	$(COMPOSE_DEV) up -d --wait

down:
	$(COMPOSE_DEV) down

migrate:
	go run ./cmd/migrate

migrate-down:
	go run ./cmd/migrate down

web-install:
	pnpm -C web install --frozen-lockfile

web-build: web-install
	pnpm -C web build

run: web-build
	go run ./cmd/api

dev:
	@trap 'kill 0' INT TERM EXIT; \
	( go run ./cmd/api ) & \
	pnpm -C web dev

test:
	go test -race -shuffle=on ./...

test-int:
	go test -race -shuffle=on -tags=integration ./...

lint:
	golangci-lint run

build: web-build
	CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS)" -o bin/ ./cmd/...

tidy:
	go mod tidy
