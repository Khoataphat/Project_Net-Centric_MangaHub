# story_week4_release_demo.md

# PRD — MangaHub Release, Packaging, Testing & Multi-Machine Demo

## 1. Overview (Tổng quan)

### Mục tiêu giai đoạn Release & Demo

Giai đoạn này tập trung vào việc hoàn thiện quy trình đóng gói, kiểm thử tương tranh, triển khai demo đa máy và chuẩn hóa tài liệu báo cáo cho hệ thống MangaHub.

Mục tiêu cuối cùng là biến hệ thống từ trạng thái “development modules” sang trạng thái “portable demo-ready distributed system”, cho phép:

* Khởi động toàn bộ backend bằng một lệnh duy nhất thông qua Docker Compose.
* Đảm bảo hệ thống ổn định khi xử lý nhiều kết nối đồng thời bằng cơ chế kiểm thử Race Detection của Golang.
* Trình diễn hệ thống hoạt động thực tế trên hai máy tính khác nhau trong cùng mạng LAN.
* Hoàn thiện bộ tài liệu phục vụ nghiệm thu môn học và báo cáo cuối kỳ.

---

### Phạm vi triển khai

Khâu Release & Demo bao gồm:

* Docker Packaging
* Multi-stage Build Optimization
* SQLite Persistent Volume Mapping
* Port Exposure & Network Binding
* Concurrency & Race-condition Testing
* Multi-machine LAN Deployment
* Demo Scenario Documentation
* Final Academic Report Packaging

---

### Các giao thức cần hoạt động sau khi đóng gói

Hệ thống phải hoạt động đầy đủ với các protocol sau:

| Protocol            | Port  | Mục đích                          |
| ------------------- | ----- | --------------------------------- |
| HTTP REST API       | 8080  | Auth + CRUD Manga                 |
| TCP Socket          | 9090  | Progress Sync                     |
| UDP Notifier        | 9999  | Push thông báo                    |
| UDP Bridge Listener | 8888  | Forward UDP → WebSocket           |
| gRPC                | 50051 | Internal high-performance service |

---

## 2. Acceptance Criteria (Tiêu chí nghiệm thu)

---

## [AC1] Docker Infrastructure

### Mục tiêu

Toàn bộ backend MangaHub phải chạy được bằng Docker Compose với một lệnh duy nhất.

### Điều kiện đạt

* Có file:

  * `Dockerfile`
  * `docker-compose.yml`

* Docker Compose phải:

  * Build image thành công từ source code Go.
  * Chạy container không crash.
  * Expose đầy đủ 5 ports:

| Service      | Port     |
| ------------ | -------- |
| HTTP         | 8080     |
| TCP          | 9090     |
| UDP Notifier | 9999/udp |
| UDP Bridge   | 8888/udp |
| gRPC         | 50051    |

* SQLite database phải được mount qua Docker Volume:

  * Không mất dữ liệu sau khi restart container.
  * File `sqlite.db` được persist bên ngoài container filesystem.

### Điều kiện kiểm thử

* Chạy:

```bash
docker compose up --build
```

* Kiểm tra:

```bash
docker ps
```

* Kiểm tra API:

```bash
curl http://localhost:8080/health
```

* Restart container:

```bash
docker compose down
docker compose up
```

* Xác nhận dữ liệu SQLite vẫn còn.

---

## [AC2] Concurrency Testing

### Mục tiêu

Đảm bảo hệ thống không phát sinh Race Condition khi xử lý đồng thời nhiều goroutine hoặc nhiều kết nối mạng.

### Điều kiện đạt

Phải có:

* Script test tải đồng thời
  hoặc:
* Hướng dẫn chạy kiểm thử race detector.

### Luồng kiểm thử tối thiểu

```bash
go test -race ./...
```

hoặc:

```bash
go run -race main.go
```

### Kịch bản kiểm thử bắt buộc

Hệ thống phải được giả lập:

* Nhiều HTTP requests đồng thời.
* Nhiều TCP clients sync progress cùng lúc.
* UDP notification burst.
* WebSocket clients subscribe đồng thời.
* gRPC concurrent requests.

