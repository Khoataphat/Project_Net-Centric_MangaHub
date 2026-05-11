# MangaHub — Hệ Thống Quản Lý Và Đồng Bộ Tiến Độ Đọc Manga Đa Giao Thức

MangaHub là dự án Net-Centric mô phỏng một hệ thống đọc manga phân tán hỗ trợ **5 giao thức mạng chính**: HTTP/REST, TCP, gRPC, WebSocket và UDP. Hệ thống cho phép người dùng đăng ký, đăng nhập, duyệt danh sách manga, đồng bộ tiến độ đọc giữa nhiều thiết bị, chat real-time và nhận thông báo cập nhật chương truyện mới.

---

## 📋 Mục Lục

1. [Kiến Trúc Nhanh](#kiến-trúc-nhanh)
2. [Hướng Dẫn Cài Đặt Và Chạy Project](#hướng-dẫn-cài-đặt-và-chạy-project)
3. [Thông Tin Đăng Nhập Mặc Định](#thông-tin-đăng-nhập-mặc-định)
4. [Biến Môi Trường (.env)](#biến-môi-trường-env)
5. [Bản Đồ Port & Giao Thức](#bản-đồ-port--giao-thức)
6. [Cấu Trúc Database](#cấu-trúc-database)
7. [Ví Dụ Sử Dụng CLI Client](#ví-dụ-sử-dụng-cli-client)
8. [Tính Năng Chính](#tính-năng-chính)

---

## 🏗️ Kiến Trúc Nhanh

### Cấu Trúc Thư Mục

```
Project_Net-Centric_MangaHub/
├── backend/
│   ├── cmd/
│   │   ├── server/main.go           # Backend server entrypoint
│   │   ├── client/main.go           # CLI client REPL
│   │   └── seed/main.go             # Seed data script
│   ├── internal/
│   │   ├── auth/                    # JWT & password hashing
│   │   ├── database/                # SQLite initialization
│   │   ├── models/                  # Data structures
│   │   └── protocols/               # Network protocol handlers
│   │       ├── http/                # REST API & web serving
│   │       ├── tcp/                 # Progress sync protocol
│   │       ├── grpc/                # Binary manga queries
│   │       ├── websocket/           # Chat & log broadcasting
│   │       ├── udp/                 # Event notifications
│   │       └── bridge/              # UDP-to-WebSocket bridge
│   ├── proto/                       # gRPC protobuf definitions
│   ├── data/                        # SQLite database file (runtime)
│   ├── Dockerfile                   # Docker build config
│   ├── go.mod                       # Go dependencies
│   └── seed.go                      # Alternative seed source
├── frontend/
│   ├── auth.html                    # Login & registration UI
│   ├── dashboard.html               # Admin panel + Web Terminal
│   ├── reading.html                 # Manga reading + sync UI
│   ├── notification.js              # Toast notification system
│   ├── sync.js                      # TCP sync bridge client
│   └── scratch/                     # Development utilities
├── docs/                            # Documentation & demos
├── docker-compose.yml               # Multi-service orchestration
└── run_mangahub.bat                 # Windows batch starter script
```

### Thành Phần Chính

| Thành Phần | Vị Trí | Vai Trò |
|---|---|---|
| **Backend Server** | `backend/cmd/server/main.go` | Khởi chạy toàn bộ hệ thống (HTTP, TCP, gRPC, UDP, WebSocket) |
| **CLI Client** | `backend/cmd/client/main.go` | REPL interactivo để test 5 giao thức mạng |
| **Seed Database** | `backend/cmd/seed/main.go` | Tạo dữ liệu mẫu (5 manga + 3 user accounts) |
| **HTTP Handler** | `backend/internal/protocols/http/` | REST API endpoints, JWT auth, static file serving |
| **TCP Server** | `backend/internal/protocols/tcp/` | Raw socket progress sync đa thiết bị |
| **gRPC Server** | `backend/internal/protocols/grpc/` | Efficient binary queries cho manga |
| **WebSocket Hub** | `backend/internal/protocols/websocket/` | Chat, log streaming, notifications |
| **UDP Server** | `backend/internal/protocols/udp/` | Event notifications broadcaster |
| **Bridge Service** | `backend/internal/protocols/bridge/` | Convert UDP events → WebSocket messages |
| **Database Layer** | `backend/internal/database/` | SQLite auto-initialization & schema |
| **Frontend** | `frontend/*.html` | Static HTML/JS interfaces (auth, dashboard, reading) |

---

## 🚀 Hướng Dẫn Cài Đặt Và Chạy Project

### 📋 Điều Kiện Tiên Quyết

**Option A: Chạy Local (Go 1.25+)**
- Go 1.25.0 trở lên (download từ https://golang.org/dl)
- **Lưu ý quan trọng:** Phải cài đặt Go **64-bit** (`windows/amd64`), không dùng 32-bit (`windows/386`). Điều này do thư viện `modernc.org/sqlite` không hỗ trợ 32-bit.
- Kiểm tra: `go version` phải in ra `go1.25.x windows/amd64`

**Option B: Chạy Docker (Khuyến Nghị)**
- Docker Desktop (https://www.docker.com/products/docker-desktop)
- Không cần cài Go hay dependencies khác

### 🔧 Các Bước Cài Đặt Chi Tiết

#### **Cách 1: Chạy Bằng Docker Compose (Khuyến Nghị Nhất)**

Cách này tối giản, chỉ cần 1 lệnh duy nhất:

```bash
# Từ thư mục gốc Project_Net-Centric_MangaHub/
docker-compose up --build -d
```

Hệ thống sẽ:
- Build Docker image cho backend
- Tạo container `mangahub-backend` 
- Tự động khởi tạo database SQLite
- Chạy tất cả 5 protocol servers
- Serve frontend static files

Kiểm tra status:
```bash
docker-compose ps
docker-compose logs -f mangahub-backend
```

Dừng hệ thống:
```bash
docker-compose down
```

---

#### **Cách 2: Chạy Local (Go CLI)**

**Bước 1: Cài đặt Dependencies Backend**

```bash
cd backend
go mod download
```

**Bước 2: Tạo Seed Data (Lựa Chọn)**

Script `cmd/seed/main.go` tạo dữ liệu mẫu: 5 manga + 3 user accounts.

```bash
# Từ thư mục backend/
go run ./cmd/seed/main.go
```

Output mong đợi:
```
--- Đang chèn dữ liệu mẫu vào MangaHub ---
Đã thêm user thành công: admin (Password: admin123)
Đã thêm user thành công: user1 (Password: 123)
Đã thêm user thành công: user2 (Password: 123)
Đã thêm thành công: One Piece
Đã thêm thành công: Naruto
...
--- Hoàn tất! Bây giờ bạn có thể mở Dashboard để xem ---
```

**Bước 3: Chạy Backend Server**

```bash
# Từ thư mục backend/
go run ./cmd/server/main.go
```

Server sẽ khởi chạy toàn bộ 5 giao thức. Output mong đợi:

```
[HTTP] Server đang chạy tại 0.0.0.0:8080
[TCP] Sync Server đang chạy tại 0.0.0.0:9090
[gRPC] Server đang chạy tại 0.0.0.0:50051
[UDP] Notifier Server đang khởi động tại 0.0.0.0:9999
[Bridge] UDP Bridge đang chạy tại 0.0.0.0:8888
```

**Bước 4: Mở Frontend**

Mở trình duyệt:
```
http://localhost:8080
```

Bạn sẽ thấy trang đăng nhập `auth.html`.

**Bước 5: (Tùy Chọn) Chạy CLI Client**

Mở terminal khác (giữ server đang chạy):

```bash
cd backend
go run ./cmd/client/main.go
```

Interactive prompt:
```
mangahub> help
Ví dụ lệnh:
  http register admin 123
  http login admin 123
  http mangas "one piece"
  grpc scan 101
  tcp sync 101 5 12
  ws chat "Hello"
  udp ping
  exit
```

---

#### **Cách 3: Chạy Bằng Windows Batch Script (Windows Chỉ)**

File `run_mangahub.bat` tự động xử lý cleanup ports và khởi chạy Docker:

```bash
# Từ thư mục gốc Project_Net-Centric_MangaHub/
run_mangahub.bat
```

Script sẽ:
1. Dừng tất cả Docker containers cũ
2. Kill các process đang dùng ports 8080, 9090, 50051, 9999, 8888
3. Build & run `docker-compose up --build -d`
4. Mở trình duyệt tại http://localhost:8080

---

### ⚡ Quick Start (Tóm Tắt)

**Nhanh nhất (Docker):**
```bash
docker-compose up --build -d
# Sau 5 giây, truy cập: http://localhost:8080
```

**Chạy Local (Go):**
```bash
cd backend
go mod download
go run ./cmd/seed/main.go    # (Optional) Seed data
go run ./cmd/server/main.go  # Run backend

# Terminal khác: Mở http://localhost:8080
# (Hoặc chạy CLI client: go run ./cmd/client/main.go)
```

---

## 🔐 Thông Tin Đăng Nhập Mặc Định

Dữ liệu seed được tạo từ file [backend/cmd/seed/main.go](backend/cmd/seed/main.go). Khi khởi chạy backend lần đầu, database SQLite sẽ tự động khởi tạo bảng `users` nếu chưa tồn tại.

### Tài Khoản Test Sẵn Có

| Username | Password | Vai Trò | Ghi Chú |
|---|---|---|---|
| `admin` | `admin123` | Admin | Có quyền truy cập `/api/admin/scan-manga` |
| `user1` | `123` | User | Người dùng thường |
| `user2` | `123` | User | Người dùng thường |

### Hướng Dùng

**Trên Frontend (auth.html):**
1. Nhập Username: `admin`
2. Nhập Password: `admin123`
3. Click "Đăng Nhập"
4. Token JWT sẽ được lưu vào `localStorage` dưới key `token`
5. Redirect tới dashboard

**Trên CLI Client:**
```bash
mangahub> http login admin admin123
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Tạo Tài Khoản Mới:**

Trên Frontend (auth.html) → tab "Đăng Ký":
```
Username: myuser
Password: mypassword123
Click "Đăng Ký"
```

Hoặc CLI:
```bash
mangahub> http register myuser mypassword123
{
  "message": "Đăng ký thành công"
}
```

---

## 🌍 Biến Môi Trường (.env)

Backend server đọc các biến môi trường từ hệ thống. Nếu chưa set, sẽ dùng giá trị mặc định.

**Tất cả biến được quản lý trong file:** [backend/cmd/server/main.go](backend/cmd/server/main.go)

### Danh Sách Biến Môi Trường

| Biến | Mặc Định | Kiểu | Ví Dụ | Ghi Chú |
|---|---|---|---|---|
| `HOST` | `0.0.0.0` | string | `127.0.0.1` hoặc `0.0.0.0` | IP bind cho tất cả servers |
| `HTTP_PORT` | `8080` | string | `8080` | Port cho HTTP REST & WebSocket |
| `TCP_PORT` | `9090` | string | `9090` | Port cho TCP Sync Server |
| `GRPC_PORT` | `50051` | string | `50051` | Port cho gRPC Server |
| `UDP_PORT` | `9999` | string | `9999` | Port cho UDP Notifier |
| `BRIDGE_PORT` | `8888` | string | `8888` | Port cho UDP-to-WebSocket Bridge |
| `DB_PATH` | `./data/mangahub.db` | string | `/app/data/mangahub.db` | Đường dẫn SQLite database file |

### Cách Set Biến Môi Trường

**Linux/Mac/PowerShell:**
```bash
export HOST=127.0.0.1
export HTTP_PORT=3000
export DB_PATH=/tmp/mangahub.db
go run ./cmd/server/main.go
```

**Windows CMD:**
```cmd
set HOST=127.0.0.1
set HTTP_PORT=3000
set DB_PATH=C:\tmp\mangahub.db
go run ./cmd/server/main.go
```

**Docker (docker-compose.yml):**
```yaml
environment:
  - HOST=0.0.0.0
  - HTTP_PORT=8080
  - TCP_PORT=9090
  - UDP_PORT=9999
  - BRIDGE_PORT=8888
  - GRPC_PORT=50051
  - DB_PATH=/app/data/mangahub.db
```

**Ví Dụ .env File (Nếu muốn dùng):**

Tạo file `.env` trong thư mục `backend/`:
```env
HOST=0.0.0.0
HTTP_PORT=8080
TCP_PORT=9090
GRPC_PORT=50051
UDP_PORT=9999
BRIDGE_PORT=8888
DB_PATH=./data/mangahub.db
```

Sau đó load trước khi chạy (tùy theo shell):
```bash
source .env  # Linux/Mac
# hoặc
export $(cat .env | xargs)  # Linux/Mac
```

> **Lưu ý:** Backend hiện tại đọc trực tiếp từ `os.LookupEnv()`, không dùng thư viện `.env` loader. Nếu muốn tự động load từ `.env`, cần thêm thư viện như `github.com/joho/godotenv`.

---

## 🔌 Bản Đồ Port & Giao Thức

| Giao Thức | Port | URL/Addr | Vai Trò | Kiểu Kết Nối |
|---|---|---|---|---|
| **HTTP REST** | 8080 | `http://localhost:8080/api/` | REST API, Auth, Manga CRUD | Stateless, request-reply |
| **WebSocket Chat** | 8080 | `ws://localhost:8080/api/ws/chat` | Real-time chat room | Bidirectional |
| **WebSocket Logs** | 8080 | `ws://localhost:8080/api/ws-logs` | Stream backend logs to admin UI | Unidirectional (server → client) |
| **WebSocket TCP Bridge** | 8080 | `ws://localhost:8080/api/ws-tcp-bridge` | Browser → TCP sync proxy | Bidirectional proxy |
| **TCP Sync** | 9090 | `tcp://localhost:9090` | Raw socket progress sync (đa thiết bị) | Connection-oriented, binary JSON |
| **gRPC** | 50051 | `localhost:50051` | Efficient protobuf queries | Binary, multiplexed |
| **UDP Notifier** | 9999 | `udp://localhost:9999` | Event broadcaster (chương mới) | Datagram, fire-and-forget |
| **UDP Bridge** | 8888 | `udp://localhost:8888` | Convert UDP → WebSocket | Receiver only |

### Luồng Kết Nối Chi Tiết

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend Browser                         │
│  (auth.html, dashboard.html, reading.html)                      │
└──────┬─────────────────────────────────────────────────────────┘
       │
       ├─ HTTP GET /          ────→ Serve auth.html
       │
       ├─ POST /api/login     ────→ JWT Token
       │
       ├─ WS /api/ws/chat     ────→ Chat Hub (Bidirectional)
       │
       ├─ WS /api/ws-logs     ←──── Server Logs Stream
       │
       └─ WS /api/ws-tcp-bridge ──→ (Tunnel to TCP Server)
                                      │
                                      ├─ Binary Sync Messages
                                      └─ TCP://9090

┌──────────────────────────────────────────────────────────┐
│        CLI Client (cmd/client/main.go)                   │
├──────────────────────────────────────────────────────────┤
│ ├─ HTTP queries (register, login, mangas, scan)  │
│ ├─ TCP connections (sync progress)               │
│ ├─ gRPC calls (manga detail lookup)               │
│ ├─ WebSocket (chat messages)                     │
│ └─ UDP (ping notifier)                           │
└──────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────┐
│  UDP Event Stream (Admin/System)                   │
│  127.0.0.1:9999  ←────── (Generated events)       │
│      │                                             │
│      └──→ Bridge (8888)                           │
│           └──→ WebSocket Hub                      │
│               └──→ Notify all connected clients    │
└────────────────────────────────────────────────────┘
```

---

## 💾 Cấu Trúc Database

SQLite database được tự động khởi tạo từ [backend/internal/database/db.go](backend/internal/database/db.go) khi backend start lần đầu.

### Bảng Chính

#### **Bảng `users`**
```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT DEFAULT 'user',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

| Cột | Kiểu | Ghi Chú |
|---|---|---|
| `id` | INTEGER | Primary key, auto-increment |
| `username` | TEXT | Unique, không được trống |
| `password` | TEXT | Bcrypt hashed |
| `role` | TEXT | `'admin'` hoặc `'user'` |
| `created_at` | DATETIME | Dấu thời gian tạo |

**Dữ liệu mẫu:** Xem [Thông Tin Đăng Nhập Mặc Định](#thông-tin-đăng-nhập-mặc-định)

---

#### **Bảng `mangas`**
```sql
CREATE TABLE IF NOT EXISTS mangas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    author TEXT,
    description TEXT,
    thumbnail TEXT,
    mangadex_id TEXT UNIQUE
);
```

| Cột | Kiểu | Ghi Chú |
|---|---|---|
| `id` | INTEGER | Primary key |
| `title` | TEXT | Tên truyện |
| `author` | TEXT | Tác giả |
| `description` | TEXT | Mô tả dài |
| `thumbnail` | TEXT | URL ảnh bìa |
| `mangadex_id` | TEXT | ID từ MangaDex (tùy chọn) |

**Dữ liệu mẫu:** 5 manga được seed từ script `cmd/seed/main.go`
- One Piece
- Naruto
- Dragon Ball
- Doraemon
- Conan

---

#### **Bảng `user_progress`** (Mới - Tuần 4)
```sql
CREATE TABLE IF NOT EXISTS user_progress (
    user_id INTEGER NOT NULL,
    manga_id INTEGER NOT NULL,
    last_chapter INTEGER DEFAULT 1,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, manga_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (manga_id) REFERENCES mangas(id)
);
```

| Cột | Kiểu | Ghi Chú |
|---|---|---|
| `user_id` | INTEGER | Foreign key → `users(id)` |
| `manga_id` | INTEGER | Foreign key → `mangas(id)` |
| `last_chapter` | INTEGER | Chương cuối cùng đang đọc |
| `updated_at` | DATETIME | Thời gian đồng bộ gần nhất |

---

#### **Bảng `chapters` & `pages`**
```sql
CREATE TABLE IF NOT EXISTS chapters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    manga_id INTEGER NOT NULL,
    chapter_number REAL,
    title TEXT,
    mangadex_id TEXT UNIQUE,
    FOREIGN KEY (manga_id) REFERENCES mangas(id)
);

CREATE TABLE IF NOT EXISTS pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chapter_id INTEGER NOT NULL,
    page_number INTEGER NOT NULL,
    image_url TEXT NOT NULL,
    FOREIGN KEY (chapter_id) REFERENCES chapters(id)
);
```

---

## 🎮 Ví Dụ Sử Dụng CLI Client

CLI client được khởi chạy từ [backend/cmd/client/main.go](backend/cmd/client/main.go).

### Khởi Động

```bash
cd backend
go run ./cmd/client/main.go
```

Output:
```
=== MANGAHUB CLI CLIENT ===
Hỗ trợ 5 giao thức: HTTP, TCP, GRPC, WS, UDP
Gõ "help" để xem ví dụ, "exit" để thoát.
mangahub>
```

### Các Lệnh Ví Dụ

#### **1. HTTP Protocol**

```bash
# Xem danh sách lệnh
mangahub> help

# Đăng ký tài khoản mới
mangahub> http register myuser 123456
Output: {"message": "Đăng ký thành công"}

# Đăng nhập (lấy JWT token)
mangahub> http login admin admin123
Output: {"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}

# Tìm kiếm manga (lấy danh sách)
mangahub> http mangas "one piece"
Output: [
  {
    "id": 1,
    "title": "One Piece",
    "author": "Eiichiro Oda",
    "description": "Hành trình trở thành Vua Hải Tặc.",
    "thumbnail": "https://..."
  }
]

# Scan manga (admin only)
mangahub> http scan 1
```

#### **2. TCP Protocol** (Progress Sync)

```bash
# Sync tiến độ: tcp sync <user_id> <manga_id> <chapter_number>
mangahub> tcp sync 1 1 50
Output: [TCP] Kết nối tới 127.0.0.1:9090
        [TCP] Gửi AUTH: User 1
        [TCP] UPDATE_PROGRESS nhận được: Manga 1, Chapter 50
```

#### **3. gRPC Protocol**

```bash
# Scan chi tiết manga theo ID
mangahub> grpc scan 1
Output: {
  "id": 1,
  "title": "One Piece",
  "author": "Eiichiro Oda",
  "description": "..."
}
```

#### **4. WebSocket Chat**

```bash
# Gửi tin nhắn chat
mangahub> ws chat "Hello everyone!"
Output: [WS] Kết nối tới ws://127.0.0.1:8080/api/ws/chat
        [WS] Gửi: {"message":"Hello everyone!"}
        [WS] Nhận: {"username":"admin","message":"Hello everyone!","timestamp":"2026-05-11T10:30:00Z"}
```

#### **5. UDP Protocol**

```bash
# Ping UDP server (thử kết nối)
mangahub> udp ping
Output: [UDP] Gửi PING tới 127.0.0.1:8888

# Gửi thông báo (admin feature)
mangahub> udp notify 1 50 "Chapter mới đã ra mắt"
```

#### **6. Thoát**

```bash
mangahub> exit
Đang thoát...
```

---

## ✨ Tính Năng Chính

### ✅ Đã Hoàn Thành

1. **Xác Thực & Phân Quyền (HTTP)**
   - Đăng ký (POST `/api/register`)
   - Đăng nhập (POST `/api/login`) + JWT Token
   - Mã hóa mật khẩu bcrypt
   - Phân biệt role `admin` vs `user`
   - Middleware kiểm tra quyền admin (RoleMiddleware)

2. **Quản Lý Manga (HTTP REST)**
   - Lấy danh sách toàn bộ manga (GET `/api/mangas`)
   - Tìm kiếm manga theo từ khóa (GET `/api/mangas?q=keyword`)
   - Lấy chi tiết manga theo ID (GET `/api/mangas/:id`)
   - Lấy danh sách chapter (GET `/api/mangas/:id/chapters`)
   - Lấy danh sách page mỗi chapter (GET `/chapters/:chapter_id/pages`)

3. **Đồng Bộ Tiến Độ Đa Thiết Bị (TCP)**
   - Raw socket server trên port 9090
   - Giao thức: Binary JSON lines
   - Hỗ trợ AUTH + UPDATE_PROGRESS messages
   - Broadcast progress tới các client khác (cùng user + manga)
   - Live presence: đếm số người đang đọc mỗi manga
   - Mutex-protected connection pool để tránh race condition

4. **Truy Vấn Hiệu Năng Cao (gRPC)**
   - gRPC server trên port 50051
   - Proto definition: `proto/manga.proto`
   - Service: `MangaService.ScanManga(id)` → trả protobuf binary
   - Kết nối persistent + multiplexing

5. **Chat Real-time (WebSocket)**
   - Hub-based broadcast system
   - Endpoint: `ws://localhost:8080/api/ws/chat`
   - Hỗ trợ multiple concurrent connections
   - Message format: JSON `{username, message, timestamp}`

6. **Thông Báo Sự Kiện (UDP)**
   - UDP Notifier server trên 9999
   - Broadcast events (chương mới, etc.)
   - UDP Bridge (8888) → chuyển UDP → WebSocket
   - Toast notifications trên dashboard

7. **Log Streaming Real-time (WebSocket)**
   - Backend logs được stream tới `/api/ws-logs`
   - Web Terminal trên dashboard nhận log thực tế
   - Multiwriter: log vừa xuất stdout, vừa broadcast WebSocket

8. **Web Frontend**
   - `auth.html`: Login + Registration UI (Glassmorphism design)
   - `dashboard.html`: Admin panel + Web Terminal + Manga management
   - `reading.html`: Manga reader + TCP sync UI + Chat widget
   - Responsive design (TailwindCSS)

9. **CLI Client REPL**
   - Interactive shell (cmd/client/main.go)
   - Hỗ trợ test tất cả 5 protocol
   - Helper functions + error handling

10. **Docker Support**
    - Dockerfile multi-stage (build + runtime)
    - docker-compose.yml orchestration
    - Volume mounting untuk persistent database
    - Environment variables config

---

### 🔄 Lưu Lượng Dữ Liệu Chính

```
┌── REGISTER ──────────────────────────────────────────┐
│ POST /api/register                                   │
│ {username, password}  ──→  Bcrypt hash  ──→  SQLite │
│                                                      │
└──────────────────────────────────────────────────────┘

┌── LOGIN + JWT ──────────────────────────────────────┐
│ POST /api/login                                      │
│ {username, password}  ──→  Check hash  ──→  JWT     │
│                           (24h exp)                   │
│                                                      │
└──────────────────────────────────────────────────────┘

┌── TCP SYNC ─────────────────────────────────────────┐
│ Browser ──[WS]──→ TCPBridgeHandler ──[TCP]──→ TCP   │
│                                         Server:9090   │
│                                              │        │
│                   {"type":"AUTH", ...}      │        │
│                   {"type":"UPDATE_..."}    │        │
│                                              ↓        │
│                         Broadcast to other devices   │
│                         (same user + manga)          │
│                                                      │
└──────────────────────────────────────────────────────┘

┌── UDP → WebSocket Bridge ───────────────────────────┐
│ Admin Action  ──→  UDP:9999  ──→  Bridge:8888       │
│ (chương mới)       (event)        (UDP listener)     │
│                                      │               │
│                                      └─→ WS Hub      │
│                                          └─→ Toast   │
│                                              Notify  │
│                                                      │
└──────────────────────────────────────────────────────┘
```

---

## 🔗 Các Endpoint HTTP API Chính

Đầy đủ chi tiết tại [backend/internal/protocols/http/router.go](backend/internal/protocols/http/router.go)

| Method | Endpoint | Auth | Quyền | Ghi Chú |
|---|---|---|---|---|
| GET | `/` | ❌ | - | Serve `auth.html` |
| GET | `/api/health` | ❌ | - | Health check |
| POST | `/api/register` | ❌ | - | Tạo tài khoản mới |
| POST | `/api/login` | ❌ | - | Lấy JWT token |
| GET | `/api/mangas` | ❌ | - | Danh sách manga (có search) |
| GET | `/api/mangas/:id` | ❌ | - | Chi tiết manga |
| GET | `/api/mangas/:id/chapters` | ❌ | - | Danh sách chapter |
| GET | `/api/chapters/:chapter_id/pages` | ❌ | - | Danh sách page |
| GET | `/api/admin/scan-manga` | ✅ | Admin | Scan chi tiết (admin only) |
| GET | `/api/ws-tcp-bridge` | ❌ | - | WebSocket → TCP proxy |
| GET | `/api/ws/chat` | ❌ | - | WebSocket chat room |
| GET | `/api/ws-logs` | ❌ | - | Stream server logs |
| GET | `/*` (NoRoute) | ❌ | - | Serve static files (`./web/*`) |

---

## 📱 Frontend File Organization

| File | Mục Đích | Đường Dẫn |
|---|---|---|
| `auth.html` | Login & Registration UI | [frontend/auth.html](frontend/auth.html) |
| `dashboard.html` | Admin panel + Web Terminal | [frontend/dashboard.html](frontend/dashboard.html) |
| `reading.html` | Manga reader + sync + chat | [frontend/reading.html](frontend/reading.html) |
| `notification.js` | Toast notification system | [frontend/notification.js](frontend/notification.js) |
| `sync.js` | TCP bridge WebSocket client | [frontend/sync.js](frontend/sync.js) |
| `notification.test.js` | Unit tests | [frontend/notification.test.js](frontend/notification.test.js) |

---

## 🐛 Troubleshooting

### Backend không khởi động

**Lỗi:** `undefined: sqlite3_index_constraint`

**Nguyên Nhân:** Go 32-bit (`windows/386`) không tương thích với `modernc.org/sqlite`

**Giải Pháp:**
1. Cài Go 64-bit: https://golang.org/dl
2. Xóa Go 32-bit (nếu có)
3. Kiểm tra: `go version` phải là `go1.25.x windows/amd64`

### Port đã bị chiếm

**Lỗi:** `bind: port already in use`

**Giải Pháp (Windows):**
```bash
# Tìm process chiếm port
netstat -aon | findstr :8080

# Kill process (PID = số cuối cùng)
taskkill /F /PID <PID>
```

**Giải Pháp (Linux/Mac):**
```bash
lsof -i :8080
kill -9 <PID>
```

### Frontend không kết nối được backend

**Kiểm tra:**
1. Backend đã start? `curl http://localhost:8080/api/health`
2. Kiểm tra browser console (F12) → Network tab
3. Kiểm tra CORS headers (xem [CORSMiddleware](backend/cmd/server/main.go))

### Database locked

**Lỗi:** `database is locked`

**Nguyên Nhân:** SQLite tính single-writer

**Giải Pháp:** Xem [db.go](backend/internal/database/db.go) - `SetMaxOpenConns(1)` + `PRAGMA journal_mode=DELETE`

---

## 📚 Tài Liệu Thêm

- **Feature Flows:** [feature_protocol_flows.md](feature_protocol_flows.md)
- **Project Context:** [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md)
- **Demo Scripts:** [docs/Demo_Script.md](docs/Demo_Script.md)
- **Release Stories:** [docs/stories/](docs/stories/)

---

## 📄 License

Dự án này được tạo cho mục đích học tập Net-Centric Architecture (Tuần 4).

---

**Cập nhật lần cuối:** May 2026 | **Version:** 1.0 (Complete Edition)
```text
grpc scan 101
```
5. Ky vong:
- CLI in response gRPC.
- Dashboard Web Terminal hien log backend, vi du: `[gRPC] Nhan request Scan Manga 101`.

## API va protocol nhanh
### HTTP
- `POST /api/register`
- `POST /api/login`
- `GET /api/mangas?q=keyword`
- `GET /api/mangas/:id`
- `GET /api/admin/scan-manga?id=101`

### WebSocket
- `/api/ws/chat`: chat room.
- `/api/ws-tcp-bridge`: bridge browser <-> TCP server.
- `/api/ws-logs`: stream log backend xuong dashboard.

### TCP
Payload JSON ket thuc bang `\n`:
```json
{
  "type": "UPDATE_PROGRESS",
  "user_id": 1,
  "manga_id": 101,
  "chapter": 12
}
```

### gRPC
Service: `MangaService.GetMangaDetail`

### UDP
Payload:
```json
{
  "manga_id": "101",
  "chapter": 12,
  "title": "Chapter moi da ra mat",
  "timestamp": 1710000000
}
```

## Kiem tra minh da lam
- Verify `go run ./cmd/client`: pass.
- Verify CLI REPL in help va exit dung.
- Verify package `internal/protocols/websocket`: build pass.
- Verify `go run ./cmd/server`: fail do SQLite dependency tren `windows/386`.

## Ghi chu
Trong workspace hien tai, entrypoint CLI dang nam o `backend/cmd/client/main.go`. Neu ban muon dong bo lai theo ten bai lam `cmd/cli`, minh co the doi ten va sua README/lenh chay cho khop hoan toan.
