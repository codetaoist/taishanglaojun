# AI Service API

This is the API service for the AI Service, which provides HTTP endpoints for interacting with the AI services.

## Features

- HTTP API for AI services
- gRPC client for communicating with AI services
- Vector database operations
- Model management
- Text generation and embedding generation

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- Access to the AI services

## Getting Started

### Building the Application

```bash
# Install dependencies
make deps

# Build the application
make build

# Generate gRPC client code
make generate-grpc
```

### Running the Application

```bash
# Run the application locally
make run

# Run with Docker Compose
docker-compose up
```

### Environment Variables

- `ENV`: Environment (dev, staging, production)
- `LOG_LEVEL`: Log level (debug, info, warn, error)
- `LAOJUN_API_PORT`: Port for the API service (default: 8082)
- `DATABASE_URL`: PostgreSQL database URL
- `AUTH_SERVICE_URL`: URL of the authentication service
- `VECTOR_DB_TYPE`: Type of vector database (milvus, weaviate, etc.)
- `VECTOR_DB_HOST`: Host of vector database
- `VECTOR_DB_PORT`: Port of vector database
- `AI_VECTOR_SERVICE_ADDR`: Address of the AI vector service (default: localhost:50051)
- `AI_MODEL_SERVICE_ADDR`: Address of the AI model service (default: localhost:50052)

## API Endpoints

### Health Check

- `GET /health` - Health check endpoint

### AI Service Endpoints

All AI service endpoints are under `/api/v1/ai` and require authentication.

#### Vector Operations

- `POST /api/v1/ai/collections` - Create a collection
- `DELETE /api/v1/ai/collections/:name` - Delete a collection
- `GET /api/v1/ai/collections` - List collections
- `POST /api/v1/ai/search` - Search vectors
- `POST /api/v1/ai/insert` - Insert vectors

#### Model Operations

- `GET /api/v1/ai/models` - List models
- `POST /api/v1/ai/models/:name/load` - Load a model
- `GET /api/v1/ai/models/:name/status` - Get model status

#### Inference Operations

- `POST /api/v1/ai/generate` - Generate text
- `POST /api/v1/ai/embed` - Generate embeddings

## Development

### Running Tests

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage
```

### Linting and Formatting

```bash
# Lint code
make lint

# Format code
make fmt
```

## Docker

### Building Docker Image

```bash
# Build Docker image
make docker-build
```

### Running Docker Container

```bash
# Run Docker container
make docker-run
```

## Architecture

The API service is built using the following technologies:

- **Gin**: HTTP web framework
- **gRPC**: For communication with AI services
- **PostgreSQL**: For data persistence
- **Milvus**: For vector database operations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

This project is licensed under the MIT License.