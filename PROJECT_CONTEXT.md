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
        4. **WebSocket:** Chat Hub & Hệ thống nhận thông báo thời gian thực.
        5. **UDP:** Hệ thống thông báo (Notifier) gửi cập nhật chương mới nhanh chóng.
        6. **Bridge Module:** Module trung gian chuyển tiếp thông báo từ UDP Server sang WebSocket Clients.
- **Port Configuration:**
    - **HTTP:** 8080
    - **TCP:** 9090
    - **UDP Notifier:** 9999
    - **UDP Bridge Listener:** 8888
    - **gRPC:** 50051

## 2. TRẠNG THÁI HIỆN TẠI (GIAI ĐOẠN TÍCH HỢP & PHÁT HIỆN HẠN CHẾ)
### Các Module Đã Triển Khai (Mức Độc Lập):
- **Authentication JWT:** Xử lý tại `internal/auth`, hỗ trợ Register/Login.
- **CRUD Manga:** API lấy danh sách và thông tin truyện (HTTP).
- **TCP Progress Sync:** Quản lý kết nối và đồng bộ tiến độ qua Socket thô.
- **gRPC Manga Service:** Cung cấp dữ liệu manga hiệu năng cao.
- **UDP Notifier & Bridge:** Hệ thống thông báo chương mới đã chạy nhưng mới chỉ ở dạng thử nghiệm rời rạc.
- **Frontend Templates:** Các file HTML (`dashboard`, `admin`, `notification`) đã có giao diện nhưng hoạt động độc lập, chưa có bộ điều hướng (Router) và quản lý trạng thái (State) tập trung.

### Hạn Chế Hiện Tại (Technical Debt - Cần Giải Quyết Trong Tuần 4):
> [!WARNING]
> **Sự rời rạc của các Giao thức (Protocol Islands):**
> - Các giao thức chưa kết nối thành một luồng dữ liệu thống nhất. Ví dụ: Khi thêm Manga qua HTTP, hệ thống không tự động kích hoạt thông báo qua UDP/WebSocket.
> - Thiếu một **Event Bus** trung tâm để điều phối sự kiện giữa các module backend.
>
> **Hạn chế của Frontend:**
> - Các file HTML đang hoạt động như các trang tĩnh rời rạc, chưa tích hợp đầy đủ tính năng vào một giao diện duy nhất.
> - Logic xử lý Auth (Lưu Token, tự động Re-login) và kết nối đa giao thức trên trình duyệt chưa được đồng bộ hóa.


### Cấu Trúc Database (Schema):
- **Bảng `users`:** `id`, `username`, `password`, `created_at`.
- **Bảng `mangas`:** `id`, `title`, `author`, `description`, `thumbnail`.

### Các Struct Cốt Lõi (Go Models):
- `User`, `Manga`, `LoginRequest`, `TokenResponse` (tại `models/models.go`).
- `UDPPayload`: Chứa thông tin thông báo chương mới (MangaID, Chapter, Title, Timestamp).

## 3. MỤC TIÊU TIẾP THEO (TUẦN 4)
- **Hoàn thiện CLI Client:** Xây dựng ứng dụng Terminal (tại `cmd/cli`) để tương tác với các service TCP/gRPC/HTTP.
- **Unit Testing & Stress Test:** Tăng cường độ bao phủ kiểm thử cho các module giao thức, đặc biệt là xử lý tranh chấp tài nguyên (Race-condition).
- **Phát triển Dashboard Admin:** Giao diện quản lý gửi thông báo UDP trực tiếp từ Web.

## 4. NGUYÊN TẮC BMAD (DEV GUIDELINES DÀNH CHO AI)
1. **Defensive Programming:** Không tin tưởng Input. Mọi thao tác I/O, Network, DB phải có xử lý lỗi chi tiết (try-catch/error logging).
2. **Layer-slicing:** Tôn trọng ranh giới file và kiến trúc. Không viết code Backend vào Frontend. 
3. **No Guessing:** Nếu lỗi xảy ra, phân tích Root-Cause trước khi sửa code. Nếu thiếu file/ngữ cảnh, yêu cầu User cung cấp, tuyệt đối không tự bịa hàm hoặc thư viện không có sẵn.
