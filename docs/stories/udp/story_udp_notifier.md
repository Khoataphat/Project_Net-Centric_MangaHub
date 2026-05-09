# story_udp_notifier.md

# PRD — UDP Notifier System for MangaHub

## 1. Overview

### Feature Name

UDP New Chapter Notifier

### Objective

Xây dựng hệ thống thông báo chương truyện mới theo thời gian thực với độ trễ thấp bằng giao thức UDP. Hệ thống cho phép Backend phát (broadcast) thông báo nhanh đến tầng trung gian và hiển thị Toast Notification trên giao diện Web mà không cần reload trang.

### Business Value

* Tăng tính realtime cho nền tảng MangaHub.
* Cải thiện retention nhờ người dùng được cập nhật chương mới ngay lập tức.
* Demonstrate khả năng tích hợp đa giao thức trong kiến trúc hệ thống:

  * REST API
  * TCP
  * UDP
  * WebSocket
  * gRPC

---

# 2. User Story

## Primary User Story

> Là một người đọc truyện, tôi muốn nhận được thông báo tức thời (Toast notification) ngay khi có chương truyện mới mà không cần phải tải lại trang.

---

# 3. Product Goals

## Functional Goals

* Server phát thông báo chương mới bằng UDP packet.
* Frontend nhận được notification gần realtime.
* Hiển thị Toast UI đẹp mắt và tự động biến mất.

## Non-functional Goals

* Độ trễ thấp.
* Không block REST API chính.
* Hệ thống chịu được nhiều notification liên tiếp.
* Chấp nhận mất packet để đổi lấy tốc độ.

---

# 4. System Architecture

## Proposed Flow

```text
[Admin Uploads New Chapter]
              |
              v
     [Go REST API Service]
              |
              v
   [UDP Notifier Goroutine]
              |
      UDP Broadcast Packet
              |
              v
 [UDP-to-WebSocket Bridge Service]
              |
          WebSocket
              |
              v
      [Browser Frontend]
              |
              v
      [Toast Notification]
```

---

# 5. Technical Design

## 5.1 UDP Packet Structure

### Payload Format (JSON)

```json
{
  "manga_title": "One Piece",
  "chapter": 1145
}
```

## Data Rules

| Field       | Type   | Required | Description    |
| ----------- | ------ | -------- | -------------- |
| manga_title | string | Yes      | Tên manga      |
| chapter     | number | Yes      | Số chapter mới |

---

# 6. Acceptance Criteria

## Scenario 1 — Backend UDP Broadcast

### Given

Server MangaHub đang chạy.

### When

Có sự kiện `"New Chapter"` được trigger.

### Then

Backend phải:

* Khởi tạo UDP connection thành công.
* Broadcast UDP packet tới port cấu hình sẵn.
* Gửi JSON payload hợp lệ.

### Expected Payload

```json
{
  "manga_title": "Naruto",
  "chapter": 701
}
```

### Technical Validation

* UDP Server chạy trên port cố định (VD: `9000`).
* Packet được gửi thành công bằng `net.UDPConn`.
* Không crash nếu client không tồn tại.
* Có logging khi gửi thất bại.

---

## Scenario 2 — Frontend Realtime Notification

### Problem

Browser không hỗ trợ UDP socket trực tiếp.

### Proposed Solution

Triển khai tầng trung gian:

## Option A — UDP-to-WebSocket Bridge (Recommended)

### Flow

```text
UDP Packet
    ↓
Bridge Service
    ↓
WebSocket Push
    ↓
Browser Client
```

### Responsibilities

Bridge Service:

* Lắng nghe UDP port.
* Parse JSON packet.
* Push message qua WebSocket tới browser clients.

### Why Recommended

| Advantage          | Reason                   |
| ------------------ | ------------------------ |
| Realtime           | Độ trễ thấp              |
| Browser Compatible | Browser hỗ trợ WebSocket |
| Scalable           | Dễ mở rộng nhiều clients |

---

## Option B — UDP-to-SSE (Alternative)

### Flow

```text
UDP Packet
    ↓
SSE Gateway
    ↓
Server-Sent Events
    ↓
Browser
```

### Limitation

* Chỉ one-way communication.
* Ít flexible hơn WebSocket.

---

## Frontend Acceptance

### Given

Người dùng đang mở MangaHub Web App.

### When

Bridge Service nhận được UDP notification.

### Then

Frontend phải:

* Nhận được event realtime.
* Parse đúng dữ liệu JSON.
* Trigger Toast notification.

---

## Scenario 3 — UI Toast Notification

### UI Behavior

* Toast xuất hiện ở góc phải phía trên màn hình.
* Hiển thị:

  * Manga title
  * Chapter mới
* Tự động biến mất sau 5 giây.

---

## Example UI

