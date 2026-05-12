# Demo Script - MangaHub Multi-Machine LAN (AC3)

Tài liệu này hướng dẫn cách chạy Demo hệ thống MangaHub trên hai máy tính khác nhau trong cùng một mạng LAN (WiFi/Ethernet).

## 1. Chuẩn bị (Máy A - Server)

### Bước 1: Xác định IP của Máy A
Mở Terminal/CMD và gõ:
```bash
ipconfig   # Windows
# hoặc
ifconfig   # Linux/Mac
```
Tìm dòng `IPv4 Address`. Ví dụ: `192.168.1.15`.

### Bước 2: Khởi chạy Server bằng Docker
Đảm bảo bạn đang ở thư mục gốc của dự án.
```bash
docker compose up --build
```
Hệ thống sẽ khởi tạo 5 cổng: 8080 (HTTP), 9090 (TCP), 9999 (UDP), 8888 (Bridge), 50051 (gRPC).

---

## 2. Kiểm thử từ Máy B (Client)

Đảm bảo Máy B kết nối cùng mạng với Máy A. Thay thế `<SERVER_IP>` bằng IP ở Bước 1.

### Kịch bản 1: Kiểm tra HTTP API
Mở trình duyệt hoặc dùng Curl:
```bash
curl http://<SERVER_IP>:8080/api/mangas
```
**Kết quả mong đợi:** Nhận được danh sách Manga (JSON).

### Kịch bản 2: Đồng bộ tiến độ (TCP)
Sử dụng CLI Client hoặc công cụ Telnet/NC để giả lập:
```bash
# Giả lập Auth
echo '{"type":"AUTH", "user_id":1}' | nc <SERVER_IP> 9090
```
**Kết quả mong đợi:** Server log nhận được kết nối từ IP Máy B.

### Kịch bản 3: Nhận thông báo Realtime (WebSocket/Frontend)
1. Mở file `frontend/dashboard.html` trên Máy B.
2. Sửa IP trong code JavaScript (nếu có) từ `localhost` thành `<SERVER_IP>`.
3. Quan sát góc phải màn hình.

### Kịch bản 4: Kiểm tra gRPC Metadata
Sử dụng công cụ `grpcurl` (nếu có) hoặc CLI Client:
```bash
grpcurl -plaintext <SERVER_IP>:50051 proto.MangaService/GetMangaDetail
```

---

## 3. Luồng Demo End-to-End (Chốt)

1. **Máy A**: Chạy Server.
2. **Máy B**: Mở Dashboard.
3. **Máy A**: Thực hiện một hành động (VD: Add Manga qua Postman/Curl).
4. **Máy B**: Thấy Toast Notification hiện lên (Realtime qua UDP -> Bridge -> WS).
5. **Máy B**: Đọc truyện và đổi chương.
6. **Máy A**: Log server hiển thị tiến độ được sync qua TCP.

---

## Lưu ý quan trọng
- **Firewall**: Đảm bảo Firewall trên Máy A không chặn các cổng 8080, 9090, 9999, 8888, 50051.
- **Network**: Hai máy phải ping được nhau.
