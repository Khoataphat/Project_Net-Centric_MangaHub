# Story_UDP_Part3_Frontend_UI.md

## 1. Overview
Phát triển giao diện hiển thị thông báo thời gian thực (Toast Notification) khi nhận được tín hiệu từ WebSocket.

## 2. Acceptance Criteria (AC)
* [AC1] Client tự động kết nối WebSocket khi tải trang Dashboard.
* [AC2] Khi nhận được event loại `NEW_CHAPTER`, hiển thị một Toast UI ở góc màn hình.
* [AC3] Toast phải hiển thị đầy đủ: Tên truyện và Số chương mới.
* [AC4] Toast tự động biến mất sau 5 giây hoặc khi người dùng click đóng.
* [AC5] Animation xuất hiện (Fade-in/Slide-in) mượt mà.

## 3. Technical Requirements
* CSS Framework: Tailwind CSS.
* JavaScript: WebSocket API.

## 4. Task List
* [ ] Viết hàm `setupWebSocket()` để lắng nghe event từ Bridge.
* [ ] Thiết kế Component Toast bằng Tailwind CSS (Hidden mặc định).
* [ ] Viết hàm `showNotification(data)` để inject dữ liệu vào DOM và kích hoạt animation.
* [ ] Xử lý logic `setTimeout` để tự động xóa thông báo khỏi màn hình.
