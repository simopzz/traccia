.PHONY: dev build test lint generate docker-up docker-down migrate-up migrate-down clean

dev:
	air

build:
	go build -o bin/app ./cmd/app

test:
	go test -v -race ./...

lint:
	golangci-lint run

generate:
	sqlc generate
	templ generate

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "postgres://traccia:traccia@localhost:5432/traccia?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://traccia:traccia@localhost:5432/traccia?sslmode=disable" down

clean:
	rm -rf bin/ tmp/
