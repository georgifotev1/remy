MIGRATIONS_DIR := migrations
DATABASE_URL ?= postgres://remy:remy@localhost:5432/remy?sslmode=disable

.PHONY: build dev migrate-up migrate-down migrate-create

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-create:
	@test -n "$(name)" || (echo "usage: make migrate-create name=<name>" && exit 1)
	goose -dir $(MIGRATIONS_DIR) -s create $(name) sql

build:
	go build -o bin/api ./cmd/api

dev: build
	./bin/api
