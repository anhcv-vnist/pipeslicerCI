# TÀI LIỆU PHÂN TÍCH YÊU CẦU
## CÔNG CỤ PIPESLICERCI CHO DỰ ÁN QLCV MICROSERVICES

**Mã tài liệu:** PRD-PIPESLICERCI-001  
**Phiên bản:** 1.0  
**Ngày tạo:** 03/04/2025  
**Người tạo:** Cline  
**Trạng thái:** Dự thảo

---

## 1. TỔNG QUAN

### 1.1. Mục đích tài liệu
Tài liệu này mô tả chi tiết các yêu cầu chức năng và phi chức năng cho việc mở rộng công cụ PipeslicerCI nhằm hỗ trợ quy trình DevOps cho dự án QLCV Microservices. Tài liệu sẽ làm rõ các vấn đề hiện tại, đề xuất giải pháp, và mô tả chi tiết các use case cho từng chức năng.

### 1.2. Phạm vi dự án
Dự án bao gồm việc phát triển và mở rộng công cụ PipeslicerCI hiện có để cung cấp một giải pháp DevOps toàn diện, tập trung vào các vấn đề sau:
- Tự động hóa quy trình build và đóng gói Docker image
- Quản lý phiên bản và lưu trữ Docker image
- Tự động hóa kiểm thử
- Quản lý cấu hình tập trung
- Quản lý bí mật (secrets)
- Triển khai tự động với khả năng phân biệt môi trường
- Phân tích phụ thuộc giữa các service

### 1.3. Định nghĩa và từ viết tắt
- **CI/CD**: Continuous Integration/Continuous Deployment
- **QLCV**: Quản lý công việc (tên dự án)
- **DevOps**: Development and Operations
- **API**: Application Programming Interface
- **UI**: User Interface
- **YAML**: Yet Another Markup Language

---

## 2. PHÂN TÍCH HIỆN TRẠNG

### 2.1. Mô tả hệ thống hiện tại
PipeslicerCI hiện tại là một công cụ CI đơn giản được viết bằng Golang, cung cấp API để thực thi các pipeline được định nghĩa trong file YAML. Công cụ có khả năng clone repository từ Git và thực thi các lệnh được định nghĩa trong pipeline.

Kiến trúc hiện tại bao gồm:
- Web server sử dụng Fiber framework
- API endpoint `/pipelines/build` để thực thi pipeline
- Các module core: workspace, pipeline, executor

### 2.2. Vấn đề tồn tại
Dự án QLCV Microservices đang trong quá trình chuyển đổi từ kiến trúc monolithic sang microservices và gặp phải các vấn đề sau:

1. **Build trực tiếp trong container**:
   - **Hiện trạng**: Mỗi service sử dụng base image và build trực tiếp trong container khi chạy docker-compose up
   - **Tác động**: Thời gian khởi động hệ thống kéo dài, tài nguyên máy chủ bị lãng phí, khó kiểm soát phiên bản

2. **Thiếu quản lý phiên bản Docker image**:
   - **Hiện trạng**: Không có quy trình đóng gói và đẩy image lên registry
   - **Tác động**: Khó khăn trong việc rollback, không đảm bảo tính nhất quán giữa các môi trường

3. **Thiếu cơ chế kiểm thử tự động**:
   - **Hiện trạng**: Không có bước test tự động trước khi build và triển khai
   - **Tác động**: Rủi ro cao về lỗi khi triển khai, phát hiện lỗi muộn

4. **Quản lý biến môi trường phân tán**:
   - **Hiện trạng**: Mỗi service có file .env riêng, được tạo từ .env.example thông qua script
   - **Tác động**: Khó đồng bộ cấu hình giữa các môi trường, dễ gây lỗi khi thay đổi cấu hình

5. **Thiếu cơ chế quản lý bí mật**:
   - **Hiện trạng**: Các thông tin nhạy cảm được lưu trực tiếp trong file .env
   - **Tác động**: Rủi ro bảo mật cao, khó quản lý quyền truy cập vào thông tin nhạy cảm

