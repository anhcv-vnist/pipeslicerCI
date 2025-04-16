#!/bin/bash
# QLCV CI/CD Pipeline Implementation Script
# This script implements the CI/CD pipeline for the QLCV project using PipeslicerCI
# It is designed for localhost deployment for testing purposes

set -e  # Exit on error

# Configuration
PIPESLICER_DIR="./"
QLCV_REPO="git@github.com:vanhcao3/pipeslicerCI.git"
QLCV_BRANCH="main"
LOCAL_REGISTRY="localhost:5000"
WORKDIR="/tmp/qlcv-pipeline"
CONFIG_DIR="$HOME/.pipeslicerci"
PIPESLICER_PORT=3000

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print section headers
print_header() {
    echo -e "\n${BLUE}==== $1 ====${NC}\n"
}

# Function to print success messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to print error messages
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Function to print info messages
print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"
    
    # Check if Docker is installed
    if command_exists docker; then
        print_success "Docker is installed"
    else
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check if curl is installed
    if command_exists curl; then
        print_success "curl is installed"
    else
        print_error "curl is not installed. Please install curl first."
        exit 1
    fi
    
    # Check if jq is installed
    if command_exists jq; then
        print_success "jq is installed"
    else
        print_error "jq is not installed. Please install jq first."
        exit 1
    fi
    
    # Check if Git is installed
    if command_exists git; then
        print_success "Git is installed"
    else
        print_error "Git is not installed. Please install Git first."
        exit 1
    fi
}

# Function to set up PipeslicerCI
setup_pipeslicerci() {
    print_header "Setting Up PipeslicerCI"
    
    # Create necessary directories
    mkdir -p "$CONFIG_DIR/data"
    mkdir -p "$CONFIG_DIR/logs"
    
    print_info "Creating Docker volumes for PipeslicerCI..."
    
    # Create Docker volumes if they don't exist
    if ! docker volume ls | grep -q "pipeslicerci-data"; then
        docker volume create pipeslicerci-data
        print_success "Created Docker volume: pipeslicerci-data"
    else
        print_info "Docker volume pipeslicerci-data already exists"
    fi
    
    if ! docker volume ls | grep -q "pipeslicerci-logs"; then
        docker volume create pipeslicerci-logs
        print_success "Created Docker volume: pipeslicerci-logs"
    else
        print_info "Docker volume pipeslicerci-logs already exists"
    fi
    
    # Check if PipeslicerCI Docker image exists
    if ! docker images | grep -q "pipeslicerci"; then
        print_info "Building PipeslicerCI Docker image..."
        cd "$PIPESLICER_DIR"
        docker build -t pipeslicerci:latest .
        print_success "Built PipeslicerCI Docker image"
    else
        print_info "PipeslicerCI Docker image already exists"
    fi
    
    # Check if PipeslicerCI container is already running
    if docker ps | grep -q "pipeslicerci"; then
        print_info "PipeslicerCI container is already running"
    else
        # Check if container exists but is stopped
        if docker ps -a | grep -q "pipeslicerci"; then
            print_info "PipeslicerCI container exists but is not running. Starting it..."
            docker start pipeslicerci
        else
            # Create and start new PipeslicerCI container
            print_info "Creating and starting new PipeslicerCI container..."
            docker run -d \
                --name pipeslicerci \
                -p $PIPESLICER_PORT:3000 \
                -v pipeslicerci-data:/data \
                -v pipeslicerci-logs:/logs \
                -v /var/run/docker.sock:/var/run/docker.sock \
                -e LOGGING_LEVEL=debug \
                pipeslicerci:latest
        fi
        
        sleep 5  # Wait for PipeslicerCI to start
        
        # Check if PipeslicerCI container started successfully
        if docker ps | grep -q "pipeslicerci"; then
            print_success "PipeslicerCI container started successfully"
            
            # Show container logs
            print_info "PipeslicerCI container logs:"
            docker logs --tail=20 pipeslicerci
        else
            print_error "Failed to start PipeslicerCI container"
            print_info "Container logs:"
            docker logs pipeslicerci
            exit 1
        fi
    fi
    
    # Test connection to PipeslicerCI API
    # print_info "Testing connection to PipeslicerCI API..."
    # if curl -s "http://127.0.0.1:$PIPESLICER_PORT/health" | grep -q "ok"; then
    #     print_success "PipeslicerCI API is accessible"
    # else
    #     print_error "PipeslicerCI API is not accessible"
    #     print_info "Container logs:"
    #     docker logs pipeslicerci
    #     exit 1
    # fi
}

