# Default recipe - show help
default:
    @just --list

# Build and test
all: build test

# Hot reload via air
dev:
    air

# Build binary to bin/app
build:
    go build -o bin/app ./cmd/app

# Run the application
run: build
    ./bin/app

# Run tests with race detection
test:
    go test -v -race ./...

# Auto-fix formatting and imports
fmt:
    gofmt -w .
    goimports -w -local github.com/simopzz/traccia .

# Run formatter then golangci-lint
lint: fmt
    golangci-lint run

# Run sqlc and templ code generation
generate:
    sqlc generate
    templ generate

# Tidy and verify go modules
tidy:
    go mod tidy
    go mod verify

# Install required tools
tools:
    @command -v air >/dev/null || go install github.com/air-verse/air@latest
    @command -v sqlc >/dev/null || go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    @command -v templ >/dev/null || go install github.com/a-h/templ/cmd/templ@latest
    @command -v golangci-lint >/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    @command -v goimports >/dev/null || go install golang.org/x/tools/cmd/goimports@latest
    @command -v migrate >/dev/null || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Start PostgreSQL container
docker-up:
    docker compose up -d

# Stop PostgreSQL container
docker-down:
    docker compose down

# Apply database migrations
migrate-up:
    migrate -path migrations -database "postgres://traccia:traccia@localhost:5432/traccia?sslmode=disable" up

# Rollback database migrations
migrate-down:
    migrate -path migrations -database "postgres://traccia:traccia@localhost:5432/traccia?sslmode=disable" down

# Remove build artifacts
clean:
    rm -rf bin/ tmp/
