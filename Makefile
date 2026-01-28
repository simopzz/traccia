.PHONY: help all build run test clean watch tailwind-install docker-run docker-down itest templ-install lint fmt vet tidy verify

DEFAULT_GOAL := help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: build test ## Build & test the application

fmt: ## Format the code
	@echo "Formatting..."
	@go fmt ./...

vet: ## Vet the code
	@echo "Vetting..."
	@go vet ./...

tidy: ## Tidy the Go modules
	@echo "Tidying modules..."
	@go mod tidy

verify: ## Verify the Go modules
	@echo "Verifying modules..."
	@go mod verify

lint: fmt vet tidy verify ## Lint the code (fmt, vet, tidy, verify)

templ-install:
	@if ! command -v templ > /dev/null; then \
		read -p "Go's 'templ' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/a-h/templ/cmd/templ@latest; \
			if [ ! -x "$$(command -v templ)" ]; then \
				echo "templ installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install templ. Exiting..."; \
			exit 1; \
		fi; \
	fi

tailwind-install:
	@if [ ! -f tailwindcss ]; then curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o tailwindcss; fi
	
	@chmod +x tailwindcss

build: tailwind-install templ-install ## Build the application
	@echo "Building..."
	@templ generate
	@./tailwindcss -i web/assets/css/input.css -o web/assets/css/output.css
	@go build -o main cmd/api/main.go

run: ## Run the application
	@go run cmd/api/main.go

docker-run: ## Create DB container
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

docker-down: ## Shutdown DB container
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

test: ## Test the application
	@echo "Testing..."
	@go test ./... -v

itest: ## Integrations Tests for the application
	@echo "Running integration tests..."
	@go test ./internal/database -v

clean: ## Clean the binary
	@echo "Cleaning..."
	@rm -f main

watch: ## Live Reload
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi
