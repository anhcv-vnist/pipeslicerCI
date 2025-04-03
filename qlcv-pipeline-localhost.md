# CI/CD Pipeline for QLCV Project Using PipeslicerCI (Localhost Deployment)

## 1. Proposed CI/CD Pipeline Stages

The proposed CI/CD pipeline for the QLCV project consists of the following stages:

| Stage | Description |
|-------|-------------|
| 1. Code Checkout | Retrieve source code from the repository |
| 2. Testing | Run automated tests to ensure code quality |
| 3. Building | Build Docker images for all microservices |
| 4. Registry Management | Store and manage Docker images |
| 5. Local Deployment | Deploy to localhost for testing |
| 6. Monitoring | Monitor deployment and verify functionality |

## 2. PipeslicerCI Components and Their Functions

PipeslicerCI provides several key components that are used in different stages of the CI/CD pipeline:

| Component | Description |
|-----------|-------------|
| Image Builder | Builds Docker images from source code |
| Registry Manager | Manages metadata about Docker images |
| Registry Connector | Connects to Docker registries |
| Config Manager | Manages configuration for deployments |
| Pipeline Executor | Orchestrates the CI/CD pipeline |

## 3. Detailed Pipeline Stages

### 3.1. Code Checkout Stage

**PipeslicerCI Functionality**: Pipeline Executor

**Inputs**:
- Git repository URL
- Branch name
- Credentials (if needed)

**Outputs**:
- Local copy of the source code

**Problems Solved**:
- Automates the process of retrieving the latest code
- Ensures consistent code checkout across all pipeline runs
- Handles authentication with Git repositories

### 3.2. Testing Stage

**PipeslicerCI Functionality**: Pipeline Executor

**Inputs**:
- Source code from Code Checkout stage
- Test configuration

**Outputs**:
- Test results
- Test coverage reports

**Problems Solved**:
- Ensures code quality before proceeding to build
- Provides consistent test environment
- Fails fast if tests don't pass, saving resources

### 3.3. Building Stage

**PipeslicerCI Functionality**: Image Builder

**Inputs**:
- Source code from Code Checkout stage
- Dockerfile for each service
- Build configuration

**Outputs**:
- Docker images for each microservice
- Build logs

**Problems Solved**:
- Automates the build process for all microservices
- Ensures consistent build environment
- Handles multi-stage builds efficiently
- Manages build dependencies

### 3.4. Registry Management Stage

**PipeslicerCI Functionality**: Registry Manager, Registry Connector

**Inputs**:
- Docker images from Building stage
- Registry configuration

**Outputs**:
- Stored Docker images in registry
- Image metadata in Registry Manager

**Problems Solved**:
- Manages Docker image metadata (service, tag, commit, branch, build time)
- Provides versioning for Docker images
- Enables rollback to previous versions if needed
- Tracks the history of builds

### 3.5. Local Deployment Stage

**PipeslicerCI Functionality**: Config Manager, Pipeline Executor

**Inputs**:
- Docker images from Registry Management stage
- Deployment configuration

**Outputs**:
- Deployed application on localhost
- Deployment logs

**Problems Solved**:
- Automates the deployment process
- Manages environment-specific configurations
- Handles service dependencies
- Ensures services are deployed in the correct order

### 3.6. Monitoring Stage

**PipeslicerCI Functionality**: Pipeline Executor

**Inputs**:
- Deployed application from Local Deployment stage
- Monitoring configuration

**Outputs**:
- Health check results
- Monitoring logs

**Problems Solved**:
- Verifies successful deployment
- Checks application health
- Provides feedback on deployment status
- Enables automated rollback if needed

## 4. PipeslicerCI Installation for Localhost Deployment

