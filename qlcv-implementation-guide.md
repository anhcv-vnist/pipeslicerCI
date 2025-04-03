# Hướng Dẫn Triển Khai CI/CD Pipeline cho Dự Án QLCV

Tài liệu này cung cấp hướng dẫn chi tiết về cách triển khai CI/CD Pipeline cho dự án QLCV (Quản Lý Công Việc) sử dụng PipeslicerCI, dựa trên kịch bản demo đã được mô tả trong `demo-script-qlcv.md`.

## Phần 1: Chuẩn Bị Môi Trường

### 1.1. Yêu Cầu Hệ Thống

- Docker và Docker Compose
- Git
- Go (phiên bản 1.16 trở lên)
- Quyền truy cập vào Docker Registry (Harbor, DockerHub, hoặc tương tự)
- Máy chủ Staging và Production (có thể sử dụng Kubernetes hoặc Docker Swarm)

### 1.2. Cài Đặt PipeslicerCI

```bash
# Clone PipeslicerCI repository
git clone https://github.com/vanhcao3/pipeslicerCI.git
cd pipeslicerCI

# Cài đặt dependencies
go mod download

# Build PipeslicerCI
make build

# Tạo thư mục cấu hình
mkdir -p /etc/pipeslicerci
cp config.example.yaml /etc/pipeslicerci/config.yaml

# Chỉnh sửa file cấu hình
nano /etc/pipeslicerci/config.yaml
```

Cấu hình mẫu cho `/etc/pipeslicerci/config.yaml`:

```yaml
server:
  port: 3000
  host: "0.0.0.0"

database:
  path: "/var/lib/pipeslicerci/data.db"

registry:
  default_type: "harbor"  # hoặc "dockerhub", "generic"

logging:
  level: "info"
  file: "/var/log/pipeslicerci/server.log"

workdir:
  path: "/var/lib/pipeslicerci/workdir"
```

### 1.3. Khởi Động PipeslicerCI

```bash
# Tạo thư mục dữ liệu và logs
mkdir -p /var/lib/pipeslicerci
mkdir -p /var/log/pipeslicerci

# Khởi động PipeslicerCI
./build/pipeslicerci-web
```

Hoặc sử dụng Docker:

```bash
docker run -d \
  --name pipeslicerci \
  -p 3000:3000 \
  -v /etc/pipeslicerci:/etc/pipeslicerci \
  -v /var/lib/pipeslicerci:/var/lib/pipeslicerci \
  -v /var/log/pipeslicerci:/var/log/pipeslicerci \
  vanhcao3/pipeslicerci:latest
```

## Phần 2: Cấu Hình Docker Registry

### 2.1. Tạo Tài Khoản Registry

Đảm bảo bạn đã có tài khoản trên Docker Registry (Harbor, DockerHub, v.v.) với quyền push images.

### 2.2. Cấu Hình Registry Connector

Sử dụng API để cấu hình kết nối với Docker Registry:

```bash
curl -X POST "http://localhost:3000/registry-connector/authenticate" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://registry.example.com",
    "username": "username",
    "password": "password",
    "type": "harbor",
    "insecure": false
  }'
```

Hoặc sử dụng giao diện web tại http://localhost:3000/swagger.

### 2.3. Kiểm Tra Kết Nối

```bash
curl -X GET "http://localhost:3000/registry-connector/repositories?url=https://registry.example.com&username=username&password=password"
```

## Phần 3: Cấu Hình CI/CD Pipeline cho QLCV

### 3.1. Tạo Pipeline Configuration

Tạo file `qlcv-pipeline.yaml` với nội dung như đã mô tả trong kịch bản demo.

### 3.2. Tùy Chỉnh Pipeline cho Dự Án QLCV

Cập nhật các thông tin sau trong file `qlcv-pipeline.yaml`:

1. **Repository URL**: Cập nhật `GIT_REPO` thành URL của repository QLCV
2. **Registry Information**: Cập nhật thông tin Docker Registry
3. **Services**: Thêm tất cả các microservices của dự án QLCV
4. **Environment Configuration**: Cập nhật cấu hình cho môi trường staging và production

### 3.3. Thêm Kiểm Thử Tự Động

Bổ sung các bước kiểm thử tự động vào pipeline:

```yaml
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
      
      # Thêm các microservices khác tương tự
```

## Phần 4: Tích Hợp với Hệ Thống QLCV

### 4.1. Cấu Hình Webhook

Cấu hình webhook từ Git repository để tự động kích hoạt pipeline khi có commit mới:

1. Truy cập cài đặt repository trên GitHub/GitLab
2. Thêm webhook với URL: `http://your-pipeslicerci-server:3000/webhooks/github` (hoặc `/webhooks/gitlab`)
3. Chọn các sự kiện: Push, Pull Request

### 4.2. Cấu Hình Môi Trường Staging

Chuẩn bị môi trường staging với các thành phần sau:

1. **Docker Swarm hoặc Kubernetes**: Để triển khai các microservices
2. **Consul**: Để quản lý cấu hình
3. **Vault**: Để quản lý secrets
4. **MongoDB**: Cơ sở dữ liệu cho các microservices

Tạo file cấu hình triển khai cho môi trường staging:

```yaml
# staging-deploy.yaml
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

  api-gateway:
    image: ${REGISTRY_URL}/api-gateway-service:${TAG}
    ports:
      - "3000:3000"
    environment:
      - CONSUL_URL=http://consul:8500
      - VAULT_URL=http://vault:8200
    deploy:
      replicas: 2

  auth-service:
    image: ${REGISTRY_URL}/auth-service:${TAG}
    environment:
      - MONGO_URI=mongodb://mongo:27017/auth
      - CONSUL_URL=http://consul:8500
      - VAULT_URL=http://vault:8200
    deploy:
      replicas: 2

  # Thêm các microservices khác tương tự

  mongo:
    image: mongo:4.4
    volumes:
      - mongo-data:/data/db

  consul:
    image: consul:1.9
    ports:
      - "8500:8500"
    volumes:
      - consul-data:/consul/data

  vault:
    image: vault:1.6
    ports:
      - "8200:8200"
    environment:
      - VAULT_DEV_ROOT_TOKEN_ID=root
    cap_add:
      - IPC_LOCK

volumes:
  mongo-data:
  consul-data:
```

