.PHONY: build-cli build-worker migrate up down clean test

DATABASE_URL ?= postgres://beacon:beacon123@localhost:5432/beacon?sslmode=disable

build-cli:
	go build -o bin/beacon cmd/cli/main.go

build-worker:
	go build -o bin/worker cmd/worker/main.go

build: build-cli build-worker

migrate:
	docker exec -i beacon-postgres-1 psql -U beacon -d beacon < migrations/001_initial.sql

up:
	docker compose up -d

down:
	docker compose down

clean:
	docker compose down -v
	rm -rf bin/

test:
	go test ./...

run-cli: build-cli
	./bin/beacon --db "$(DATABASE_URL)"

run-worker: build-worker
	DATABASE_URL="$(DATABASE_URL)" TEMPORAL_HOST=localhost:7233 ./bin/worker