### Tiêu chí PASS

* Không xuất hiện log:

```text
WARNING: DATA RACE
```

* Không deadlock.
* Không panic do concurrent map access.
* Không memory corruption.

### Tiêu chí FAIL

* Có log DATA RACE.
* Goroutine leak.
* Concurrent write vào SQLite gây crash.
* Shared state không có mutex/channel protection.

---

## [AC3] Multi-Machine Demo

### Mục tiêu

Trình diễn hệ thống MangaHub hoạt động thực tế trên hai máy tính khác nhau trong cùng mạng LAN/WiFi.

---

### Kiến trúc Demo

#### Máy A — Server

* Chạy Docker Compose.
* Chứa toàn bộ backend MangaHub.
* Expose 5 ports ra mạng LAN.

#### Máy B — Client

* Chạy:

  * Browser frontend
  * CLI Client
  * TCP test client
  * gRPC client
  * UDP sender/listener

---

### Điều kiện đạt

Phải có tài liệu:

```text
Demo_Script.md
```

Bao gồm đầy đủ:

---

### Step 1 — Xác định IP LAN Máy A

Ví dụ:

```bash
ipconfig
```

Hoặc:

```bash
ifconfig
```

Kết quả:

```text
192.168.x.x
```

---

### Step 2 — Binding Server về 0.0.0.0

Server không được bind:

```text
localhost
127.0.0.1
```

Mà phải bind:

```text
0.0.0.0
```

để máy khác truy cập được.

---

### Step 3 — Docker Startup

```bash
docker compose up --build
```

---

### Step 4 — Kiểm thử HTTP từ Máy B

```bash
curl http://<SERVER_IP>:8080/mangas
```

---

### Step 5 — Kiểm thử TCP Sync

CLI Client từ Máy B kết nối:

```text
<SERVER_IP>:9090
```

---

### Step 6 — Kiểm thử UDP Notification

Máy B gửi hoặc nhận packet:

```text
<SERVER_IP>:9999
```

---

### Step 7 — Kiểm thử UDP Bridge

Bridge listener nhận packet:

```text
<SERVER_IP>:8888
```

---

### Step 8 — Kiểm thử gRPC

gRPC client từ Máy B gọi:

```text
<SERVER_IP>:50051
```

---

### Step 9 — Demo End-to-End Flow

Luồng demo hoàn chỉnh:

1. Add Manga qua HTTP.
2. Trigger Event.
3. UDP gửi notification.
4. Bridge nhận packet.
5. WebSocket push realtime lên frontend.
6. CLI/TCP sync tiến độ đọc.
7. gRPC trả metadata tốc độ cao.

---

### Tiêu chí PASS

* Hai máy kết nối được qua LAN.
* Không bị firewall block.
* Tất cả protocol hoạt động xuyên máy.
* Không hardcode localhost trong client.

---

## [AC4] Documentation

### Mục tiêu

Hoàn thiện bộ báo cáo phục vụ nghiệm thu học phần.

### Điều kiện đạt

Phải có:

* Report Word/PDF.
* Cấu trúc đúng template giảng viên.
* Nội dung đồng bộ với hệ thống thực tế.

---

### Nội dung tối thiểu của Report

| Section                      | Nội dung                    |
| ---------------------------- | --------------------------- |
| Introduction                 | Giới thiệu đề tài           |
| System Architecture          | Kiến trúc đa giao thức      |
| Tech Stack                   | Go, Docker, SQLite          |
| Protocol Design              | HTTP/TCP/UDP/gRPC/WebSocket |
| Database Design              | SQLite schema               |
| Concurrency Handling         | Goroutine + Mutex           |
| Docker Deployment            | Compose + Volume            |
| Testing Strategy             | Race test + Stress test     |
| Multi-machine Demo           | LAN deployment              |
| Challenges & Lessons Learned | Technical debt + fixes      |
| Conclusion                   | Tổng kết                    |

---

