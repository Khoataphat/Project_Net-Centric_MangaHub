# MangaHub — Hệ Thống Quản Lý Và Đồng Bộ Tiến Độ Đọc Manga Đa Giao Thức
## Tài Liệu Tổng Quan Hệ Thống (Dành Cho Báo Cáo Kỹ Thuật)

Tài liệu này phản ánh thực tế cấu trúc và tính năng đã được triển khai trong mã nguồn dự án MangaHub.

---

### 1. Thông Tin Chung (Technology Stack)
*   **Ngôn ngữ lập trình:** Go (Golang) phiên bản 1.25+.
*   **Framework:** 
    *   **Backend:** [Gin Gonic](https://github.com/gin-gonic/gin) (HTTP Framework).
    *   **Frontend:** HTML5, Vanilla JavaScript, TailwindCSS.
*   **Database:** SQLite (Sử dụng driver `modernc.org/sqlite` thuần Go).
    *   **File database:** `./data/mangahub.db`.
*   **Quản lý phụ thuộc:** `go.mod`.

---

### 2. The 5 Protocols (Giao Thức Lõi)
Hệ thống tích hợp đồng thời 5 giao thức mạng tại các cổng (port) mặc định như sau:

| Giao Thức | Port | File Xử Lý Chính | Vai Trò Thực Tế |
| :--- | :--- | :--- | :--- |
| **HTTP** | `8080` | `internal/protocols/http/` | REST API cho Authentication (JWT), CRUD Manga, Chapters, Pages. |
| **TCP** | `9090` | `internal/protocols/tcp/` | Đồng bộ tiến độ đọc (Progress Sync) và đếm người đọc trực tuyến (Live Presence). |
| **UDP** | `9999` | `internal/protocols/udp/` | Phát (Broadcast) thông báo về các sự kiện hệ thống (ví dụ: chương mới). |
| **gRPC** | `50051` | `internal/protocols/grpc/` | Truy vấn chi tiết Manga bằng định dạng nhị phân hiệu năng cao. |
| **WebSocket** | `8080` | `internal/protocols/websocket/` | Chat real-time, Stream Server Logs và Bridge kết nối Browser tới TCP Server. |

---

### 3. Kết Quả & Tính Năng (Result & Features)

#### **Tính năng đồng bộ TCP (Progress Sync)**
*   **Cơ chế:** Hoạt động dựa trên kết nối Socket (Raw Socket). Server nhóm các kết nối vào `clientsMap` theo cặp `(userID, mangaID)`.
*   **Hoạt động:** Khi một thiết bị gửi gói tin cập nhật chương (`UPDATE_PROGRESS`), server sẽ broadcast dữ liệu này tới tất cả các thiết bị khác của cùng user đó đang mở cùng bộ truyện.
*   **Live Presence:** Sử dụng `presenceMap` để theo dõi tất cả kết nối theo `mangaID`, từ đó broadcast số lượng "Người đang đọc" thực tế tới toàn bộ client.
*   **File:** [`internal/protocols/tcp/server.go`](backend/internal/protocols/tcp/server.go).

#### **Chat WebSocket và UDP Notification**
*   **WebSocket Chat:** Sử dụng mô hình **Hub** để quản lý các client. Tin nhắn được truyền qua **Channels** và phát đi (Broadcast) tới toàn bộ thành viên trong phòng chat. 
    *   File: [`internal/protocols/websocket/hub.go`](backend/internal/protocols/websocket/hub.go).
*   **UDP Notification:** Khi Admin thêm chương mới qua HTTP, hệ thống kích hoạt `BroadcastUpdate`. Gói tin được serialize sang JSON và gửi "fire-and-forget" tới các Bridge node (port 8888) để đẩy thông báo Toast lên UI người dùng.
    *   File: [`internal/protocols/udp/server.go`](backend/internal/protocols/udp/server.go).

#### **Xử lý Concurrency (Đồng quy)**
*   **Goroutines:** Được sử dụng để chạy song song 5 protocol servers mà không làm nghẽn luồng chính.
*   **Channels:** Dùng trong WebSocket Hub để đăng ký (`register`), hủy đăng ký (`unregister`) và phát tin nhắn (`broadcast`) an toàn.
*   **Mutex / RWMutex:** Sử dụng `sync.Mutex` trong TCP Pool và `sync.RWMutex` trong UDP Server để bảo vệ các biến global (Map, Slice) khỏi tình trạng Race Condition khi nhiều client kết nối cùng lúc.
*   **Context:** Sử dụng `context.Context` phối hợp với tín hiệu hệ điều hành (`SIGTERM`) để thực hiện **Graceful Shutdown** (đóng server sạch sẽ).

---

### 4. Hướng Dẫn Chạy & Triển Khai (How to run / deploy)

#### **Các lệnh Terminal cần thiết (Tại thư mục `backend/`):**
1.  **Tải thư viện:**
    ```bash
    go mod download
    ```
2.  **Khởi tạo dữ liệu mẫu (Seed Database):**
    ```bash
    go run ./cmd/seed/main.go
    ```
3.  **Chạy Server chính:**
    ```bash
    go run ./cmd/server/main.go
    ```
4.  **Chạy CLI Client (Để test các giao thức):**
    ```bash
    go run ./cmd/client/main.go
    ```

#### **Thiết lập biến môi trường (.env):**
Hệ thống hỗ trợ cấu hình qua Environment Variables (mặc định sẽ dùng các giá trị sau nếu không thiết lập):
*   `HTTP_PORT=8080`
*   `TCP_PORT=9090`
*   `UDP_PORT=9999`
*   `GRPC_PORT=50051`
*   `DB_PATH=./data/mangahub.db`

#### **Yêu cầu hệ thống:**
*   Hệ điều hành: Windows/Linux/MacOS (64-bit).
*   Go version 1.25.0+.
*   Nếu dùng Docker: Chạy lệnh `docker-compose up --build -d` tại thư mục gốc.
