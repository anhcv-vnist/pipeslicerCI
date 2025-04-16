# QLCV CI/CD Pipeline Documentation

This repository contains comprehensive documentation and implementation scripts for setting up a CI/CD pipeline for the QLCV (Quản Lý Công Việc) project using PipeslicerCI.

## Documentation Overview

The documentation is organized into several files, each focusing on different aspects of the CI/CD pipeline:

1. **qlcv-cicd-summary.md**: A high-level overview of the CI/CD pipeline, including the stages, components, and benefits.
2. **demo-script-qlcv.md**: A detailed demo script that outlines the CI/CD pipeline for the QLCV project.
3. **qlcv-implementation-guide.md**: Step-by-step implementation guide for setting up the CI/CD pipeline.
4. **qlcv-deployment-example.md**: Practical example with specific commands and configurations for deployment.
5. **qlcv-pipeline-localhost.md**: Detailed explanation of the CI/CD pipeline stages, inputs, outputs, and PipeslicerCI functionality for localhost deployment.
6. **qlcv-pipeline-script.sh**: Executable script that implements the CI/CD pipeline for localhost deployment.

## CI/CD Pipeline Stages

The CI/CD pipeline for the QLCV project consists of the following stages:

1. **Code Checkout**: Retrieve source code from the repository
2. **Testing**: Run automated tests to ensure code quality
3. **Building**: Build Docker images for all microservices
4. **Registry Management**: Store and manage Docker images
5. **Local Deployment**: Deploy to localhost for testing
6. **Monitoring**: Monitor deployment and verify functionality

## PipeslicerCI Components

PipeslicerCI provides several key components that are used in different stages of the CI/CD pipeline:

| Component | Description |
|-----------|-------------|
| Image Builder | Builds Docker images from source code |
| Registry Manager | Manages metadata about Docker images |
| Registry Connector | Connects to Docker registries |
| Config Manager | Manages configuration for deployments |
| Pipeline Executor | Orchestrates the CI/CD pipeline |

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Git
- Go (version 1.16 or later)
- curl
- jq

### Installation

1. Clone the PipeslicerCI repository:
   ```bash
   git clone https://github.com/vanhcao3/pipeslicerCI.git
   cd pipeslicerCI
   ```

2. Build PipeslicerCI:
   ```bash
   make build
   ```

3. Make the pipeline script executable:
   ```bash
   chmod +x qlcv-pipeline-script.sh
   ```

### Running the CI/CD Pipeline

To run the CI/CD pipeline for localhost deployment:

```bash
./qlcv-pipeline-script.sh
```

This script will:
1. Check prerequisites
2. Set up PipeslicerCI (dockerized for better isolation and reliability)
3. Set up a local Docker registry
4. Create pipeline configuration
5. Execute the pipeline using the correct API format
6. Verify deployment
7. Display deployment information

The script has been updated to use the correct API call format for the PipeslicerCI `/pipelines/build` endpoint:

```bash
curl -X POST "http://127.0.0.1:3000/pipelines/build" \
  -H "Content-Type: multipart/form-data" \
  -F "url=git@github.com:vanhcao3/pipeslicerCI.git" \
  -F "branch=main" \
  -F "file=@$WORKDIR/qlcv-pipeline.yaml"
```

This ensures that the pipeline is executed correctly and the output is displayed in real-time.

## Customizing the Pipeline

You can customize the pipeline by modifying the following files:

- **qlcv-pipeline-script.sh**: Update the configuration variables at the top of the file
- **qlcv-pipeline-localhost.md**: Modify the pipeline stages, inputs, outputs, and PipeslicerCI functionality

## Troubleshooting

If you encounter issues with the CI/CD pipeline, check the following:

1. PipeslicerCI logs:
   ```bash
   tail -f ~/.pipeslicerci/logs/pipeslicerci.log
   ```

2. Docker container logs:
   ```bash
   docker-compose -f /tmp/qlcv-pipeline/docker-compose.yml logs
   ```

3. Docker container status:
   ```bash
   docker-compose -f /tmp/qlcv-pipeline/docker-compose.yml ps
   ```

## Additional Resources

- [PipeslicerCI Documentation](https://github.com/vanhcao3/pipeslicerCI)
- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Microservices Architecture](https://microservices.io/)

## Conclusion

This documentation provides a comprehensive guide to setting up a CI/CD pipeline for the QLCV project using PipeslicerCI. By following the steps and scripts provided, you can quickly set up a local CI/CD pipeline for the QLCV project, enabling faster development cycles and more reliable deployments.
