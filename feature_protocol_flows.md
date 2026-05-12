# MangaHub — Tính năng × Protocol × Flow

## Tổng quan Ports

| Protocol | Port | Vai trò |
|---|---|---|
| HTTP/REST (Gin) | 8080 | API + serve static files |
| WebSocket (ws-tcp-bridge) | 8080 `/api/ws-tcp-bridge` | Cầu nối Browser ↔ TCP |
| WebSocket (chat/notify) | 8080 `/api/ws/chat` | Broadcast thông báo chương mới |
| WebSocket (logs) | 8080 `/api/ws-logs` | Stream log server ra Admin UI |
| TCP Sync | 9090 | Đồng bộ tiến độ đọc đa thiết bị |
| UDP Notifier | 9999 | Phát sự kiện chương mới (fire & forget) |
| UDP Bridge | 8888 | Nhận UDP → chuyển tiếp vào WS Hub |
| gRPC | 50051 | Truy vấn chi tiết manga (binary) |

---

## 1. Đăng ký tài khoản
**Protocol:** `HTTP REST (POST)`

```
[auth.html] → fetch POST /api/register
  → RegisterHandler
    → ShouldBindJSON(&user)
    → auth.HashPassword(password)       // bcrypt
    → DB.Exec("INSERT INTO users ...")
  ← JSON { message: "Đăng ký thành công" }
```

---

## 2. Đăng nhập / Cấp JWT
**Protocol:** `HTTP REST (POST)`

```
[auth.html] → fetch POST /api/login
  → LoginHandler
    → ShouldBindJSON(&req)
    → DB.QueryRow("SELECT id,password FROM users ...")
    → auth.CheckPasswordHash(plain, hashed)
    → auth.GenerateJWT(userID, username)
  ← JSON { token: "eyJ..." }
[Frontend] → localStorage.setItem('token', ...)
```

---

## 3. Lấy danh sách Manga (có tìm kiếm)
**Protocol:** `HTTP REST (GET)`

```
[dashboard.html] → fetch GET /api/mangas?q=<keyword>
  → GetMangas(c)
    → c.Query("q")
    → DB.Query("SELECT ... FROM mangas WHERE title LIKE ?")
    → rows.Scan(&m)
  ← JSON [ { id, title, author, ... }, ... ]
[dashboard.html] → renderMangaCards(mangas)
```

---

## 4. Lấy chi tiết Manga theo ID
**Protocol:** `HTTP REST (GET)`

```
[reading.html] → fetch GET /api/mangas/:id
  → GetMangaByID(c)
    → c.Param("id")
    → DB.QueryRow("SELECT ... FROM mangas WHERE id = ?")
  ← JSON { id, title, author, description, thumbnail }
```

---

## 5. Đồng bộ tiến độ đọc đa thiết bị (Resume Reading)
**Protocol:** `WebSocket ↔ TCP (qua Bridge Handler)`

> Frontend không thể mở TCP trực tiếp → dùng WS làm tunnel vào TCP :9090

### 5a. Kết nối & Auth

```
[reading.html] load sync.js
  → connectTCP()
    → new WebSocket("ws://localhost:8080/api/ws-tcp-bridge")
      → TCPBridgeHandler (HTTP → Upgrade WS)
        → net.Dial("tcp", "127.0.0.1:9090")   // mở TCP nội bộ
        → go pump: TCP→WS | pump: WS→TCP       // 2 luồng song song

ws.onopen
  → ws.send({ type:"AUTH", user_id, manga_id } + "\n")
    → [TCP] handleConnection → json.Unmarshal
      → case "AUTH":
          → addConnection(userID, conn)
          → DB.QueryRow("SELECT last_chapter FROM user_progress ...")
          → conn.Write({ type:"SYNC_RESUME", chapter:N } + "\n")   // nếu có tiến độ cũ

ws.onmessage (type="SYNC_RESUME")
  → resumeModal.classList.remove('hidden')
  → [btnAcceptResume].onclick → currentChapter = payload.chapter → updateUI()

### 5c. Live Presence (Đếm số người đang đọc)
**Protocol:** `WebSocket ↔ TCP`

```
[reading.html] kết nối thành công (AUTH)
  → [TCP Server] addConnection()
    → presenceMap[mangaID][conn] = struct{}{}
    → broadcastPresence(mangaID)
      → payload: { type: "PRESENCE_UPDATE", count: N }
      → Gửi tới TẤT CẢ client đang đọc manga đó

