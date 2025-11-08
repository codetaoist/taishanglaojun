FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# -----------------------------
# Configure Go module proxy to mitigate network instability
# and implement retry logic for dependency download
# -----------------------------
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

# Leverage Docker BuildKit cache mounts for faster rebuilds
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    /bin/sh -c 'go env -w GOPROXY=${GOPROXY} GOSUMDB=off && \
    for i in 1 2 3; do \
        echo "Attempt $i: downloading modules" && \
        go mod download && break || sleep 20; \
    done'

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migrations directory
COPY --from=builder /app/migrations ./migrations

# Expose port
EXPOSE 8082

# Run the binary
CMD ["./main"]