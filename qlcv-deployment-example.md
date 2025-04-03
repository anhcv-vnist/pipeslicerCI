# Ví Dụ Triển Khai Thực Tế QLCV với PipeslicerCI

Tài liệu này cung cấp một ví dụ thực tế về cách sử dụng PipeslicerCI để triển khai dự án QLCV (Quản Lý Công Việc) từ đầu đến cuối. Ví dụ này bao gồm các lệnh cụ thể, cấu hình và kịch bản triển khai.

## Kịch Bản Triển Khai

Trong ví dụ này, chúng ta sẽ triển khai dự án QLCV với các thành phần sau:

- Frontend: React application (client-vite)
- Backend: Các microservices (auth-service, api-gateway-service, v.v.)
- Docker Registry: Harbor (tại registry.example.com)
- Môi trường: Staging và Production (sử dụng Docker Swarm)

## Bước 1: Cài Đặt và Khởi Động PipeslicerCI

```bash
# Clone PipeslicerCI repository
git clone https://github.com/vanhcao3/pipeslicerCI.git
cd pipeslicerCI

# Build PipeslicerCI
make build

# Tạo thư mục cấu hình và dữ liệu
sudo mkdir -p /etc/pipeslicerci
sudo mkdir -p /var/lib/pipeslicerci
sudo mkdir -p /var/log/pipeslicerci

# Tạo file cấu hình
cat > /etc/pipeslicerci/config.yaml << EOF
server:
  port: 3000
  host: "0.0.0.0"

database:
  path: "/var/lib/pipeslicerci/data.db"

registry:
  default_type: "harbor"

logging:
  level: "info"
  file: "/var/log/pipeslicerci/server.log"

workdir:
  path: "/var/lib/pipeslicerci/workdir"
EOF

# Khởi động PipeslicerCI
./build/pipeslicerci-web
```

## Bước 2: Cấu Hình Kết Nối với Docker Registry

```bash
# Xác thực với Docker Registry
curl -X POST "http://localhost:3000/registry-connector/authenticate" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://registry.example.com",
    "username": "qlcv-cicd",
    "password": "your-secure-password",
    "type": "harbor",
    "insecure": false
  }'

# Kiểm tra kết nối
curl -X GET "http://localhost:3000/registry-connector/repositories?url=https://registry.example.com&username=qlcv-cicd&password=your-secure-password"
```

## Bước 3: Tạo Pipeline Configuration cho QLCV

Tạo file `qlcv-pipeline.yaml` với nội dung sau:

