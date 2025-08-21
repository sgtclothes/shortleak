# Makefile for Docker operations
.PHONY: help up down build rebuild logs clean dev prod migrate seed test

# Default target
help:
	@echo "Available commands:"
	@echo "  up         - Start all services in detached mode"
	@echo "  down       - Stop and remove all containers"
	@echo "  build      - Build all services"
	@echo "  rebuild    - Rebuild all services from scratch"
	@echo "  logs       - Show logs from all services"
	@echo "  clean      - Remove all containers, images, and volumes"
	@echo "  dev        - Start development environment"
	@echo "  prod       - Start production environment"
	@echo "  migrate    - Run database migrations"
	@echo "  seed       - Run database seeding"
	@echo "  test       - Run tests in backend container"
	@echo "  pgadmin    - Start with pgAdmin tool"

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# Build all services
build:
	docker-compose build

# Rebuild all services from scratch
rebuild:
	docker-compose build --no-cache

# Show logs
logs:
	docker-compose logs -f

# Clean everything
clean:
	docker-compose down -v --rmi all --remove-orphans
	docker system prune -a --volumes -f

# Development environment
dev:
	docker-compose up --build

# Production environment
prod:
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Run migrations
migrate:
	docker-compose exec backend go run ./cmd/migrate/main.go migrate

# Run seeds
seed:
	docker-compose exec backend go run ./cmd/seed/main.go

# Run tests
test:
	docker-compose exec backend go test ./... -coverprofile=coverage && go tool cover -html=coverage

# Start with pgAdmin
pgadmin:
	docker-compose --profile tools up -d