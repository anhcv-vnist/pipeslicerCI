# PipeslicerCI

PipeslicerCI is a custom CI/CD tool designed specifically for microservices architectures. It provides a comprehensive set of features to address common challenges in microservices development, testing, and deployment.

## Features

### 1. Image Builder Service

The Image Builder Service automates the process of building Docker images from source code and pushing them to a registry. This eliminates the need to build images directly in containers during deployment, resulting in faster startup times and more efficient resource usage.

**Key capabilities:**
- Build Docker images from Git repositories
- Push images to Docker registries
- Detect which services have changed and need rebuilding
- Support for multi-stage builds to optimize image size

### 2. Registry Manager

The Registry Manager provides a centralized system for managing Docker image metadata. It keeps track of all images, their versions, and associated metadata, making it easy to find the right image for deployment and to roll back to previous versions if needed.

**Key capabilities:**
- Store metadata about built images
- Query images by service, tag, or commit
- Track image history
- Support for tagging images

### 3. Registry Connector

The Registry Connector provides a unified interface for interacting with different Docker registries, including Docker Hub and Harbor. It handles authentication, pushing images, and managing repositories and tags.

**Key capabilities:**
- Connect to Docker Hub, Harbor, or any Docker-compatible registry
- Authenticate with registries
- Push images to registries
- List repositories and tags
- Delete tags

### 3. Configuration Manager

The Configuration Manager provides centralized configuration management for all services. It allows you to store and manage configuration values in a central location, making it easy to maintain consistent configuration across environments and to update configuration values without rebuilding images.

**Key capabilities:**
- Store configuration values for different services and environments
- Securely manage sensitive information
- Generate .env files for services
- Import and export configuration in JSON format

### 4. Dependency Analyzer

The Dependency Analyzer helps you understand the dependencies between services, making it easier to determine the correct order for deployment and to understand the impact of changes.

**Key capabilities:**
- Analyze service dependencies
- Generate dependency graphs
- Recommend deployment order
- Identify circular dependencies

## Getting Started

### Prerequisites

- Go 1.16 or later
- Docker
- Git

### Installation

1. Clone the repository:

```bash
git clone https://github.com/vanhcao3/pipeslicerCI.git
cd pipeslicerCI
```

2. Build the application:

```bash
go build -o pipeslicerci ./cmd/web
```

3. Run the application:

```bash
./pipeslicerci
```

The server will start on port 3000 by default.

## API Documentation

PipeslicerCI provides a comprehensive API documentation using Swagger/OpenAPI. You can access the API documentation by navigating to the `/swagger` endpoint in your browser after starting the server.

For example, if the server is running on localhost port 3000, you can access the API documentation at:

```
http://localhost:3000/swagger
```

The API documentation provides detailed information about all available endpoints, including:

- Request parameters
- Request body schemas
- Response schemas
- Example requests and responses

### Available APIs

PipeslicerCI exposes the following API groups:

1. **Image Builder API** - For building and pushing Docker images
2. **Registry API** - For managing Docker image metadata
3. **Registry Connector API** - For interacting with Docker registries
4. **Configuration API** - For managing configuration values
5. **Pipeline API** - For executing pipelines

## Usage Examples

### Building a Docker Image

```bash
curl -X POST http://localhost:3000/imagebuilder/build \
  -F "url=https://github.com/username/repo.git" \
  -F "branch=main" \
  -F "servicePath=micro-services/auth-service" \
  -F "tag=v1.0.0" \
  -F "registry=docker.io" \
  -F "username=myusername" \
  -F "password=mypassword"
```

### Pushing an Image to Docker Hub

```bash
curl -X POST http://localhost:3000/registry-connector/push \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://registry.hub.docker.com",
    "username": "myusername",
    "password": "mypassword",
    "imageName": "myusername/myimage"
  }'
```

### Pushing an Image to Harbor

```bash
curl -X POST http://localhost:3000/registry-connector/push \
  -H "Content-Type: application/json" \
  -d '{
    "type": "harbor",
    "url": "https://harbor.example.com",
    "username": "myusername",
    "password": "mypassword",
    "imageName": "myproject/myimage"
  }'
```

### Setting a Configuration Value

```bash
curl -X POST http://localhost:3000/config/services/auth-service/environments/development/values \
  -H "Content-Type: application/json" \
  -d '{
    "key": "DB_HOST",
    "value": "localhost",
    "isSecret": false
  }'
```

### Generating a .env File

```bash
curl -X GET http://localhost:3000/config/services/auth-service/environments/development/env \
  -H "Authorization: Bearer mytoken"
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