```yaml
version: "1.0"
name: "qlcv-pipeline"
description: "CI/CD Pipeline cho dự án QLCV"

variables:
  REGISTRY_URL: "registry.example.com"
  REGISTRY_USERNAME: "qlcv-cicd"
  REGISTRY_PASSWORD: "your-secure-password"
  GIT_REPO: "https://github.com/username/qlcv-refactor.git"

stages:
  - name: test
    steps:
      - name: test-frontend
        action: run-tests
        params:
          url: "${GIT_REPO}"
          branch: "${BRANCH:-main}"
          servicePath: "client-vite"
          command: "npm test"
      
      - name: test-auth-service
        action: run-tests
        params:
          url: "${GIT_REPO}"
          branch: "${BRANCH:-main}"
          servicePath: "micro-services/auth-service"
          command: "npm test"
      
      - name: test-api-gateway
        action: run-tests
        params:
          url: "${GIT_REPO}"
          branch: "${BRANCH:-main}"
          servicePath: "micro-services/api-gateway-service"
          command: "./mvnw test"

  - name: build
    steps:
      - name: build-frontend
        action: build-image
        params:
          url: "${GIT_REPO}"
          branch: "${BRANCH:-main}"
          servicePath: "client-vite"
          tag: "${TAG:-latest}"
          registry: "${REGISTRY_URL}"
          username: "${REGISTRY_USERNAME}"
          password: "${REGISTRY_PASSWORD}"
          dockerfile: "Dockerfile"
      
      - name: build-auth-service
        action: build-image
        params:
          url: "${GIT_REPO}"
          branch: "${BRANCH:-main}"
          servicePath: "micro-services/auth-service"
          tag: "${TAG:-latest}"
          registry: "${REGISTRY_URL}"
          username: "${REGISTRY_USERNAME}"
          password: "${REGISTRY_PASSWORD}"
          dockerfile: "Dockerfile"
      
      - name: build-api-gateway
        action: build-image
        params:
          url: "${GIT_REPO}"
          branch: "${BRANCH:-main}"
          servicePath: "micro-services/api-gateway-service"
          tag: "${TAG:-latest}"
          registry: "${REGISTRY_URL}"
          username: "${REGISTRY_USERNAME}"
          password: "${REGISTRY_PASSWORD}"
          dockerfile: "Dockerfile"
      
      - name: build-asset-service
        action: build-image
        params:
          url: "${GIT_REPO}"
          branch: "${BRANCH:-main}"
          servicePath: "micro-services/asset-service"
          tag: "${TAG:-latest}"
          registry: "${REGISTRY_URL}"
          username: "${REGISTRY_USERNAME}"
          password: "${REGISTRY_PASSWORD}"
          dockerfile: "Dockerfile"
      
      - name: build-bidding-service
        action: build-image
        params:
          url: "${GIT_REPO}"
          branch: "${BRANCH:-main}"
          servicePath: "micro-services/bidding-service"
          tag: "${TAG:-latest}"
          registry: "${REGISTRY_URL}"
          username: "${REGISTRY_USERNAME}"
          password: "${REGISTRY_PASSWORD}"
          dockerfile: "Dockerfile"
      
      - name: build-dashboard-service
        action: build-image
        params:
          url: "${GIT_REPO}"
          branch: "${BRANCH:-main}"
          servicePath: "micro-services/dashboard-service"
          tag: "${TAG:-latest}"
          registry: "${REGISTRY_URL}"
          username: "${REGISTRY_USERNAME}"
          password: "${REGISTRY_PASSWORD}"
          dockerfile: "Dockerfile"

  - name: deploy-staging
    steps:
      - name: prepare-staging-config
        action: prepare-config
        params:
          environment: "staging"
          configTemplate: "staging-config.yaml"
          outputPath: "/tmp/qlcv-staging-config.yaml"
          variables:
            REGISTRY_URL: "${REGISTRY_URL}"
            TAG: "${TAG:-latest}"
      
      - name: deploy-to-staging
        action: deploy
        params:
          environment: "staging"
          configFile: "/tmp/qlcv-staging-config.yaml"
          sshHost: "staging.example.com"
          sshUser: "deploy"
          sshKeyPath: "/var/lib/pipeslicerci/ssh/id_rsa"
          command: "docker stack deploy -c /tmp/qlcv-staging-config.yaml qlcv"

  - name: deploy-production
    when: manual
    steps:
      - name: prepare-production-config
        action: prepare-config
        params:
          environment: "production"
          configTemplate: "production-config.yaml"
          outputPath: "/tmp/qlcv-production-config.yaml"
          variables:
            REGISTRY_URL: "${REGISTRY_URL}"
            TAG: "${TAG:-latest}"
      
      - name: deploy-to-production
        action: deploy
        params:
          environment: "production"
          configFile: "/tmp/qlcv-production-config.yaml"
          sshHost: "production.example.com"
          sshUser: "deploy"
          sshKeyPath: "/var/lib/pipeslicerci/ssh/id_rsa"
          command: "docker stack deploy -c /tmp/qlcv-production-config.yaml qlcv"
```

## Bước 4: Tạo Cấu Hình Triển Khai cho Staging

Tạo file `staging-config.yaml` với nội dung sau:

