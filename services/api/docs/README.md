# Taoist API Documentation

This directory contains the OpenAPI 3.0 specification for the Taoist platform API.

## Overview

The Taoist platform is divided into two main domains:

1. **Laojun Domain** - Focuses on plugin management, configuration management, and audit logging
2. **Taishang Domain** - Focuses on AI model management, vector collections, and task management

## API Documentation

The complete API specification is available in [openapi.yaml](./openapi.yaml).

### Viewing the Documentation

You can view the interactive API documentation using any of the following methods:

#### 1. Swagger UI (Recommended)

```bash
# Using Docker
docker run -p 80:8080 -e SWAGGER_JSON=/openapi.yaml -v $(pwd):/openapi swaggerapi/swagger-ui

# Or using Node.js
npx swagger-ui-server -p 8080 openapi.yaml
```

Then open http://localhost:8080 in your browser.

#### 2. Redoc

```bash
# Using Docker
docker run -p 80:8080 -e SPEC_URL=/openapi.yaml -v $(pwd):/openapi redocly/redoc

# Or using Node.js
npx redoc-cli serve openapi.yaml
```

Then open http://localhost:8080 in your browser.

#### 3. VS Code Extension

Install the "OpenAPI (Swagger) Editor" extension in VS Code and open the `openapi.yaml` file.

## API Endpoints

### Laojun Domain

#### Health Check
- `GET /api/laojun/health` - Check Laojun domain health

#### Configuration Management
- `GET /api/laojun/config` - Get configuration
- `PUT /api/laojun/config` - Update configuration

#### Plugin Management
- `GET /api/laojun/plugins` - List plugins
- `POST /api/laojun/plugins` - Install plugin
- `POST /api/laojun/plugins/{id}/start` - Start plugin
- `POST /api/laojun/plugins/{id}/stop` - Stop plugin
- `POST /api/laojun/plugins/{id}/upgrade` - Upgrade plugin
- `DELETE /api/laojun/plugins/{id}` - Uninstall plugin

#### Audit Logs
- `GET /api/laojun/audit-logs` - List audit logs

### Taishang Domain

#### Health Check
- `GET /api/taishang/health` - Check Taishang domain health

#### Model Management
- `GET /api/taishang/models` - List models
- `POST /api/taishang/models` - Register model
- `GET /api/taishang/models/{id}` - Get model
- `PUT /api/taishang/models/{id}` - Update model
- `DELETE /api/taishang/models/{id}` - Delete model

#### Vector Collection Management
- `GET /api/taishang/collections` - List collections
- `POST /api/taishang/collections` - Create collection
- `GET /api/taishang/collections/{id}` - Get collection
- `DELETE /api/taishang/collections/{id}` - Delete collection
- `POST /api/taishang/collections/{id}/rebuild-index` - Rebuild collection index

#### Task Management
- `GET /api/taishang/tasks` - List tasks
- `POST /api/taishang/tasks` - Create task
- `GET /api/taishang/tasks/{id}` - Get task
- `PUT /api/taishang/tasks/{id}` - Update task
- `DELETE /api/taishang/tasks/{id}` - Delete task

## Authentication

All protected endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Response Format

All API responses follow a consistent format:

```json
{
  "code": 200,
  "message": "success",
  "traceId": "abc123",
  "data": {
    // Response data here
  }
}
```

## Error Handling

The API returns standard HTTP status codes along with detailed error messages:

- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

## SDK Generation

You can generate client SDKs in various languages using the OpenAPI specification:

### Using OpenAPI Generator

```bash
# Install OpenAPI Generator
npm install @openapitools/openapi-generator-cli -g

# Generate a TypeScript client
openapi-generator-cli generate -i openapi.yaml -g typescript-axios -o ./client/ts

# Generate a Python client
openapi-generator-cli generate -i openapi.yaml -g python -o ./client/python

# Generate a Go client
openapi-generator-cli generate -i openapi.yaml -g go -o ./client/go
```

### Using Swagger Codegen

```bash
# Install Swagger Codegen
npm install swagger-codegen -g

# Generate a JavaScript client
swagger-codegen generate -i openapi.yaml -l javascript -o ./client/js
```

## Contributing

When making changes to the API:

1. Update the `openapi.yaml` file with your changes
2. Validate the specification using an online tool or CLI
3. Update this README if necessary
4. Consider regenerating client SDKs if needed

## Validation

You can validate the OpenAPI specification using:

```bash
# Using the OpenAPI CLI
npx @apidevtools/swagger-parser validate openapi.yaml

# Using Redoc CLI
npx redoc-cli lint openapi.yaml
```