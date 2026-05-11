.PHONY: run build migrate-up migrate-down seed reindex sqlc test lint docker-up docker-down

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

seed:
	go run ./scripts/seed

reindex:
	go run ./scripts/reindex

sqlc:
	sqlc generate

test:
	go test ./...

lint:
	golangci-lint run

docker-up:
	docker compose up -d

docker-down:
	docker compose down