ws.onmessage (type="PRESENCE_UPDATE")
  → document.getElementById('readerCount').innerText = payload.count

[User đóng tab / Disconnect]
  → [TCP Server] defer removeConnection()
    → delete presenceMap[mangaID][conn]
    → broadcastPresence(mangaID) (Cập nhật lại số lượng mới)
```
```

### 5b. Cập nhật tiến độ & Broadcast đa thiết bị

```
[btnNext / btnPrev].onclick
  → currentChapter ±1
  → sendUpdate(ws)
    → ws.send({ type:"UPDATE_PROGRESS", user_id, manga_id, chapter } + "\n")
      → [TCP] handleConnection
        → case "UPDATE_PROGRESS":
            → DB.Exec("INSERT INTO user_progress ... ON CONFLICT DO UPDATE ...")
            → broadcast(userID, message, senderConn)
              → conn.Write(message)   // gửi tới tất cả thiết bị KHÁC cùng user

ws.onmessage (type="UPDATE_PROGRESS", manga_id === MANGA_ID)
  → currentChapter = payload.chapter
  → updateUI()
```

---

## 6. Thông báo chương mới Real-time
**Protocol:** `UDP → Bridge → WebSocket`

### 6a. Phát thông báo (từ Admin/trigger)

```
[Admin / scan trigger]
  → udp.BroadcastUpdate(UDPPayload{ title, chapter })     // goroutine, non-blocking
    → json.Marshal(data)
    → conn.WriteToUDP(payload, bridgeAddr)   // gửi tới Bridge :8888
```

### 6b. Bridge nhận UDP → đẩy vào WebSocket Hub

```
[Bridge] UDPBridge.listen() (goroutine)
  → conn.ReadFromUDP(buf)
  → json.Unmarshal → payload["type"] = "NEW_CHAPTER"
  → hub.Broadcast(finalPayload)
    → Hub.Run() loop: hub.broadcast channel
      → client.send <- message   // gửi đến mọi WS client đang kết nối
```

### 6c. Frontend nhận và hiển thị Toast

```
[dashboard.html] load notification.js
  → DOMContentLoaded → setupWebSocket()
    → new WebSocket("ws://localhost:8080/api/ws/chat")
      → ServeWS(chatHub, c)
        → chatHub.register <- client

ws.onmessage (type="NEW_CHAPTER")
  → showNotification(data)
    → createElement('div') toast
    → container.appendChild(toast)
    → requestAnimationFrame → slide-in animation
    → setTimeout(removeToast, 5000)
```

---

## 7. Scan Manga qua gRPC (Admin Panel)
**Protocol:** `HTTP REST (GET) → gRPC (binary)`

```
[admin.html] → fetch GET /api/admin/scan-manga?id=<N>
  → ScanMangaHandler(c)
    → strconv.Atoi(idStr)
    → grpc.NewClient("127.0.0.1:50051")
    → proto.NewMangaServiceClient(conn)
    → client.GetMangaDetail(ctx, &MangaRequest{Id: N})
      → [gRPC Server] MangaServer.GetMangaDetail
          → DB.QueryRow("SELECT title,author,description FROM mangas WHERE id=?")
          → return &MangaResponse{ id, title, author, description }
  ← JSON { protocol_used:"gRPC (Binary)", data: {...} }
```

---

## 8. Stream Log Server ra Admin UI
**Protocol:** `WebSocket (Log Broadcaster)`