```bash
# Clone PipeslicerCI repository
git clone https://github.com/vanhcao3/pipeslicerCI.git
cd pipeslicerCI

# Install Go dependencies
go mod download

# Build PipeslicerCI
make build

# Create necessary directories
mkdir -p ~/.pipeslicerci/data
mkdir -p ~/.pipeslicerci/logs

# Create configuration file
cat > ~/.pipeslicerci/config.yaml << EOF
server:
  port: 3000
  host: "0.0.0.0"

database:
  path: "~/.pipeslicerci/data/pipeslicerci.db"

registry:
  default_type: "local"
  local:
    path: "~/.pipeslicerci/data/registry"

logging:
  level: "info"
  file: "~/.pipeslicerci/logs/pipeslicerci.log"

workdir:
  path: "~/.pipeslicerci/data/workdir"
EOF

# Start PipeslicerCI
./build/pipeslicerci-web
```

## 5. Third-Party Dependencies

For localhost deployment, we need Docker and Docker Compose:

### 5.1. Docker Installation

```bash
# Update package index
sudo apt-get update

# Install prerequisites
sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common

# Add Docker's official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

# Add Docker repository
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"

# Update package index again
sudo apt-get update

# Install Docker CE
sudo apt-get install -y docker-ce

# Add current user to docker group
sudo usermod -aG docker $USER

# Apply group changes
newgrp docker
```

### 5.2. Docker Compose Installation

```bash
# Download Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

# Apply executable permissions
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker-compose --version
```

## 6. Detailed Script for Complete CI/CD Pipeline on Localhost

Create a file named `qlcv-pipeline.yaml` with the following content:

```yaml
version: "1.0"
name: "qlcv-pipeline-localhost"
description: "CI/CD Pipeline for QLCV project (localhost deployment)"

variables:
  GIT_REPO: "https://github.com/username/qlcv-refactor.git"
  BRANCH: "main"
  LOCAL_REGISTRY: "localhost:5000"

stages:
  - name: checkout
    steps:
      - name: clone-repository
        action: git-clone
        params:
          url: "${GIT_REPO}"
          branch: "${BRANCH}"
          path: "/tmp/qlcv-source"

  - name: test
    steps:
      - name: test-frontend
        action: run-tests
        params:
          workdir: "/tmp/qlcv-source/client-vite"
          command: "npm test"
      
      - name: test-auth-service
        action: run-tests
        params:
          workdir: "/tmp/qlcv-source/micro-services/auth-service"
          command: "npm test"
      
      - name: test-api-gateway
        action: run-tests
        params:
          workdir: "/tmp/qlcv-source/micro-services/api-gateway-service"
          command: "./mvnw test"

  - name: build
    steps:
      - name: build-frontend
        action: build-image
        params:
          workdir: "/tmp/qlcv-source/client-vite"
          dockerfile: "Dockerfile"
          tag: "${LOCAL_REGISTRY}/client-vite:latest"
      
      - name: build-auth-service
        action: build-image
        params:
          workdir: "/tmp/qlcv-source/micro-services/auth-service"
          dockerfile: "Dockerfile"
          tag: "${LOCAL_REGISTRY}/auth-service:latest"
      
      - name: build-api-gateway
        action: build-image
        params:
          workdir: "/tmp/qlcv-source/micro-services/api-gateway-service"
          dockerfile: "Dockerfile"
          tag: "${LOCAL_REGISTRY}/api-gateway-service:latest"
      
      - name: build-asset-service
        action: build-image
        params:
          workdir: "/tmp/qlcv-source/micro-services/asset-service"
          dockerfile: "Dockerfile"
          tag: "${LOCAL_REGISTRY}/asset-service:latest"
      
      - name: build-bidding-service
        action: build-image
        params:
          workdir: "/tmp/qlcv-source/micro-services/bidding-service"
          dockerfile: "Dockerfile"
          tag: "${LOCAL_REGISTRY}/bidding-service:latest"
      
      - name: build-dashboard-service
        action: build-image
        params:
          workdir: "/tmp/qlcv-source/micro-services/dashboard-service"
          dockerfile: "Dockerfile"
          tag: "${LOCAL_REGISTRY}/dashboard-service:latest"

  - name: registry
    steps:
      - name: push-frontend
        action: push-image
        params:
          image: "${LOCAL_REGISTRY}/client-vite:latest"
          registry: "${LOCAL_REGISTRY}"
      
      - name: push-auth-service
        action: push-image
        params:
          image: "${LOCAL_REGISTRY}/auth-service:latest"
          registry: "${LOCAL_REGISTRY}"
      
      - name: push-api-gateway
        action: push-image
        params:
          image: "${LOCAL_REGISTRY}/api-gateway-service:latest"
          registry: "${LOCAL_REGISTRY}"
      
      - name: push-asset-service
        action: push-image
        params:
          image: "${LOCAL_REGISTRY}/asset-service:latest"
          registry: "${LOCAL_REGISTRY}"
      
      - name: push-bidding-service
        action: push-image
        params:
          image: "${LOCAL_REGISTRY}/bidding-service:latest"
          registry: "${LOCAL_REGISTRY}"
      
      - name: push-dashboard-service
        action: push-image
        params:
          image: "${LOCAL_REGISTRY}/dashboard-service:latest"
          registry: "${LOCAL_REGISTRY}"
      
      - name: record-metadata
        action: record-metadata
        params:
          services:
            - name: "client-vite"
              tag: "latest"
              commit: "${GIT_COMMIT}"
              branch: "${BRANCH}"
            - name: "auth-service"
              tag: "latest"
              commit: "${GIT_COMMIT}"
              branch: "${BRANCH}"
            - name: "api-gateway-service"
              tag: "latest"
              commit: "${GIT_COMMIT}"
              branch: "${BRANCH}"
            - name: "asset-service"
              tag: "latest"
              commit: "${GIT_COMMIT}"
              branch: "${BRANCH}"
            - name: "bidding-service"
              tag: "latest"
              commit: "${GIT_COMMIT}"
              branch: "${BRANCH}"
            - name: "dashboard-service"
              tag: "latest"
              commit: "${GIT_COMMIT}"
              branch: "${BRANCH}"

  - name: deploy
    steps:
      - name: generate-compose-file
        action: generate-config
        params:
          template: "/tmp/qlcv-source/docker-compose.yml"
          output: "/tmp/qlcv-docker-compose.yml"
          variables:
            REGISTRY: "${LOCAL_REGISTRY}"
            TAG: "latest"
      
      - name: deploy-to-localhost
        action: deploy-compose
        params:
          compose_file: "/tmp/qlcv-docker-compose.yml"
          project_name: "qlcv"

  - name: monitor
    steps:
      - name: health-check
        action: health-check
        params:
          urls:
            - "http://localhost:80"  # Frontend
            - "http://localhost:3000"  # API Gateway
          timeout: 60
          interval: 5
      
      - name: log-deployment-status
        action: log-status
        params:
          message: "QLCV deployment to localhost completed"
          status: "success"
```

