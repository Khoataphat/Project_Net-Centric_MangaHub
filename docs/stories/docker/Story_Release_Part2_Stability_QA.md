# Story_Release_Part2_Stability_QA.md
## 1. Overview
Kiểm thử tương tranh và độ ổn định hệ thống.
## 2. Acceptance Criteria (AC)
* [AC1] Pass `go test -race`.
* [AC2] Chạy server với flag `-race` ổn định.
* [AC3] Xử lý Error Handling khi ngắt kết nối đột ngột.
## 3. Task List
* [ ] Chạy race detector.
* [ ] Fix lỗi shared-state nếu có.
* [ ] Stress test thủ công.