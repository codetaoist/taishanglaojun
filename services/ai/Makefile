# Makefile for the AI service

# Variables
APP_NAME = ai-service
VERSION = $(shell git describe --tags --always --dirty)

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	docker build -t $(APP_NAME):$(VERSION) .

# Run the application
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	python grpc_server.py

# Generate gRPC code
.PHONY: generate-grpc
generate-grpc:
	@echo "Generating gRPC code..."
	python -m grpc_tools.protoc --python_out=. --grpc_python_out=. --proto_path=app/proto app/proto/ai_service.proto

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	pip install -r requirements.txt

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf __pycache__/
	rm -rf app/proto/*_pb2.py
	rm -rf app/proto/*_pb2_grpc.py

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	python -m pytest

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	python -m pytest --cov=app --cov-report=html

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	flake8 app/
	black --check app/

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	black app/

# Run Docker container
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 50051:50051 -p 50052:50052 -p 8000:8000 $(APP_NAME):$(VERSION)

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Build the application (default)"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  generate-grpc - Generate gRPC code"
	@echo "  deps          - Install dependencies"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Lint code"
	@echo "  fmt           - Format code"
	@echo "  docker-run    - Run Docker container"
	@echo "  help          - Show this help message"