## 7. Step-by-Step Execution Script

Create a bash script named `run-qlcv-pipeline.sh` with the following content:

```bash
#!/bin/bash

# Set up local Docker registry if not already running
if ! docker ps | grep -q registry; then
  echo "Starting local Docker registry..."
  docker run -d -p 5000:5000 --name registry registry:2
else
  echo "Local Docker registry is already running."
fi

# Start PipeslicerCI if not already running
if ! pgrep -f pipeslicerci-web > /dev/null; then
  echo "Starting PipeslicerCI..."
  cd ~/pipeslicerCI
  ./build/pipeslicerci-web &
  sleep 5  # Wait for PipeslicerCI to start
else
  echo "PipeslicerCI is already running."
fi

# Execute the pipeline
echo "Executing QLCV pipeline..."
curl -X POST "http://localhost:3000/pipelines/build" \
  -F "file=@qlcv-pipeline.yaml"

# Monitor pipeline execution
echo "Monitoring pipeline execution..."
pipeline_id=$(curl -s "http://localhost:3000/pipelines/executions" | jq -r '.executions[0].id')
echo "Pipeline ID: $pipeline_id"

# Wait for pipeline to complete
status="running"
while [ "$status" = "running" ]; do
  sleep 10
  status=$(curl -s "http://localhost:3000/pipelines/executions/$pipeline_id" | jq -r '.status')
  echo "Pipeline status: $status"
done

if [ "$status" = "success" ]; then
  echo "Pipeline completed successfully!"
  echo "QLCV is now deployed to localhost."
  echo "Frontend: http://localhost:80"
  echo "API Gateway: http://localhost:3000"
else
  echo "Pipeline failed with status: $status"
  echo "Check logs for details: ~/.pipeslicerci/logs/pipeslicerci.log"
fi
```