6. **Thiếu cơ chế phân biệt cấu hình theo môi trường**:
   - **Hiện trạng**: Không có cơ chế rõ ràng để phân biệt cấu hình giữa các môi trường
   - **Tác động**: Khó khăn khi triển khai trên nhiều môi trường, dễ xảy ra lỗi cấu hình

7. **Thiếu tính module hóa trong quy trình CI/CD**:
   - **Hiện trạng**: Các script triển khai hiện tại xử lý tất cả các service cùng một lúc
   - **Tác động**: Không thể triển khai độc lập từng service, tăng rủi ro khi triển khai

8. **Thiếu cơ chế phát hiện sự phụ thuộc giữa các service**:
   - **Hiện trạng**: Không có công cụ phân tích sự phụ thuộc giữa các service
   - **Tác động**: Khó xác định thứ tự triển khai, dễ gây lỗi khi một service phụ thuộc chưa sẵn sàng

---

## 3. YÊU CẦU CHỨC NĂNG

### 3.1. Image Builder Service

#### 3.1.1. Mô tả chức năng
Dịch vụ tự động xây dựng Docker image từ mã nguồn và lưu trữ trong registry nội bộ.

#### 3.1.2. Đầu vào
- Repository URL
- Branch/Tag
- Service path
- Build configuration

#### 3.1.3. Đầu ra
- Docker image được build và push lên registry
- Build logs
- Build status (success/failure)
- Build metadata (commit hash, timestamp, etc.)

#### 3.1.4. Use Case Scenario
**Tên use case**: Build và Push Docker Image  
**Tác nhân chính**: Developer, CI System  
**Điều kiện tiên quyết**: Repository đã được cấu hình với Dockerfile  
**Luồng chính**:
1. Developer push code lên repository
2. CI System phát hiện thay đổi và trigger build
3. Image Builder Service clone repository
4. Image Builder Service xác định các service bị ảnh hưởng bởi thay đổi
5. Image Builder Service build Docker image cho từng service
6. Image Builder Service tag và push image lên registry nội bộ
7. Image Builder Service thông báo kết quả build

**Luồng thay thế**:
- Nếu build thất bại, Image Builder Service gửi thông báo lỗi và không push image lên registry

#### 3.1.5. Mức độ tự động hóa
Hoàn toàn tự động

### 3.2. Registry Manager

#### 3.2.1. Mô tả chức năng
Quản lý phiên bản image, lưu trữ metadata và cung cấp API để truy vấn.

#### 3.2.2. Đầu vào
- Image metadata (service name, tag, commit hash, etc.)
- Query parameters (service, version, environment, etc.)

#### 3.2.3. Đầu ra
- Image metadata
- Image history
- Latest image information

#### 3.2.4. Use Case Scenario
**Tên use case**: Quản lý Phiên bản Docker Image  
**Tác nhân chính**: Developer, Ops, Deployment System  
**Điều kiện tiên quyết**: Image đã được build và push lên registry  
**Luồng chính**:
1. Image Builder Service push image lên registry
2. Registry Manager lưu trữ metadata (commit hash, branch, build time, etc.)
3. Deployment System truy vấn Registry Manager để lấy image phù hợp
4. Registry Manager cung cấp API để rollback về phiên bản trước
5. Ops có thể xem lịch sử các phiên bản image

**Luồng thay thế**:
- Nếu không tìm thấy image phù hợp, Registry Manager trả về thông báo lỗi

#### 3.2.5. Mức độ tự động hóa
Hoàn toàn tự động

### 3.3. Test Runner Service

#### 3.3.1. Mô tả chức năng
Dịch vụ tự động chạy các bộ test và báo cáo kết quả.

#### 3.3.2. Đầu vào
- Repository URL
- Branch/Tag
- Service path
- Test configuration

#### 3.3.3. Đầu ra
- Test results (pass/fail)
- Test coverage
- Test logs