```
[server startup]
  → io.MultiWriter(os.Stdout, WsLogWriter{})
  → log.SetOutput(multiWriter)
  → go ServerLogs.Run()   // LogBroadcaster loop

[mọi log.Printf(...)]
  → WsLogWriter.Write(p)
    → ServerLogs.broadcast <- msg   // non-blocking select

[admin.html] → new WebSocket("ws://localhost:8080/api/ws-logs")
  → ServeLogWS(c)
    → upgrader.Upgrade → LogClient
    → ServerLogs.register <- client
    → go client.writePump()   // gửi log liên tục
    → client.readPump()       // chờ disconnect

ws.onmessage → append log line ra #terminalContent
```

---

## Sơ đồ tổng quan kiến trúc

```
Browser
  │
  ├─[HTTP]──────────► Gin Router :8080
  │                       ├── /api/register  → RegisterHandler
  │                       ├── /api/login     → LoginHandler
  │                       ├── /api/mangas    → GetMangas / GetMangaByID
  │                       ├── /api/admin/scan-manga → ScanMangaHandler
  │                       │       └──[gRPC]──► MangaServer :50051
  │                       │
  │                       ├── /api/ws-tcp-bridge → TCPBridgeHandler
  │                       │       └──[TCP]───► TCP Sync Server :9090
  │                       │                       └─ handleConnection
  │                       │                           ├── AUTH → addConnection, SYNC_RESUME
  │                       │                           └── UPDATE_PROGRESS → DB + broadcast
  │                       │
  │                       ├── /api/ws/chat   → ServeWS → Hub
  │                       │       ▲  (receives from Bridge)
  │                       │
  │                       └── /api/ws-logs  → ServeLogWS → LogBroadcaster
  │
  └─[WebSocket]──────► (3 endpoints above)

Admin / Trigger
  └─[UDP]────────────► BroadcastUpdate :9999
                            └──► Bridge :8888 (UDPBridge.listen)
                                      └──► Hub.Broadcast
```

---

## 9. Lấy danh sách chương truyện (Chapters)
**Protocol:** `HTTP REST (GET)`

```
[reading.html] → fetch GET /api/mangas/:id/chapters
  → GetMangaChapters(c)
    → c.Param("id")
    → DB.Query("SELECT ... FROM chapters WHERE manga_id = ? ORDER BY chapter_number ASC")
    → rows.Scan(&ch)
  ← JSON [ { id, manga_id, chapter_number, title }, ... ]
```

---

## 10. Lấy danh sách trang truyện (Pages)
**Protocol:** `HTTP REST (GET)`

```
[reading.html] → fetch GET /api/chapters/:chapter_id/pages
  → GetChapterPages(c)
    → c.Param("chapter_id")
    → DB.Query("SELECT ... FROM pages WHERE chapter_id = ? ORDER BY page_number ASC")
    → rows.Scan(&p)
  ← JSON [ { id, chapter_id, page_number, image_url }, ... ]
```

---

## 11. Phân quyền truy cập (RBAC Middleware)
**Protocol:** `HTTP Middleware (JWT)`

```
[Request] → Authorization Header (Bearer <token>)
  → RoleMiddleware(requiredRole)
    → jwt.Parse(tokenString)
    → Check claims["role"]
    → IF role == requiredRole OR role == "admin":
        → c.Set("userID", ...) → c.Next()
      ELSE:
        → c.AbortWithStatus(403)
```

---

## Sơ đồ tổng quan kiến trúc (Cập nhật)

```
Browser
  │
  ├─[HTTP]──────────► Gin Router :8080
  │                       ├── [Middleware] RoleMiddleware (JWT / RBAC)
  │                       │
  │                       ├── /api/register  → RegisterHandler
  │                       ├── /api/login     → LoginHandler
  │                       ├── /api/mangas    → GetMangas / GetMangaByID
  │                       ├── /api/mangas/:id/chapters → GetMangaChapters (Mới)
  │                       ├── /api/chapters/:id/pages  → GetChapterPages (Mới)
  │                       │
  │                       ├── /api/admin/scan-manga → ScanMangaHandler (Protected by RBAC)
  │                       │       └──[gRPC]──► MangaServer :50051
  │                       │
  │                       ├── /api/ws-tcp-bridge → TCPBridgeHandler
  │                       │       └──[TCP]───► TCP Sync Server :9090
  │                       │
  │                       ├── /api/ws/chat   → ServeWS → Hub
  │                       └── /api/ws-logs  → ServeLogWS → LogBroadcaster
  │
  └─[WebSocket]──────► (3 endpoints above)
```