```text
🔔 New Chapter Released!

One Piece - Chapter 1145
```

---

## UI Requirements

| Requirement | Value                       |
| ----------- | --------------------------- |
| Framework   | TailwindCSS hoặc Ant Design |
| Duration    | 5 seconds                   |
| Position    | Top-right                   |
| Animation   | Fade in / Fade out          |
| Dismiss     | Auto dismiss                |

---

# 7. Technical Constraints

## Protocol Constraint

* Bắt buộc sử dụng UDP cho tầng notifier.
* Chấp nhận:

  * Packet loss
  * Out-of-order delivery
* Ưu tiên:

  * Speed
  * Low latency

---

## Concurrency Constraint

### Requirement

UDP listener phải chạy trên Goroutine riêng.

### Reason

Không làm block:

* REST API
* TCP Sync Service
* WebSocket Hub

### Expected Pattern

```text
main.go
 ├── HTTP Server Goroutine
 ├── TCP Server Goroutine
 ├── WebSocket Hub Goroutine
 └── UDP Notifier Goroutine
```

---

## Defensive Programming Constraint

### Required Error Handling

| Case                | Required Handling             |
| ------------------- | ----------------------------- |
| Port already in use | Log error + graceful shutdown |
| Invalid JSON        | Ignore packet                 |
| Oversized packet    | Reject packet                 |
| Empty payload       | Ignore                        |
| UDP read failure    | Retry loop                    |
| Bridge disconnected | Reconnect                     |

---

## Packet Size Limitation

### Requirement

Giới hạn kích thước UDP packet.

### Recommendation

```text
Max UDP Payload = 1024 bytes
```

### Reason

* Tránh buffer overflow.
* Tránh fragmentation.
* Giảm memory pressure.

---

# 8. Suggested Backend Architecture

## Recommended Folder Structure

```text
internal/
 └── protocols/
      └── udp/
           ├── notifier.go
           ├── packet.go
           ├── broadcaster.go
           └── bridge.go
```

---

# 9. API / Event Contracts

## Event Name

```text
NEW_CHAPTER_RELEASED
```

## Internal Event Flow

```text
Admin uploads chapter
        ↓
Manga Service
        ↓
NotifyNewChapter(data)
        ↓
UDP Broadcast
        ↓
Bridge
        ↓
Frontend Toast
```

---

# 10. Observability & Logging

## Required Logs

### Backend

```text
[UDP] Notification sent: One Piece Chapter 1145
```

### Error

```text
[UDP][ERROR] Failed to send packet
```

### Bridge

```text
[Bridge] Forwarded event to 12 websocket clients
```

---

# 11. Security Considerations

## Risks

UDP dễ bị:

* Spoof packet
* Flood packet
* Invalid payload spam

## Mitigation

* Chỉ bind localhost/internal network.
* Validate JSON schema.
* Giới hạn packet size.
* Optional secret token trong payload nội bộ.

---

# 12. Performance Expectations

| Metric                     | Target         |
| -------------------------- | -------------- |
| Notification latency       | < 1 second     |
| Packet size                | < 1KB          |
| Concurrent browser clients | 100+           |
| Notification throughput    | 100 events/min |

---

# 13. Future Enhancements

## Possible Improvements

* Retry queue cho missed notifications.
* Notification history.
* User-specific subscriptions.
* Kafka/NATS integration.
* Push notification mobile.
* Redis Pub/Sub bridge.

---

# 14. Task List — Assigned to Person B

## Backend Tasks

* [ ] Khởi tạo UDP Server trong Go.
* [ ] Tạo UDP Broadcaster Goroutine.
* [ ] Viết hàm `NotifyNewChapter(data)`.
* [ ] Serialize JSON payload.
* [ ] Thêm logging và error handling.
* [ ] Validate packet size.
* [ ] Handle graceful shutdown.

---

## Bridge Tasks

* [ ] Xây dựng UDP-to-WebSocket Bridge.
* [ ] Parse UDP packet.
* [ ] Broadcast tới WebSocket clients.
* [ ] Handle reconnect logic.

---

## Frontend Tasks

* [ ] Kết nối WebSocket từ Browser.
* [ ] Lắng nghe event realtime.
* [ ] Parse notification payload.
* [ ] Hiển thị Toast UI bằng Tailwind hoặc Ant Design.
* [ ] Auto dismiss sau 5 giây.
* [ ] Thêm animation fade in/out.

---

# 15. Definition of Done (DoD)

Feature được xem là hoàn thành khi:

* UDP notifier hoạt động ổn định.
* Browser nhận được notification realtime.
* Toast UI hiển thị đúng.
* Không block REST API.
* Có xử lý lỗi đầy đủ.
* Pass toàn bộ Acceptance Criteria.
* Không xảy ra race-condition khi concurrent events.