#### 3.3.4. Use Case Scenario
**Tên use case**: Chạy Automated Tests  
**Tác nhân chính**: Developer, CI System  
**Điều kiện tiên quyết**: Repository đã được cấu hình với test scripts  
**Luồng chính**:
1. Developer push code lên repository
2. CI System phát hiện thay đổi và trigger test
3. Test Runner Service clone repository
4. Test Runner Service xác định các service bị ảnh hưởng bởi thay đổi
5. Test Runner Service chạy unit tests, integration tests, và end-to-end tests
6. Test Runner Service tạo báo cáo kết quả test
7. Test Runner Service thông báo kết quả test

**Luồng thay thế**:
- Nếu tests thất bại, Test Runner Service gửi thông báo lỗi và dừng quy trình CI/CD

#### 3.3.5. Mức độ tự động hóa
Hoàn toàn tự động

### 3.4. Configuration Manager

#### 3.4.1. Mô tả chức năng
Quản lý cấu hình tập trung cho tất cả các service.

#### 3.4.2. Đầu vào
- Configuration key-value pairs
- Environment (dev, staging, production)
- Service name

#### 3.4.3. Đầu ra
- Configuration values
- Configuration history
- Configuration diff between environments

#### 3.4.4. Use Case Scenario
**Tên use case**: Quản lý Cấu hình Tập trung  
**Tác nhân chính**: Developer, Ops, Services  
**Điều kiện tiên quyết**: Configuration Manager đã được cài đặt và cấu hình  
**Luồng chính**:
1. Ops định nghĩa cấu hình cho mỗi service trong mỗi môi trường
2. Configuration Manager lưu trữ cấu hình trong database
3. Services truy vấn Configuration Manager để lấy cấu hình khi khởi động
4. Developer có thể xem và cập nhật cấu hình thông qua UI
5. Configuration Manager theo dõi lịch sử thay đổi cấu hình

**Luồng thay thế**:
- Nếu service không thể kết nối với Configuration Manager, service sử dụng cấu hình mặc định

#### 3.4.5. Mức độ tự động hóa
Một phần tự động (cần sự tham gia của Ops để định nghĩa cấu hình)

### 3.5. Secrets Manager

#### 3.5.1. Mô tả chức năng
Quản lý bí mật (secrets) an toàn cho tất cả các service.

#### 3.5.2. Đầu vào
- Secret key-value pairs
- Access policies
- Service name

#### 3.5.3. Đầu ra
- Encrypted secrets
- Access logs
- Secret rotation status

#### 3.5.4. Use Case Scenario
**Tên use case**: Quản lý Bí mật An toàn  
**Tác nhân chính**: Developer, Ops, Services  
**Điều kiện tiên quyết**: Secrets Manager đã được cài đặt và cấu hình  
**Luồng chính**:
1. Ops định nghĩa bí mật cho mỗi service
2. Secrets Manager mã hóa và lưu trữ bí mật
3. Services xác thực với Secrets Manager và lấy bí mật khi cần
4. Secrets Manager ghi log mỗi lần truy cập bí mật
5. Ops có thể xoay vòng (rotate) bí mật định kỳ

**Luồng thay thế**:
- Nếu service không có quyền truy cập bí mật, Secrets Manager trả về lỗi

#### 3.5.5. Mức độ tự động hóa
Một phần tự động (cần sự tham gia của Ops để định nghĩa bí mật)

### 3.6. Deployment Orchestrator

#### 3.6.1. Mô tả chức năng
Quản lý quy trình triển khai các service, đảm bảo thứ tự triển khai đúng.

#### 3.6.2. Đầu vào
- Deployment configuration
- Environment (dev, staging, production)
- Service dependencies

#### 3.6.3. Đầu ra
- Deployment status
- Deployment logs
- Service health status

#### 3.6.4. Use Case Scenario
**Tên use case**: Triển khai Microservices Theo Thứ tự Phụ thuộc  
**Tác nhân chính**: Ops, CI System  
**Điều kiện tiên quyết**: Services đã được build và push lên registry  
**Luồng chính**:
1. CI System trigger deployment sau khi build và test thành công
2. Deployment Orchestrator phân tích dependency graph
3. Deployment Orchestrator xác định thứ tự triển khai tối ưu
4. Deployment Orchestrator triển khai từng service theo thứ tự
5. Deployment Orchestrator kiểm tra health check sau mỗi lần triển khai
6. Deployment Orchestrator thông báo kết quả triển khai

