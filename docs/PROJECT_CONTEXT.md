# PROJECT_CONTEXT.md - MangaHub Source of Truth

## 1. TỔNG QUAN DỰ ÁN VÀ TECH STACK
- **Mục tiêu:** MangaHub là hệ thống quản lý và đồng bộ tiến độ đọc truyện đa giao thức, cho phép người dùng theo dõi và cập nhật trạng thái đọc mượt mà trên nhiều nền tảng và thiết bị khác nhau.
- **Tech Stack:**
    - **Language:** Go (Golang)
    - **Framework:** Gin Gonic (HTTP Server)
    - **Database:** SQLite (Dễ dàng triển khai, lưu trữ local)
    - **Protocols:**
        1. **HTTP (REST API):** Đăng ký, đăng nhập, CRUD Manga.
        2. **TCP:** Đồng bộ tiến độ đọc thời gian thực (Progress Sync).
        3. **gRPC:** Giao tiếp nội bộ hoặc dịch vụ dữ liệu hiệu năng cao.
        4. **WebSocket:** Chat Hub cho cộng đồng người đọc.
        5. **UDP:** Hệ thống thông báo (Notifier) nhanh, không tin cậy nhưng độ trễ thấp.
- **Port Configuration (dựa trên .env.example & Source Code):**
    - **HTTP:** 8080
    - **TCP:** 9090
    - **UDP:** 9999
    - **gRPC:** 50051

## 2. TRẠNG THÁI HIỆN TẠI (ĐÃ HOÀN THÀNH TỚI TUẦN 3)
### Các Module Đã Hoàn Thành:
- **Authentication JWT:** Xử lý tại `internal/auth`, hỗ trợ Register/Login.
- **CRUD Manga:** API lấy danh sách và thông tin truyện.
- **TCP Progress Sync:** Server tại `internal/protocols/tcp`, quản lý kết nối và broadcast tiến độ.
- **gRPC Manga Service:** Cung cấp service lấy chi tiết manga qua giao thức nhị phân.
- **WebSocket Chat Hub:** Hub xử lý tin nhắn nhóm tại `internal/protocols/websocket`.
- **UDP Notifier:** Cấu trúc đã sẵn sàng cho việc gửi thông báo nhanh.

### Cấu Trúc Database (Schema):
Dựa trên `backend/internal/database/db.go`:
- **Bảng `users`:**
    - `id`: INTEGER PRIMARY KEY AUTOINCREMENT
    - `username`: TEXT UNIQUE NOT NULL
    - `password`: TEXT NOT NULL
    - `created_at`: DATETIME DEFAULT CURRENT_TIMESTAMP
- **Bảng `mangas`:**
    - `id`: INTEGER PRIMARY KEY AUTOINCREMENT
    - `title`: TEXT NOT NULL
    - `author`: TEXT
    - `description`: TEXT
    - `thumbnail`: TEXT

### Các Struct Cốt Lõi (Go Models):
Dựa trên `backend/internal/models/models.go`:
- `User`: ID, Username, Password, CreatedAt.
- `Manga`: ID, Title, Author, Description, Thumbnail.
- `LoginRequest` / `TokenResponse`: Dùng cho luồng Auth.
- `TCPPayload`: (từ `tcp_payload.go`) Dùng để đóng gói dữ liệu đồng bộ qua TCP.

## 3. MỤC TIÊU TIẾP THEO (TUẦN 4)
- **Tích hợp & Xử lý lỗi:** 
    - Rà soát các Map lưu trữ Clients (như `clientsMap` trong TCP hoặc `clients` trong WebSocket Hub).
    - Triển khai `Mutex/RWMutex` để bảo vệ tài nguyên chia sẻ, tránh Race-condition khi có nhiều kết nối đồng thời.
- **Phát triển CLI Client:**
    - Xây dựng ứng dụng Terminal (tại `cmd/cli`) cho phép người dùng tương tác với hệ thống (xem danh sách, cập nhật tiến độ) mà không cần qua Web UI.

## 4. NGUYÊN TẮC BMAD (DEV GUIDELINES DÀNH CHO AI)
1. **Defensive Programming:** Không tin tưởng Input. Mọi thao tác I/O, Network, DB phải có xử lý lỗi chi tiết (try-catch/error logging).
2. **Layer-slicing:** Tôn trọng ranh giới file và kiến trúc. Không viết code Backend vào Frontend. 
3. **No Guessing:** Nếu lỗi xảy ra, phân tích Root-Cause trước khi sửa code. Nếu thiếu file/ngữ cảnh, yêu cầu User cung cấp, tuyệt đối không tự bịa hàm hoặc thư viện không có sẵn.