# Function to set up local Docker registry
setup_registry() {
    print_header "Setting Up Local Docker Registry"
    
    # Check if registry is already running
    if docker ps | grep -q "registry"; then
        print_info "Local Docker registry is already running"
    else
        # Check if container exists but is stopped
        if docker ps -a | grep -q "registry"; then
            print_info "Registry container exists but is not running. Starting it..."
            docker start registry
        else
            # Start registry
            print_info "Creating and starting new registry container..."
            docker run -d -p 5000:5000 --name registry registry:2
        fi
        
        # Check if registry started successfully
        if docker ps | grep -q "registry"; then
            print_success "Local Docker registry started successfully"
        else
            print_error "Failed to start local Docker registry"
            print_info "Registry container logs:"
            docker logs registry
            exit 1
        fi
    fi
    
    # Test connection to registry
    print_info "Testing connection to registry..."
    if curl -s -f "http://localhost:5000/v2/" > /dev/null; then
        print_success "Registry API is accessible"
    else
        print_error "Registry API is not accessible"
        print_info "Registry container logs:"
        docker logs registry
        exit 1
    fi
}

# Function to create pipeline configuration
create_pipeline_config() {
    print_header "Creating Pipeline Configuration"
    
    mkdir -p "$WORKDIR"
    
    # Create pipeline configuration file
    cat > "$WORKDIR/qlcv-pipeline.yaml" << EOF
version: "1.0"
name: "qlcv-pipeline-localhost"
description: "CI/CD Pipeline for QLCV project (localhost deployment)"

variables:
  GIT_REPO: "$QLCV_REPO"
  BRANCH: "$QLCV_BRANCH"
  LOCAL_REGISTRY: "$LOCAL_REGISTRY"

stages:
  - name: checkout
    steps:
      - name: clone-repository
        action: git-clone
        params:
          url: "\${GIT_REPO}"
          branch: "\${BRANCH}"
          path: "$WORKDIR/source"

  - name: test
    steps:
      - name: test-frontend
        action: run-tests
        params:
          workdir: "$WORKDIR/source/client-vite"
          command: "npm test"
      
      - name: test-auth-service
        action: run-tests
        params:
          workdir: "$WORKDIR/source/micro-services/auth-service"
          command: "npm test"
      
      - name: test-api-gateway
        action: run-tests
        params:
          workdir: "$WORKDIR/source/micro-services/api-gateway-service"
          command: "./mvnw test"

  - name: build
    steps:
      - name: build-frontend
        action: build-image
        params:
          workdir: "$WORKDIR/source/client-vite"
          dockerfile: "Dockerfile"
          tag: "\${LOCAL_REGISTRY}/client-vite:latest"
      
      - name: build-auth-service
        action: build-image
        params:
          workdir: "$WORKDIR/source/micro-services/auth-service"
          dockerfile: "Dockerfile"
          tag: "\${LOCAL_REGISTRY}/auth-service:latest"
      
      - name: build-api-gateway
        action: build-image
        params:
          workdir: "$WORKDIR/source/micro-services/api-gateway-service"
          dockerfile: "Dockerfile"
          tag: "\${LOCAL_REGISTRY}/api-gateway-service:latest"
      
      - name: build-asset-service
        action: build-image
        params:
          workdir: "$WORKDIR/source/micro-services/asset-service"
          dockerfile: "Dockerfile"
          tag: "\${LOCAL_REGISTRY}/asset-service:latest"
      
      - name: build-bidding-service
        action: build-image
        params:
          workdir: "$WORKDIR/source/micro-services/bidding-service"
          dockerfile: "Dockerfile"
          tag: "\${LOCAL_REGISTRY}/bidding-service:latest"
      
      - name: build-dashboard-service
        action: build-image
        params:
          workdir: "$WORKDIR/source/micro-services/dashboard-service"
          dockerfile: "Dockerfile"
          tag: "\${LOCAL_REGISTRY}/dashboard-service:latest"

  - name: registry
    steps:
      - name: push-frontend
        action: push-image
        params:
          image: "\${LOCAL_REGISTRY}/client-vite:latest"
          registry: "\${LOCAL_REGISTRY}"
      
      - name: push-auth-service
        action: push-image
        params:
          image: "\${LOCAL_REGISTRY}/auth-service:latest"
          registry: "\${LOCAL_REGISTRY}"
      
      - name: push-api-gateway
        action: push-image
        params:
          image: "\${LOCAL_REGISTRY}/api-gateway-service:latest"
          registry: "\${LOCAL_REGISTRY}"
      
      - name: push-asset-service
        action: push-image
        params:
          image: "\${LOCAL_REGISTRY}/asset-service:latest"
          registry: "\${LOCAL_REGISTRY}"
      
      - name: push-bidding-service
        action: push-image
        params:
          image: "\${LOCAL_REGISTRY}/bidding-service:latest"
          registry: "\${LOCAL_REGISTRY}"
      
      - name: push-dashboard-service
        action: push-image
        params:
          image: "\${LOCAL_REGISTRY}/dashboard-service:latest"
          registry: "\${LOCAL_REGISTRY}"
      
      - name: record-metadata
        action: record-metadata
        params:
          services:
            - name: "client-vite"
              tag: "latest"
              commit: "\${GIT_COMMIT}"
              branch: "\${BRANCH}"
            - name: "auth-service"
              tag: "latest"
              commit: "\${GIT_COMMIT}"
              branch: "\${BRANCH}"
            - name: "api-gateway-service"
              tag: "latest"
              commit: "\${GIT_COMMIT}"
              branch: "\${BRANCH}"
            - name: "asset-service"
              tag: "latest"
              commit: "\${GIT_COMMIT}"
              branch: "\${BRANCH}"
            - name: "bidding-service"
              tag: "latest"
              commit: "\${GIT_COMMIT}"
              branch: "\${BRANCH}"
            - name: "dashboard-service"
              tag: "latest"
              commit: "\${GIT_COMMIT}"
              branch: "\${BRANCH}"

  - name: deploy
    steps:
      - name: generate-compose-file
        action: generate-config
        params:
          template: "$WORKDIR/source/docker-compose.yml"
          output: "$WORKDIR/docker-compose.yml"
          variables:
            REGISTRY: "\${LOCAL_REGISTRY}"
            TAG: "latest"
      
      - name: deploy-to-localhost
        action: deploy-compose
        params:
          compose_file: "$WORKDIR/docker-compose.yml"
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
EOF
    
    print_success "Pipeline configuration created at $WORKDIR/qlcv-pipeline.yaml"
}

