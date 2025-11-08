# AI Service

This is the AI service, which provides gRPC services for vector operations and model inference.

## Features

- Vector database operations
- Model management
- Text generation and embedding generation
- gRPC services
- FastAPI HTTP API

## Prerequisites

- Python 3.9 or later
- Docker and Docker Compose
- Access to vector database (Milvus)

## Getting Started

### Installing Dependencies

```bash
# Install dependencies
pip install -r requirements.txt
```

### Generating gRPC Code

```bash
# Generate gRPC code
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
- `VECTOR_SERVICE_PORT`: Port for the vector service (default: 50051)
- `MODEL_SERVICE_PORT`: Port for the model service (default: 50052)
- `VECTOR_DB_TYPE`: Type of vector database (milvus, weaviate, etc.)
- `VECTOR_DB_HOST`: Host of vector database
- `VECTOR_DB_PORT`: Port of vector database

## Architecture

The AI service is built using the following technologies:

- **FastAPI**: HTTP web framework
- **gRPC**: For service communication
- **Milvus**: For vector database operations
- **Transformers**: For model inference

## Services

### Vector Service

The vector service provides the following gRPC methods:

- `HealthCheck`: Check the health of the service
- `CreateCollection`: Create a collection
- `DropCollection`: Drop a collection
- `ListCollections`: List collections
- `HasCollection`: Check if a collection exists
- `CreateIndex`: Create an index
- `DropIndex`: Drop an index
- `Search`: Search vectors
- `Insert`: Insert vectors
- `Upsert`: Upsert vectors
- `Delete`: Delete vectors

### Model Service

The model service provides the following gRPC methods:

- `HealthCheck`: Check the health of the service
- `ListModels`: List available models
- `LoadModel`: Load a model
- `UnloadModel`: Unload a model
- `GetModelStatus`: Get the status of a model
- `GenerateText`: Generate text
- `GenerateEmbedding`: Generate embeddings

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
make build
```

### Running Docker Container

```bash
# Run Docker container
make docker-run
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

This project is licensed under the MIT License.