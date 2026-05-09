# Story_UDP_Part2_Bridge_Service.md

## 1. Overview
Xây dựng thành phần trung gian (Bridge) để nhận gói tin UDP từ Backend và chuyển tiếp tới trình duyệt thông qua WebSocket.

## 2. Acceptance Criteria (AC)
* [AC1] Bridge phải chạy một Goroutine ngầm để lắng nghe liên tục gói tin từ UDP Server.
* [AC2] Khi nhận được gói tin UDP, Bridge phải parse thành công dữ liệu thô về định dạng JSON.
* [AC3] Dữ liệu sau khi parse phải được đẩy vào `WebSocket Hub`.
* [AC4] Phải đảm bảo logic không bị treo nếu không có WebSocket client nào đang kết nối.

## 3. Technical Requirements
* Kết nối giữa Backend và Bridge: UDP.
* Kết nối giữa Bridge và Client: WebSocket (Gorilla WebSocket).

## 4. Task List
* [ ] Tạo Goroutine `UDPReceiver` lắng nghe gói tin từ Backend.
* [ ] Kết nối `UDPReceiver` với `WebSocket Hub` hiện có.
* [ ] Xử lý logic chuyển đổi: `UDP Packet -> Internal Channel -> WebSocket Broadcast`.
* [ ] Log chi tiết các gói tin nhận được để debug trên Console.