# Function to execute pipeline
execute_pipeline() {
    print_header "Executing Pipeline"
    
    # Show pipeline configuration
    print_info "Pipeline configuration:"
    cat "$WORKDIR/qlcv-pipeline.yaml"
    
    # Execute pipeline
    print_info "Submitting pipeline to PipeslicerCI..."
    
    # Use the correct API call format
    response=$(curl -X POST "http://127.0.0.1:$PIPESLICER_PORT/pipelines/build" \
        -H "Content-Type: multipart/form-data" \
        -F "url=$QLCV_REPO" \
        -F "branch=$QLCV_BRANCH" \
        -F "file=@$WORKDIR/qlcv-pipeline.yaml")
    
    # Check if the pipeline execution was successful
    if [[ "$response" == *"Successfully executed pipeline"* ]]; then
        print_success "Pipeline executed successfully"
        
        # Print the pipeline output
        print_info "Pipeline Output:"
        echo "$response"
    else
        print_error "Pipeline execution failed"
        print_info "Error Output:"
        echo "$response"
        
        # Show logs
        print_info "Container logs:"
        docker logs --tail=100 pipeslicerci
        
        exit 1
    fi
}

# Function to verify deployment
verify_deployment() {
    print_header "Verifying Deployment"
    
    # Check if frontend is accessible
    print_info "Checking frontend..."
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:80 | grep -q "200"; then
        print_success "Frontend is accessible"
    else
        print_error "Frontend is not accessible"
    fi
    
    # Check if API Gateway is accessible
    print_info "Checking API Gateway..."
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:3000 | grep -q "200"; then
        print_success "API Gateway is accessible"
    else
        print_error "API Gateway is not accessible"
    fi
    
    # Check Docker containers
    print_info "Checking Docker containers..."
    docker-compose -f "$WORKDIR/docker-compose.yml" ps
}

# Function to display deployment information
display_info() {
    print_header "Deployment Information"
    
    echo -e "${YELLOW}QLCV Application URLs:${NC}"
    echo -e "Frontend: ${GREEN}http://localhost:80${NC}"
    echo -e "API Gateway: ${GREEN}http://localhost:3000${NC}"
    echo -e "Auth Service: ${GREEN}http://localhost:3001${NC}"
    echo -e "Asset Service: ${GREEN}http://localhost:3002${NC}"
    echo -e "Bidding Service: ${GREEN}http://localhost:3003${NC}"
    echo -e "Dashboard Service: ${GREEN}http://localhost:3004${NC}"
    
    echo -e "\n${YELLOW}Useful Commands:${NC}"
    echo -e "View logs: ${GREEN}docker-compose -f $WORKDIR/docker-compose.yml logs -f${NC}"
    echo -e "Restart service: ${GREEN}docker-compose -f $WORKDIR/docker-compose.yml restart [service_name]${NC}"
    echo -e "Stop all services: ${GREEN}docker-compose -f $WORKDIR/docker-compose.yml down${NC}"
    echo -e "Start all services: ${GREEN}docker-compose -f $WORKDIR/docker-compose.yml up -d${NC}"
    echo -e "View PipeslicerCI logs: ${GREEN}docker logs pipeslicerci${NC}"
}

# Main function
main() {
    print_header "QLCV CI/CD Pipeline"
    
    # Check prerequisites
    check_prerequisites
    
    # Set up PipeslicerCI
    setup_pipeslicerci
    
    # Set up local Docker registry
    setup_registry
    
    # Create pipeline configuration
    create_pipeline_config
    
    # Execute pipeline
    execute_pipeline
    
    # Verify deployment
    verify_deployment
    
    # Display deployment information
    display_info
    
    print_header "Pipeline Execution Completed"
    print_success "QLCV has been successfully deployed to localhost!"
}

# Execute main function
main
