# API Service Tests

This directory contains all tests for the API service.

## Directory Structure

- `integration/` - Integration tests for API endpoints
- `unit/` - Unit tests for individual components and functions

## Running Tests

```bash
# Run all tests
go test ./tests/...

# Run only integration tests
go test ./tests/integration/...

# Run only unit tests
go test ./tests/unit/...
```

## Test Coverage

To generate test coverage report:

```bash
go test -cover ./tests/...
```

## Test Data

Test data should be placed in the `testdata/` directory within each test subdirectory.