```yaml
version: "3.8"
services:
  client-vite:
    image: ${REGISTRY_URL}/client-vite:${TAG}
    ports:
      - "80:80"
    environment:
      - API_URL=http://api-gateway:3000
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 10s
        order: start-first

  api-gateway:
    image: ${REGISTRY_URL}/api-gateway-service:${TAG}
    ports:
      - "3000:3000"
    environment:
      - CONSUL_URL=http://consul:8500
      - VAULT_URL=http://vault:8200
      - AUTH_SERVICE_URL=http://auth-service:3001
      - ASSET_SERVICE_URL=http://asset-service:3002
      - BIDDING_SERVICE_URL=http://bidding-service:3003
      - DASHBOARD_SERVICE_URL=http://dashboard-service:3004
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 10s
        order: start-first

  auth-service:
    image: ${REGISTRY_URL}/auth-service:${TAG}
    environment:
      - MONGO_URI=mongodb://mongo:27017/auth
      - CONSUL_URL=http://consul:8500
      - VAULT_URL=http://vault:8200
      - PORT=3001
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 10s
        order: start-first

  asset-service:
    image: ${REGISTRY_URL}/asset-service:${TAG}
    environment:
      - MONGO_URI=mongodb://mongo:27017/asset
      - CONSUL_URL=http://consul:8500
      - VAULT_URL=http://vault:8200
      - PORT=3002
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 10s
        order: start-first

  bidding-service:
    image: ${REGISTRY_URL}/bidding-service:${TAG}
    environment:
      - MONGO_URI=mongodb://mongo:27017/bidding
      - CONSUL_URL=http://consul:8500
      - VAULT_URL=http://vault:8200
      - PORT=3003
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 10s
        order: start-first

  dashboard-service:
    image: ${REGISTRY_URL}/dashboard-service:${TAG}
    environment:
      - MONGO_URI=mongodb://mongo:27017/dashboard
      - CONSUL_URL=http://consul:8500
      - VAULT_URL=http://vault:8200
      - PORT=3004
    deploy:
      replicas: 2
      update_config:
        parallelism: 1
        delay: 10s
        order: start-first

  mongo:
    image: mongo:4.4
    volumes:
      - mongo-data:/data/db
    deploy:
      placement:
        constraints:
          - node.role == manager

  consul:
    image: consul:1.9
    ports:
      - "8500:8500"
    volumes:
      - consul-data:/consul/data
    deploy:
      placement:
        constraints:
          - node.role == manager

  vault:
    image: vault:1.6
    ports:
      - "8200:8200"
    environment:
      - VAULT_DEV_ROOT_TOKEN_ID=root
    cap_add:
      - IPC_LOCK
    deploy:
      placement:
        constraints:
          - node.role == manager

volumes:
  mongo-data:
  consul-data:

networks:
  default:
    driver: overlay
```

## Bước 5: Tạo Cấu Hình Triển Khai cho Production

Tạo file `production-config.yaml` với nội dung tương tự như `staging-config.yaml` nhưng với các cấu hình phù hợp cho môi trường production:

```yaml
version: "3.8"
services:
  client-vite:
    image: ${REGISTRY_URL}/client-vite:${TAG}
    ports:
      - "80:80"
    environment:
      - API_URL=http://api-gateway:3000
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 30s
        order: start-first
      resources:
        limits:
          cpus: '0.5'
          memory: 512M

  # Các services khác tương tự, với cấu hình phù hợp cho production
  # ...
```

## Bước 6: Cấu Hình Webhook để Tự Động Kích Hoạt Pipeline

### 6.1. Tạo Webhook Endpoint

```bash
# Tạo webhook secret
WEBHOOK_SECRET=$(openssl rand -hex 20)
echo "Webhook Secret: $WEBHOOK_SECRET"

# Cấu hình webhook trong PipeslicerCI
cat >> /etc/pipeslicerci/config.yaml << EOF
webhooks:
  github:
    secret: "${WEBHOOK_SECRET}"
    pipeline: "/var/lib/pipeslicerci/qlcv-pipeline.yaml"
EOF

# Khởi động lại PipeslicerCI
# ...
```

### 6.2. Cấu Hình Webhook trên GitHub

1. Truy cập repository QLCV trên GitHub
2. Chọn "Settings" > "Webhooks" > "Add webhook"
3. Nhập thông tin:
   - Payload URL: `http://your-pipeslicerci-server:3000/webhooks/github`
   - Content type: `application/json`
   - Secret: `$WEBHOOK_SECRET`
   - Events: Chọn "Just the push event"
4. Nhấn "Add webhook"

## Bước 7: Thực Thi Pipeline Thủ Công