---

## 12. Tối ưu hóa Bridge (JSON Framing)
**Protocol:** `TCP ↔ WebSocket (Optimization)`

```
[TCP Stream] → bufio.NewScanner(tcpConn)
  → scanner.Scan() (Đọc từng dòng kết thúc bằng \n)
  → ws.WriteMessage(TextMessage, line)
```
> **Mục đích:** Đảm bảo các gói tin JSON không bị dính vào nhau khi truyền tải tốc độ cao, giúp Frontend parse JSON ổn định hơn.

---

## 13. Tự động Seeding dữ liệu (MangaDex Integration)
**Protocol:** `HTTP GET (MangaDex API) → SQLite (Local)`

```
[cmd/seed/main.go]
  → getJSON("https://api.mangadex.org/manga?...")
    → Extract Title, Author, Description, Cover
    → upsertManga(db, ...)
  → Loop Chapters (limit 10):
    → getJSON("https://api.mangadex.org/manga/{id}/feed?...")
    → upsertChapter(db, ...)
    → getJSON("https://api.mangadex.org/at-home/server/{chapter_id}")
      → Build Full Image URLs
      → insertPages(db, chapterID, imageURLs)  // Transactional insert
```

---

## 14. Cơ chế Xử lý đồng thời (Concurrency)
**Protocol:** `Go Primitives (Goroutines + Channels + Mutex)`

| Component | Cơ chế | Mục đích |
|---|---|---|
| **TCP Sync** | `sync.Mutex` | Bảo vệ `clientsMap` và `presenceMap` khi nhiều thiết bị connect/disconnect cùng lúc. |
| **WS Hub** | `chan []byte` | Dùng channel `broadcast` để đẩy tin nhắn đến client mà không làm block luồng logic chính. |
| **Log Stream** | `io.MultiWriter` | Ghi log đồng thời ra cả Console và WebSocket channel. |
| **UDP Bridge** | `go listen()` | Goroutine chạy ngầm liên tục đợi gói tin UDP mà không chặn server HTTP chính. |
| **gRPC Server** | `Worker Pool` | gRPC mặc định xử lý mỗi request trên một goroutine riêng để tối ưu performance. |

---

## Sơ đồ tổng quan kiến trúc (Cập nhật cuối)

```
Browser
  │
  ├─[HTTP]──────────► Gin Router :8080
  │                       ├── [Middleware] RoleMiddleware (JWT / RBAC)
  │                       │
  │                       ├── /api/register  → RegisterHandler
  │                       ├── /api/login     → LoginHandler
  │                       ├── /api/mangas    → GetMangas / GetMangaByID
  │                       ├── /api/mangas/:id/chapters → GetMangaChapters
  │                       ├── /api/chapters/:id/pages  → GetChapterPages
  │                       │
  │                       ├── /api/admin/scan-manga → ScanMangaHandler (Protected)
  │                       │       └──[gRPC]──► MangaServer :50051
  │                       │
  │                       ├── /api/ws-tcp-bridge → TCPBridgeHandler
  │                       │       └──[TCP]───► TCP Sync Server :9090
  │                       │                       ├── Progress Sync (Peer-to-Peer)
  │                       │                       └── Live Presence (Broadcast)
  │                       │
  │                       ├── /api/ws/chat   → ServeWS → Hub (Notifications)
  │                       └── /api/ws-logs  → ServeLogWS → LogBroadcaster
  │
  └─[WebSocket]──────► (3 endpoints above)

Seeder (External)
  └─[HTTP]──────────► MangaDex API
           └────────► SQLite DB (Initial Data)
```