**Luồng thay thế**:
- Nếu một service triển khai thất bại, Deployment Orchestrator có thể rollback hoặc dừng quy trình

#### 3.6.5. Mức độ tự động hóa
Hoàn toàn tự động cho môi trường dev và staging, một phần tự động cho production (cần approval)

### 3.7. Dependency Analyzer

#### 3.7.1. Mô tả chức năng
Phân tích và quản lý sự phụ thuộc giữa các service.

#### 3.7.2. Đầu vào
- Service definitions
- API contracts
- Network traffic data

#### 3.7.3. Đầu ra
- Dependency graph
- Impact analysis
- Deployment order recommendations

#### 3.7.4. Use Case Scenario
**Tên use case**: Phân tích Phụ thuộc Giữa Các Service  
**Tác nhân chính**: Developer, Ops, Deployment Orchestrator  
**Điều kiện tiên quyết**: Services đã được định nghĩa với API contracts  
**Luồng chính**:
1. Developer định nghĩa API contracts cho mỗi service
2. Dependency Analyzer quét mã nguồn và API contracts
3. Dependency Analyzer xây dựng dependency graph
4. Deployment Orchestrator sử dụng dependency graph để xác định thứ tự triển khai
5. Developer có thể xem dependency graph để hiểu tác động của thay đổi

**Luồng thay thế**:
- Nếu phát hiện circular dependency, Dependency Analyzer cảnh báo và đề xuất giải pháp

#### 3.7.5. Mức độ tự động hóa
Hoàn toàn tự động

### 3.8. Monitoring and Alerting System

#### 3.8.1. Mô tả chức năng
Giám sát sức khỏe và hiệu suất của các service, gửi cảnh báo khi phát hiện vấn đề.

#### 3.8.2. Đầu vào
- Service metrics
- Logs
- Health check status

#### 3.8.3. Đầu ra
- Dashboards
- Alerts
- Performance reports

#### 3.8.4. Use Case Scenario
**Tên use case**: Giám sát và Cảnh báo Hệ thống  
**Tác nhân chính**: Ops, Services  
**Điều kiện tiên quyết**: Services đã được cấu hình để export metrics  
**Luồng chính**:
1. Services export metrics và logs
2. Monitoring System thu thập và lưu trữ metrics và logs
3. Monitoring System phân tích metrics và logs để phát hiện vấn đề
4. Monitoring System gửi cảnh báo khi phát hiện vấn đề
5. Ops xem dashboards để theo dõi sức khỏe và hiệu suất của hệ thống

**Luồng thay thế**:
- Nếu một service không export metrics, Monitoring System gửi cảnh báo về việc thiếu dữ liệu

#### 3.8.5. Mức độ tự động hóa
Hoàn toàn tự động

---

## 4. YÊU CẦU PHI CHỨC NĂNG

### 4.1. Hiệu năng
- Thời gian build không quá 10 phút cho mỗi service
- Thời gian triển khai không quá 5 phút cho mỗi service
- Hệ thống phải xử lý được ít nhất 10 build đồng thời

### 4.2. Bảo mật
- Tất cả các bí mật phải được mã hóa khi lưu trữ
- Tất cả các API phải yêu cầu xác thực
- Tất cả các kết nối phải sử dụng TLS
- Phải ghi log tất cả các hoạt động liên quan đến bảo mật

### 4.3. Khả năng mở rộng
- Hệ thống phải hỗ trợ ít nhất 50 microservices
- Hệ thống phải hỗ trợ ít nhất 3 môi trường (dev, staging, production)
- Kiến trúc phải cho phép thêm các plugin mới dễ dàng

### 4.4. Độ tin cậy
- Hệ thống phải có tính sẵn sàng cao (high availability)
- Hệ thống phải có khả năng tự phục hồi sau lỗi
- Hệ thống phải có cơ chế backup và restore

