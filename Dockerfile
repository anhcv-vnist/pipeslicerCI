FROM golang:1.19-alpine AS builder

WORKDIR /app

# Install build dependencies for SQLite3
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 go build -o /app/bin/web cmd/web/main.go

# Use a smaller image for the final container
FROM alpine:3.17

WORKDIR /app

# Install runtime dependencies for SQLite3 and Docker CLI
RUN apk add --no-cache libc6-compat docker-cli

# Copy the binary from the builder stage
COPY --from=builder /app/bin/web /app/bin/web

# Copy necessary files
COPY internal/app/web/docs/swagger-ui /app/internal/app/web/docs/swagger-ui
COPY internal/app/web/docs/swagger.yaml /app/internal/app/web/docs/swagger.yaml

# Create directories for data and logs
RUN mkdir -p /data /logs

# Expose the port
EXPOSE 3000

# Set environment variables
ENV SERVER_PORT=3000
ENV SERVER_HOST=0.0.0.0
ENV DATABASE_PATH=/data/pipeslicerci.db
ENV REGISTRY_DEFAULT_TYPE=local
ENV REGISTRY_LOCAL_PATH=/data/registry
ENV LOGGING_LEVEL=debug
ENV LOGGING_FILE=/logs/pipeslicerci.log
ENV WORKDIR_PATH=/data/workdir

# Run the application
CMD ["/app/bin/web"]
