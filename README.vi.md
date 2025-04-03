# PipeslicerCI

PipeslicerCI là một công cụ CI/CD tùy chỉnh được thiết kế đặc biệt cho kiến trúc microservices. Nó cung cấp một bộ tính năng toàn diện để giải quyết các thách thức phổ biến trong phát triển, kiểm thử và triển khai microservices.

## Tính năng

### 1. Dịch vụ Image Builder

Dịch vụ Image Builder tự động hóa quá trình xây dựng Docker image từ mã nguồn và đẩy chúng lên registry. Điều này loại bỏ nhu cầu xây dựng image trực tiếp trong container trong quá trình triển khai, dẫn đến thời gian khởi động nhanh hơn và sử dụng tài nguyên hiệu quả hơn.

**Khả năng chính:**
- Xây dựng Docker image từ các repository Git
- Đẩy image lên các Docker registry
- Phát hiện các service đã thay đổi và cần xây dựng lại
- Hỗ trợ multi-stage build để tối ưu hóa kích thước image

### 2. Registry Manager

Registry Manager cung cấp một hệ thống tập trung để quản lý metadata của Docker image. Nó theo dõi tất cả các image, phiên bản và metadata liên quan, giúp dễ dàng tìm đúng image để triển khai và quay lại các phiên bản trước nếu cần.

**Khả năng chính:**
- Lưu trữ metadata về các image đã xây dựng
- Truy vấn image theo service, tag hoặc commit
- Theo dõi lịch sử image
- Hỗ trợ gắn tag cho image

### 3. Registry Connector

Registry Connector cung cấp một giao diện thống nhất để tương tác với các Docker registry khác nhau, bao gồm Docker Hub và Harbor. Nó xử lý việc xác thực, đẩy image và quản lý các repository và tag.

**Khả năng chính:**
- Kết nối với Docker Hub, Harbor hoặc bất kỳ registry tương thích Docker nào
- Xác thực với các registry
- Đẩy image lên registry
- Liệt kê các repository và tag
- Xóa tag

### 4. Configuration Manager

Configuration Manager cung cấp quản lý cấu hình tập trung cho tất cả các service. Nó cho phép bạn lưu trữ và quản lý các giá trị cấu hình ở một vị trí trung tâm, giúp dễ dàng duy trì cấu hình nhất quán giữa các môi trường và cập nhật giá trị cấu hình mà không cần xây dựng lại image.

**Khả năng chính:**
- Lưu trữ giá trị cấu hình cho các service và môi trường khác nhau
- Quản lý thông tin nhạy cảm một cách an toàn
- Tạo file .env cho các service
- Nhập và xuất cấu hình dưới định dạng JSON

### 5. Dependency Analyzer

Dependency Analyzer giúp bạn hiểu các phụ thuộc giữa các service, giúp dễ dàng xác định thứ tự triển khai chính xác và hiểu tác động của các thay đổi.

**Khả năng chính:**
- Phân tích phụ thuộc của service
- Tạo biểu đồ phụ thuộc
- Đề xuất thứ tự triển khai
- Xác định phụ thuộc vòng tròn

## Bắt đầu

### Yêu cầu tiên quyết

- Go 1.16 trở lên
- Docker
- Git

### Cài đặt

1. Clone repository:

```bash
git clone https://github.com/vanhcao3/pipeslicerCI.git
cd pipeslicerCI
```

2. Xây dựng ứng dụng:

```bash
go build -o pipeslicerci ./cmd/web
```

3. Chạy ứng dụng:

```bash
./pipeslicerci
```

Máy chủ sẽ khởi động trên cổng 3000 theo mặc định.

## Tài liệu API

PipeslicerCI cung cấp tài liệu API toàn diện sử dụng Swagger/OpenAPI. Bạn có thể truy cập tài liệu API bằng cách điều hướng đến endpoint `/swagger` trong trình duyệt của bạn sau khi khởi động máy chủ.

Ví dụ, nếu máy chủ đang chạy trên localhost cổng 3000, bạn có thể truy cập tài liệu API tại:

```
http://localhost:3000/swagger
```

Tài liệu API cung cấp thông tin chi tiết về tất cả các endpoint có sẵn, bao gồm:

- Tham số yêu cầu
- Lược đồ nội dung yêu cầu
- Lược đồ phản hồi
- Ví dụ về yêu cầu và phản hồi

### Các API có sẵn

PipeslicerCI cung cấp các nhóm API sau:

1. **API Image Builder** - Để xây dựng và đẩy Docker image
2. **API Registry** - Để quản lý metadata của Docker image
3. **API Registry Connector** - Để tương tác với các Docker registry
4. **API Configuration** - Để quản lý giá trị cấu hình
5. **API Pipeline** - Để thực thi pipeline

## Ví dụ sử dụng

### Xây dựng một Docker Image

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

### Đẩy một Image lên Docker Hub

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

### Đẩy một Image lên Harbor

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

### Đặt một giá trị cấu hình

```bash
curl -X POST http://localhost:3000/config/services/auth-service/environments/development/values \
  -H "Content-Type: application/json" \
  -d '{
    "key": "DB_HOST",
    "value": "localhost",
    "isSecret": false
  }'
```

### Tạo một file .env

```bash
curl -X GET http://localhost:3000/config/services/auth-service/environments/development/env \
  -H "Authorization: Bearer mytoken"
```

## Đóng góp

Đóng góp luôn được chào đón! Vui lòng cảm thấy tự do gửi Pull Request.

## Giấy phép

Dự án này được cấp phép theo Giấy phép MIT - xem file LICENSE để biết chi tiết.