### 4.5. Khả năng sử dụng
- UI phải trực quan và dễ sử dụng
- Hệ thống phải cung cấp API documentation đầy đủ
- Hệ thống phải cung cấp logs và metrics dễ đọc

---

## 5. KIẾN TRÚC HỆ THỐNG

### 5.1. Tổng quan kiến trúc
PipeslicerCI sẽ được mở rộng thành một hệ thống microservices với các thành phần sau:

1. **Core Services**:
   - API Gateway
   - Authentication Service
   - Web UI

2. **CI/CD Services**:
   - Image Builder Service
   - Test Runner Service
   - Deployment Orchestrator

3. **Management Services**:
   - Registry Manager
   - Configuration Manager
   - Secrets Manager
   - Dependency Analyzer

4. **Infrastructure Services**:
   - Monitoring and Alerting System
   - Logging Service
   - Database Service

### 5.2. Công nghệ sử dụng
- **Backend**: Golang
- **Frontend**: React
- **Database**: PostgreSQL, Redis
- **Message Queue**: RabbitMQ
- **Container**: Docker
- **Orchestration**: Kubernetes
- **CI/CD**: PipeslicerCI (self-hosted)

### 5.3. Mô hình triển khai
Hệ thống sẽ được triển khai trên Kubernetes với các thành phần sau:
- Namespace cho mỗi môi trường (dev, staging, production)
- StatefulSet cho các database và message queue
- Deployment cho các service
- Service và Ingress cho network routing
- ConfigMap và Secret cho cấu hình

---

## 6. KẾ HOẠCH TRIỂN KHAI

### 6.1. Lộ trình phát triển
1. **Phase 1 (2 tuần)**: Core Services
   - API Gateway
   - Authentication Service
   - Web UI (basic)

2. **Phase 2 (3 tuần)**: CI/CD Services
   - Image Builder Service
   - Test Runner Service
   - Registry Manager

3. **Phase 3 (3 tuần)**: Management Services
   - Configuration Manager
   - Secrets Manager
   - Dependency Analyzer

4. **Phase 4 (2 tuần)**: Deployment và Integration
   - Deployment Orchestrator
   - Integration với các service khác
   - End-to-end testing

5. **Phase 5 (2 tuần)**: Monitoring và Optimization
   - Monitoring and Alerting System
   - Performance optimization
   - Documentation

### 6.2. Rủi ro và giảm thiểu
1. **Rủi ro**: Độ phức tạp của hệ thống microservices
   **Giảm thiểu**: Phát triển từng phase, testing kỹ lưỡng sau mỗi phase

2. **Rủi ro**: Khó khăn trong việc migration từ hệ thống cũ
   **Giảm thiểu**: Cung cấp công cụ migration và hỗ trợ song song cả hai hệ thống trong thời gian chuyển đổi

3. **Rủi ro**: Resistance to change từ team
   **Giảm thiểu**: Training và documentation đầy đủ, demo các lợi ích của hệ thống mới

---

## 7. PHỤ LỤC

### 7.1. Glossary
- **Pipeline**: Một chuỗi các bước (steps) được thực thi tuần tự để build, test và deploy một ứng dụng
- **Registry**: Kho lưu trữ Docker images
- **Microservice**: Một service nhỏ, độc lập, thực hiện một chức năng cụ thể trong hệ thống
- **Dependency Graph**: Biểu đồ thể hiện sự phụ thuộc giữa các service
- **Health Check**: Kiểm tra trạng thái hoạt động của một service

### 7.2. References
- Docker Documentation: https://docs.docker.com/
- Kubernetes Documentation: https://kubernetes.io/docs/
- Go Documentation: https://golang.org/doc/
- Microservices Architecture: https://microservices.io/

---

**Phê duyệt bởi**:

| Vai trò | Tên | Ngày phê duyệt |
|---------|-----|----------------|
| Product Owner | | |
| Technical Lead | | |
| DevOps Lead | | |
| Security Officer | | |