Make the script executable:

```bash
chmod +x run-qlcv-pipeline.sh
```

## 8. Individual API Calls for Each Pipeline Stage

If you prefer to execute each stage manually or integrate with other tools, here are the individual API calls for each stage:

### 8.1. Start Code Checkout

```bash
curl -X POST "http://localhost:3000/pipelines/execute-stage" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_id": "qlcv-pipeline-localhost",
    "stage": "checkout",
    "params": {
      "GIT_REPO": "https://github.com/username/qlcv-refactor.git",
      "BRANCH": "main"
    }
  }'
```

### 8.2. Start Testing

```bash
curl -X POST "http://localhost:3000/pipelines/execute-stage" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_id": "qlcv-pipeline-localhost",
    "stage": "test",
    "params": {}
  }'
```

### 8.3. Start Building

```bash
curl -X POST "http://localhost:3000/pipelines/execute-stage" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_id": "qlcv-pipeline-localhost",
    "stage": "build",
    "params": {
      "LOCAL_REGISTRY": "localhost:5000"
    }
  }'
```

### 8.4. Start Registry Management

```bash
curl -X POST "http://localhost:3000/pipelines/execute-stage" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_id": "qlcv-pipeline-localhost",
    "stage": "registry",
    "params": {
      "LOCAL_REGISTRY": "localhost:5000",
      "GIT_COMMIT": "$(git rev-parse HEAD)"
    }
  }'
```

### 8.5. Start Local Deployment

```bash
curl -X POST "http://localhost:3000/pipelines/execute-stage" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_id": "qlcv-pipeline-localhost",
    "stage": "deploy",
    "params": {
      "LOCAL_REGISTRY": "localhost:5000"
    }
  }'
```

### 8.6. Start Monitoring

```bash
curl -X POST "http://localhost:3000/pipelines/execute-stage" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_id": "qlcv-pipeline-localhost",
    "stage": "monitor",
    "params": {}
  }'
```

## 9. Accessing the Deployed Application

After successful deployment, you can access the QLCV application at:

- Frontend: http://localhost:80
- API Gateway: http://localhost:3000
- Auth Service: http://localhost:3001
- Asset Service: http://localhost:3002
- Bidding Service: http://localhost:3003
- Dashboard Service: http://localhost:3004

## 10. Troubleshooting

### 10.1. Check PipeslicerCI Logs

```bash
tail -f ~/.pipeslicerci/logs/pipeslicerci.log
```

### 10.2. Check Docker Container Logs

```bash
docker-compose -f /tmp/qlcv-docker-compose.yml logs
```

### 10.3. Check Docker Container Status

```bash
docker-compose -f /tmp/qlcv-docker-compose.yml ps
```

### 10.4. Restart a Specific Service

```bash
docker-compose -f /tmp/qlcv-docker-compose.yml restart [service_name]
```

### 10.5. Rebuild and Redeploy a Specific Service

```bash
# Using PipeslicerCI API
curl -X POST "http://localhost:3000/imagebuilder/build" \
  -F "servicePath=/tmp/qlcv-source/micro-services/[service_name]" \
  -F "tag=latest" \
  -F "registry=localhost:5000"

# Redeploy
docker-compose -f /tmp/qlcv-docker-compose.yml up -d --no-deps --build [service_name]
```

## 11. Conclusion

This document has outlined a complete CI/CD pipeline for the QLCV project using PipeslicerCI, focusing on localhost deployment for testing purposes. The pipeline automates the entire process from code checkout to deployment and monitoring, making it easy to test changes and ensure quality before deploying to production environments.

By following the steps and scripts provided, you can quickly set up a local CI/CD pipeline for the QLCV project, enabling faster development cycles and more reliable deployments.
