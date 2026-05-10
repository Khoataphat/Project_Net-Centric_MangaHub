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
                                                └──► ws/chat clients
```