```bash
# Sao chép pipeline configuration
cp qlcv-pipeline.yaml /var/lib/pipeslicerci/
cp staging-config.yaml /var/lib/pipeslicerci/
cp production-config.yaml /var/lib/pipeslicerci/

# Thực thi pipeline
curl -X POST "http://localhost:3000/pipelines/build" \
  -F "url=https://github.com/username/qlcv-refactor.git" \
  -F "branch=main" \
  -F "file=@/var/lib/pipeslicerci/qlcv-pipeline.yaml"
```

## Bước 8: Giám Sát Quá Trình Thực Thi

### 8.1. Kiểm Tra Trạng Thái Pipeline

```bash
# Lấy danh sách các pipeline đang chạy
curl -X GET "http://localhost:3000/pipelines/executions"

# Lấy chi tiết về một pipeline cụ thể
curl -X GET "http://localhost:3000/pipelines/executions/1"
```

### 8.2. Kiểm Tra Logs

```bash
# Xem logs của PipeslicerCI
tail -f /var/log/pipeslicerci/server.log

# Xem logs của các containers trên môi trường staging
ssh deploy@staging.example.com "docker service logs qlcv_client-vite"
ssh deploy@staging.example.com "docker service logs qlcv_api-gateway"
ssh deploy@staging.example.com "docker service logs qlcv_auth-service"
```

## Bước 9: Phê Duyệt Triển Khai lên Production

```bash
# Phê duyệt stage "deploy-production" của pipeline có ID là 1
curl -X POST "http://localhost:3000/pipelines/executions/1/stages/deploy-production/approve"
```

## Bước 10: Kiểm Tra Ứng Dụng Đã Triển Khai

### 10.1. Kiểm Tra Môi Trường Staging

```bash
# Kiểm tra các services đang chạy
ssh deploy@staging.example.com "docker stack services qlcv"

# Kiểm tra trạng thái của từng service
ssh deploy@staging.example.com "docker service ps qlcv_client-vite"
ssh deploy@staging.example.com "docker service ps qlcv_api-gateway"
ssh deploy@staging.example.com "docker service ps qlcv_auth-service"
```

### 10.2. Kiểm Tra Môi Trường Production

```bash
# Kiểm tra các services đang chạy
ssh deploy@production.example.com "docker stack services qlcv"

# Kiểm tra trạng thái của từng service
ssh deploy@production.example.com "docker service ps qlcv_client-vite"
ssh deploy@production.example.com "docker service ps qlcv_api-gateway"
ssh deploy@production.example.com "docker service ps qlcv_auth-service"
```

## Bước 11: Rollback (Nếu Cần)

```bash
# Lấy danh sách các images đã build cho service "client-vite"
curl -X GET "http://localhost:3000/registry/services/client-vite/history?limit=10"

# Rollback service "client-vite" về phiên bản trước đó
curl -X POST "http://localhost:3000/registry/services/client-vite/tags" \
  -H "Content-Type: application/json" \
  -d '{
    "sourceTag": "previous-stable-tag",
    "newTag": "latest"
  }'

# Thực thi pipeline với tag cụ thể
curl -X POST "http://localhost:3000/pipelines/build" \
  -F "url=https://github.com/username/qlcv-refactor.git" \
  -F "branch=main" \
  -F "file=@/var/lib/pipeslicerci/qlcv-pipeline.yaml" \
  -F "variables[TAG]=previous-stable-tag"
```

## Kết Luận

Ví dụ triển khai này đã minh họa cách sử dụng PipeslicerCI để thiết lập một CI/CD pipeline hoàn chỉnh cho dự án QLCV. Quy trình này bao gồm:

1. Kiểm thử tự động
2. Build Docker images
3. Triển khai lên môi trường staging
4. Phê duyệt thủ công trước khi triển khai lên production
5. Giám sát và xử lý lỗi
6. Rollback khi cần thiết

Bằng cách tuân theo quy trình này, bạn có thể đảm bảo rằng dự án QLCV được triển khai một cách nhất quán, đáng tin cậy và an toàn.

## Tài Liệu Tham Khảo

- [PipeslicerCI Documentation](https://github.com/vanhcao3/pipeslicerCI)
- [Docker Swarm Documentation](https://docs.docker.com/engine/swarm/)
- [Harbor Registry Documentation](https://goharbor.io/docs/)
- [GitHub Webhooks Documentation](https://docs.github.com/en/developers/webhooks-and-events/webhooks)