## 3. Technical Constraints & Defense (Ràng buộc kỹ thuật)

---

## Multi-stage Docker Build (Bắt buộc)

### Mục tiêu

Giảm kích thước image cuối cùng.

### Yêu cầu

Dockerfile phải chia làm:

* Builder Stage
* Runtime Stage

### Runtime image chỉ được chứa:

* Binary Go đã compile
* Thư mục:

  * `/templates`
  * `/database`
  * config cần thiết

### Không được chứa:

* Source code dư thừa
* Git metadata
* Cache module không cần thiết

---

## Persistent SQLite Volume

### Yêu cầu

SQLite database phải được mount ra ngoài container.

Ví dụ:

```yaml
volumes:
  - ./data:/app/data
```

### Mục tiêu

Tránh mất:

* User accounts
* Manga data
* Reading progress

khi:

```bash
docker compose down
```

---

## Network Binding Defense

### Yêu cầu

Server phải hỗ trợ ENV:

```env
HOST=0.0.0.0
```

### Không được hardcode:

```text
localhost
127.0.0.1
```

### Lý do

Cho phép:

* Docker networking
* LAN demo
* Multi-machine access

---

## Defensive Runtime Rules

### Tất cả service phải:

* Có timeout.
* Có graceful shutdown.
* Có log lỗi rõ ràng.
* Không panic khi port unavailable.
* Retry hợp lý khi bridge/service chưa sẵn sàng.

---

## Port Exposure Rules

### Docker Compose phải khai báo rõ:

```yaml
ports:
  - "8080:8080"
  - "9090:9090"
  - "9999:9999/udp"
  - "8888:8888/udp"
  - "50051:50051"
```

---

## Race-condition Defense

### Shared Resources cần được bảo vệ:

* In-memory session map
* TCP client registry
* WebSocket connection hub
* UDP subscriber list
* SQLite concurrent access

### Cơ chế khuyến nghị

* Mutex
* RWMutex
* Channel ownership
* Worker queue
* Context cancellation

---

## 4. Task List (Danh sách công việc cho Người B)

### Infrastructure & Packaging

* [ ] Viết `Dockerfile`.
* [ ] Viết `docker-compose.yml`.
* [ ] Thiết lập multi-stage build.
* [ ] Mapping SQLite volume.
* [ ] Verify expose đủ 5 ports.

---

### Concurrency & Stress Testing

* [ ] Viết script giả lập tải đồng thời.
* [ ] Chạy:

```bash
go test -race ./...
```

* [ ] Chạy:

```bash
go run -race main.go
```

* [ ] Log và phân tích race-condition nếu xuất hiện.
* [ ] Fix shared-state synchronization.

---

### Multi-Machine Demo

* [ ] Viết `Demo_Script.md`.
* [ ] Kiểm thử LAN giữa 2 máy thật.
* [ ] Kiểm tra firewall/network access.
* [ ] Demo đủ:

  * HTTP
  * TCP
  * UDP
  * UDP Bridge
  * gRPC

---

### Documentation & Finalization

* [ ] Hoàn thiện Report.
* [ ] Đồng bộ screenshot hệ thống thực tế.
* [ ] Bổ sung sơ đồ kiến trúc.
* [ ] Mô tả luồng Event Bus tương lai.
* [ ] Chuẩn hóa format theo template giảng viên.
* [ ] Export:

  * `.docx`
  * `.pdf`

---

# Final Deliverables

| File                    | Mục tiêu                |
| ----------------------- | ----------------------- |
| Dockerfile              | Build production image  |
| docker-compose.yml      | Deploy toàn hệ thống    |
| Demo_Script.md          | Hướng dẫn demo 2 máy    |
| Stress_Test.md / script | Kiểm thử race-condition |
| Report.docx/pdf         | Báo cáo cuối kỳ         |

---

# Expected Outcome

Sau khi hoàn thành giai đoạn này, MangaHub phải đạt trạng thái:

* Portable
* Reproducible
* Demo-ready
* Concurrent-safe
* Multi-machine deployable
* Ready for academic presentation & defense.
