# Story_UDP_Part1_Backend_Core.md

## 1. Overview
Thiết lập hạ tầng UDP Server lõi để phát đi các thông báo về chương truyện mới dưới dạng binary payload.

## 2. Acceptance Criteria (AC)
* [AC1] Server khởi tạo thành công một UDP Socket lắng nghe trên port được cấu hình (ví dụ: 9091).
* [AC2] Payload gửi đi phải được serialize từ struct `NotificationPayload` sang JSON hoặc Binary.
* [AC3] Hàm `BroadcastUpdate()` phải hoạt động không gây block luồng chính của REST API.
* [AC4] Server có khả năng xử lý lỗi nếu socket bị chiếm dụng hoặc không khởi tạo được.

## 3. Technical Requirements
* Sử dụng gói `net` của Go.
* Dữ liệu: `{"manga_id": string, "chapter": int, "title": string, "timestamp": int}`.

## 4. Task List
* [ ] Thiết kế struct `UDPPayload` cho thông báo.
* [ ] Viết hàm `InitUDPServer(port int)` sử dụng `net.ListenUDP`.
* [ ] Viết hàm `BroadcastUpdate(data UDPPayload)` thực hiện `WriteToUDP` tới danh sách các node trung gian (Bridge).
* [ ] Unit test gửi/nhận gói tin UDP đơn giản.