### 4.3. Cấu Hình Môi Trường Production

Tương tự như môi trường staging, nhưng với cấu hình phù hợp cho production:

```yaml
# production-deploy.yaml
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
        delay: 10s
        order: start-first

  # Các services khác tương tự, với cấu hình phù hợp cho production
```

## Phần 5: Thực Thi và Giám Sát Pipeline

### 5.1. Thực Thi Pipeline Thủ Công

```bash
curl -X POST "http://localhost:3000/pipelines/build" \
  -F "url=https://github.com/username/qlcv-refactor.git" \
  -F "branch=main" \
  -F "file=@qlcv-pipeline.yaml"
```

### 5.2. Giám Sát Quá Trình Thực Thi

1. Truy cập giao diện web PipeslicerCI tại http://localhost:3000
2. Chọn "Pipelines" > "Current Executions"
3. Xem chi tiết quá trình thực thi pipeline

### 5.3. Xử Lý Lỗi

Khi gặp lỗi trong quá trình thực thi pipeline:

1. Kiểm tra logs của PipeslicerCI
2. Kiểm tra logs của các bước trong pipeline
3. Sửa lỗi và thực thi lại pipeline

## Phần 6: Quy Trình Làm Việc với CI/CD Pipeline

### 6.1. Quy Trình Phát Triển

1. **Developer**:
   - Tạo branch mới từ `main`
   - Phát triển tính năng
   - Tạo Pull Request

2. **CI/CD Pipeline**:
   - Tự động kích hoạt khi có Pull Request
   - Thực hiện kiểm thử
   - Build Docker images
   - Triển khai lên môi trường staging

3. **QA**:
   - Kiểm thử trên môi trường staging
   - Phê duyệt Pull Request nếu mọi thứ hoạt động tốt

4. **CI/CD Pipeline**:
   - Merge Pull Request vào `main`
   - Tự động kích hoạt pipeline cho branch `main`
   - Triển khai lên môi trường production (sau khi được phê duyệt)

### 6.2. Quy Trình Phát Hành

1. **Tạo Tag**:
   - Tạo tag cho phiên bản mới (ví dụ: v1.0.0)
   - Push tag lên repository

2. **CI/CD Pipeline**:
   - Tự động kích hoạt khi có tag mới
   - Build Docker images với tag tương ứng
   - Triển khai lên môi trường staging

3. **QA**:
   - Kiểm thử trên môi trường staging
   - Phê duyệt phiên bản nếu mọi thứ hoạt động tốt

4. **CI/CD Pipeline**:
   - Triển khai lên môi trường production (sau khi được phê duyệt)

### 6.3. Quy Trình Rollback

Trong trường hợp cần rollback:

1. Truy cập giao diện web PipeslicerCI
2. Chọn "Registry" > "Services" > [Service Name] > "History"
3. Chọn phiên bản cần rollback
4. Nhấn "Deploy" và chọn môi trường cần rollback

## Phần 7: Bảo Mật và Quản Lý Secrets

### 7.1. Sử Dụng Vault

1. Cấu hình Vault để lưu trữ secrets:
   - API keys
   - Database credentials
   - OAuth credentials

2. Cấu hình PipeslicerCI để sử dụng Vault:
   - Thêm Vault token vào cấu hình PipeslicerCI
   - Sử dụng Vault trong pipeline để truy xuất secrets

### 7.2. Quản Lý Credentials

1. Không lưu trữ credentials trực tiếp trong file pipeline
2. Sử dụng biến môi trường hoặc Vault để quản lý credentials
3. Sử dụng Consul để quản lý cấu hình không nhạy cảm

## Phần 8: Mở Rộng và Tùy Chỉnh

### 8.1. Thêm Microservices Mới

Khi thêm microservice mới vào dự án QLCV:

1. Thêm bước build và test cho microservice mới vào pipeline
2. Thêm cấu hình triển khai cho microservice mới
3. Cập nhật file pipeline và triển khai lại

### 8.2. Tùy Chỉnh Pipeline

Tùy chỉnh pipeline để phù hợp với yêu cầu cụ thể:

1. Thêm các bước kiểm thử bảo mật
2. Thêm các bước phân tích code
3. Tích hợp với các công cụ giám sát

### 8.3. Tích Hợp với Công Cụ Khác

Tích hợp PipeslicerCI với các công cụ khác:

1. **Slack/Teams**: Thông báo kết quả pipeline
2. **Jira**: Cập nhật trạng thái task
3. **Prometheus/Grafana**: Giám sát hiệu suất

## Kết Luận

Tài liệu này cung cấp hướng dẫn chi tiết về cách triển khai CI/CD Pipeline cho dự án QLCV sử dụng PipeslicerCI. Bằng cách tuân theo hướng dẫn này, bạn có thể thiết lập một quy trình CI/CD tự động, hiệu quả và đáng tin cậy cho dự án QLCV.

## Tài Liệu Tham Khảo

- [PipeslicerCI Documentation](https://github.com/vanhcao3/pipeslicerCI)
- [Docker Documentation](https://docs.docker.com/)
- [Consul Documentation](https://www.consul.io/docs)
- [Vault Documentation](https://www.vaultproject.io/docs)
- [Microservices Architecture](https://microservices.io/)
