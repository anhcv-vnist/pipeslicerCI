# Kịch Bản Demo CI/CD Pipeline cho Dự Án QLCV

## Giới Thiệu

Tài liệu này mô tả kịch bản demo chi tiết về cách triển khai CI/CD Pipeline cho dự án QLCV (Quản Lý Công Việc) sử dụng công cụ PipeslicerCI. Kịch bản này sẽ hướng dẫn từng bước trong quy trình CI/CD, từ commit code đến triển khai sản phẩm cuối cùng.

## Tổng Quan về Dự Án QLCV

QLCV là một hệ thống quản lý công việc dựa trên kiến trúc microservices, bao gồm:
- Frontend: Ứng dụng React (client-vite)
- Backend: Nhiều microservices (auth-service, api-gateway-service, asset-service, v.v.)
- Cơ sở dữ liệu: MongoDB
- Dịch vụ phụ trợ: Consul, Vault

## Mục Tiêu của CI/CD Pipeline

1. Tự động hóa quy trình build và test
2. Đảm bảo chất lượng code
3. Triển khai liên tục các microservices
4. Quản lý phiên bản Docker images
5. Triển khai tự động lên môi trường staging và production

## Các Thành Phần của PipeslicerCI

PipeslicerCI cung cấp các thành phần chính sau để hỗ trợ CI/CD pipeline:

1. **Image Builder**: Xây dựng Docker images từ mã nguồn
2. **Registry Manager**: Quản lý metadata của Docker images
3. **Registry Connector**: Kết nối với Docker registry
4. **Config Manager**: Quản lý cấu hình triển khai
5. **Pipeline Executor**: Thực thi các pipeline CI/CD

## Kịch Bản Demo

### Bước 1: Cài Đặt và Cấu Hình PipeslicerCI

```bash
# Clone PipeslicerCI repository
git clone https://github.com/vanhcao3/pipeslicerCI.git
cd pipeslicerCI

# Build PipeslicerCI
make build

# Khởi động PipeslicerCI
./build/pipeslicerci-web
```

Truy cập giao diện web tại http://localhost:3000/swagger để xem API documentation.

### Bước 2: Cấu Hình Kết Nối Docker Registry

Sử dụng Registry Connector để kết nối với Docker registry:

1. Truy cập giao diện web PipeslicerCI
2. Chọn "Registry Connector" > "Authenticate"
3. Nhập thông tin kết nối:
   - URL: `https://registry.example.com`
   - Username: `username`
   - Password: `password`
   - Type: `harbor` (hoặc loại registry phù hợp)

Hoặc sử dụng API:

```bash
curl -X POST "http://localhost:3000/registry-connector/authenticate" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://registry.example.com",
    "username": "username",
    "password": "password",
    "type": "harbor"
  }'
```

### Bước 3: Tạo Pipeline Configuration cho QLCV

Tạo file `qlcv-pipeline.yaml` với nội dung sau:

```yaml
version: "1.0"
name: "qlcv-pipeline"
description: "CI/CD Pipeline cho dự án QLCV"

variables:
  REGISTRY_URL: "registry.example.com"
  REGISTRY_USERNAME: "username"
  REGISTRY_PASSWORD: "password"
  GIT_REPO: "https://github.com/username/qlcv-refactor.git"

stages:
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
      
      # Thêm các microservices khác tương tự

  - name: deploy-staging
    steps:
      - name: deploy-to-staging
        action: deploy
        params:
          environment: "staging"
          services:
            - name: "client-vite"
              image: "${REGISTRY_URL}/client-vite:${TAG:-latest}"
            - name: "auth-service"
              image: "${REGISTRY_URL}/auth-service:${TAG:-latest}"
            - name: "api-gateway-service"
              image: "${REGISTRY_URL}/api-gateway-service:${TAG:-latest}"
            # Thêm các microservices khác tương tự
          config:
            consul_url: "http://consul:8500"
            vault_url: "http://vault:8200"

  - name: deploy-production
    when: manual
    steps:
      - name: deploy-to-production
        action: deploy
        params:
          environment: "production"
          services:
            - name: "client-vite"
              image: "${REGISTRY_URL}/client-vite:${TAG:-latest}"
            - name: "auth-service"
              image: "${REGISTRY_URL}/auth-service:${TAG:-latest}"
            - name: "api-gateway-service"
              image: "${REGISTRY_URL}/api-gateway-service:${TAG:-latest}"
            # Thêm các microservices khác tương tự
          config:
            consul_url: "http://consul:8500"
            vault_url: "http://vault:8200"
```

