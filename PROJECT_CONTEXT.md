# PROJECT_CONTEXT.md - MangaHub Source of Truth (Week 4 Edition)

## 1. TỔNG QUAN DỰ ÁN VÀ TECH STACK
MangaHub là hệ thống quản lý và đồng bộ tiến độ đọc truyện đa giao thức. Hệ thống sử dụng kiến trúc hướng sự kiện (Event-driven) để đảm bảo trải nghiệm người dùng liền mạch trên cả Web và CLI.

### Tech Stack Chi Tiết:
- **Ngôn ngữ:** Go (Golang) 1.25+
- **Framework & Thư viện:**
    - `gin-gonic/gin`: HTTP Server & REST API.
    - `gorilla/websocket`: Xử lý kết nối thời gian thực.
    - `google.golang.org/grpc`: Giao thức nhị phân hiệu năng cao.
    - `spf13/cobra` (Planned): Xây dựng CLI Client chuyên nghiệp.
    - `glebarez/go-sqlite`: Database SQLite (Pure Go driver).
- **Hệ thống Port (Network Map):**

| Giao thức | Port | Vai trò |
| :--- | :--- | :--- |
| **HTTP** | 8080 | Auth, CRUD Manga, Admin Control |
| **TCP** | 9090 | Sync tiến độ đọc (Raw Socket) |
| **gRPC** | 50051 | Internal Service (Scan Manga Detail) |
| **UDP** | 9999 | Notifier Server (Broadcast cập nhật) |
| **Bridge** | 8888 | Cầu nối UDP-to-WebSocket |

## 2. CẤU TRÚC DỮ LIỆU (DATABASE SCHEMA)
Hệ thống sử dụng SQLite với các liên kết chặt chẽ để đảm bảo tính toàn vẹn dữ liệu.

### Chi tiết các bảng:
- **Bảng `users`:** Lưu trữ danh tính người dùng (PK: `id`).
- **Bảng `mangas`:** Lưu trữ kho truyện (PK: `id`).
- **Bảng `user_progress` (MỚI):**
    - `user_id` (INT): Foreign Key -> `users(id)`
    - `manga_id` (INT): Foreign Key -> `mangas(id)`
    - `last_chapter` (INT): Chương truyện cuối cùng đang đọc.
    - `updated_at` (DATETIME): Thời gian cập nhật gần nhất.
    - *Constraint:* Primary Key (user_id, manga_id).

## 3. TRẠNG THÁI TRIỂN KHAI (CURRENT STATUS)
Hệ thống đã hoàn thành các module cốt lõi và đang bước vào giai đoạn đóng gói.

| Module | Trạng thái | Ghi chú |
| :--- | :--- | :--- |
| **Auth (JWT)** | ✅ Hoàn thành | Đã có Register/Login & Token Verification. |
| **Manga CRUD** | ✅ Hoàn thành | Đã có API lấy danh sách và tìm kiếm. |
| **TCP Sync** | ✅ Hoàn thành | Hỗ trợ Sync đa thiết bị, có bảo vệ Mutex. |
| **gRPC Service** | ✅ Hoàn thành | Đã biên dịch `.proto` và có Server logic. |
| **Bridge Service** | ✅ Hoàn thành | **UDP -> Bridge -> WebSocket**: Thông báo chương mới thời gian thực. |

> [!IMPORTANT]
> **Luồng tích hợp đặc biệt:** Khi Admin cập nhật truyện (HTTP/gRPC) -> UDP Server bắn tin (Port 9999) -> Bridge (Port 8888) nhận tin và chuyển đổi sang JSON -> WebSocket đẩy trực tiếp xuống trình duyệt để hiện Toast Notification.

## 4. HẠN CHẾ VÀ LỘ TRÌNH TUẦN 4
### Hạn chế hiện tại (Known Limitations):
- **Frontend Decoupling:** Các file HTML (`auth.html`, `dashboard.html`, `reading.html`) hiện vẫn hoạt động rời rạc. Cần sử dụng `localStorage` để lưu JWT Token và điều hướng trạng thái người dùng.
- **CLI Connection:** CLI Client đã có mã nguồn tại `cmd/client` nhưng cần được kiểm thử khả năng chịu tải và kết nối đồng thời 5 protocol.

### Lộ trình Tuần 4 (Roadmap):
1. **Hoàn thiện CLI Client:** Tinh chỉnh lệnh và xử lý lỗi kết nối mạng (Network Resilience).
2. **Stress Test & Race Condition:** Sử dụng flag `-race` để rà soát xung đột Goroutines trong TCP Server và WebSocket Hub.
3. **Đóng gói (Deployment):**
    - Viết **Dockerfile** tối ưu cho Go.
    - Cấu hình **docker-compose.yml** để chạy toàn bộ stack (Server + DB + Web) chỉ với một lệnh.

## 5. NGUYÊN TẮC BMAD
1. **Defensive Programming:** Xử lý lỗi tuyệt đối tại mọi điểm I/O.
2. **Layer-slicing:** Tôn trọng ranh giới giữa Domain logic và Protocol layer.
3. **No Guessing:** Chỉ sửa lỗi khi đã xác định được Root-Cause qua Log.