### Bước 4: Thực Thi Pipeline

Sử dụng API để thực thi pipeline:

```bash
curl -X POST "http://localhost:3000/pipelines/build" \
  -F "url=https://github.com/username/qlcv-refactor.git" \
  -F "branch=main" \
  -F "file=@qlcv-pipeline.yaml"
```

Hoặc sử dụng giao diện web:
1. Truy cập giao diện web PipeslicerCI
2. Chọn "Pipelines" > "Execute Pipeline"
3. Nhập thông tin:
   - Git Repository URL: `https://github.com/username/qlcv-refactor.git`
   - Branch: `main`
   - Upload file `qlcv-pipeline.yaml`
4. Nhấn "Execute"

### Bước 5: Giám Sát Quá Trình Build và Triển Khai

1. Theo dõi quá trình build trên giao diện web PipeslicerCI
2. Kiểm tra logs để phát hiện và xử lý lỗi
3. Xác nhận các Docker images đã được build và push lên registry

### Bước 6: Kiểm Tra Môi Trường Staging

1. Truy cập môi trường staging để kiểm tra ứng dụng
2. Thực hiện kiểm thử chức năng
3. Xác nhận tất cả các microservices đang hoạt động đúng

### Bước 7: Triển Khai lên Production

Sau khi xác nhận ứng dụng hoạt động tốt trên môi trường staging:

1. Truy cập giao diện web PipeslicerCI
2. Chọn "Pipelines" > "Current Executions"
3. Tìm pipeline đang chạy và chọn "Approve" cho stage "deploy-production"
4. Theo dõi quá trình triển khai lên production

## Ánh Xạ Chức Năng PipeslicerCI vào CI/CD Pipeline

### 1. Image Builder

**Chức năng**: Xây dựng Docker images từ mã nguồn
**Sử dụng trong pipeline**: 
- Stage "build" với các steps "build-frontend", "build-auth-service", "build-api-gateway"
- Tự động clone repository, build Docker image và push lên registry

### 2. Registry Manager

**Chức năng**: Quản lý metadata của Docker images
**Sử dụng trong pipeline**:
- Lưu trữ thông tin về các images đã build (service, tag, commit, branch, build time)
- Cho phép truy vấn lịch sử build
- Hỗ trợ tạo tags mới cho images hiện có

### 3. Registry Connector

**Chức năng**: Kết nối với Docker registry
**Sử dụng trong pipeline**:
- Xác thực với Docker registry
- Push images lên registry
- Liệt kê repositories và tags

### 4. Config Manager

**Chức năng**: Quản lý cấu hình triển khai
**Sử dụng trong pipeline**:
- Quản lý cấu hình cho các môi trường khác nhau (staging, production)
- Tích hợp với Consul và Vault để quản lý cấu hình và secrets

### 5. Pipeline Executor

**Chức năng**: Thực thi các pipeline CI/CD
**Sử dụng trong pipeline**:
- Phân tích và thực thi file pipeline YAML
- Quản lý các stages và steps
- Hỗ trợ điều kiện thực thi (when: manual)

## Lợi Ích của CI/CD Pipeline cho Dự Án QLCV

1. **Tự động hóa**: Giảm thiểu công việc thủ công trong quá trình build và triển khai
2. **Nhất quán**: Đảm bảo quy trình build và triển khai nhất quán giữa các môi trường
3. **Nhanh chóng**: Rút ngắn thời gian từ commit đến triển khai
4. **Chất lượng**: Phát hiện lỗi sớm thông qua tự động hóa kiểm thử
5. **Khả năng mở rộng**: Dễ dàng thêm microservices mới vào pipeline

## Kết Luận

PipeslicerCI cung cấp một giải pháp CI/CD toàn diện cho dự án QLCV, cho phép tự động hóa quy trình từ commit code đến triển khai sản phẩm. Bằng cách sử dụng các chức năng của PipeslicerCI, chúng ta có thể xây dựng một pipeline mạnh mẽ, linh hoạt và dễ quản lý cho dự án QLCV.

## Tài Liệu Tham Khảo

- [PipeslicerCI Documentation](https://github.com/vanhcao3/pipeslicerCI)
- [Docker Documentation](https://docs.docker.com/)
- [Microservices Architecture](https://microservices.io/)
