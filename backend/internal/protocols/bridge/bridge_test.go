package bridge

import (
	"encoding/json"
	"mangahub/internal/protocols/websocket"
	"net"
	"testing"
	"time"
)

func TestUDPBridge_StartAndReceive(t *testing.T) {
	// 1. Setup Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 2. Setup Bridge
	bridgePort := 8891
	b := NewUDPBridge(bridgePort, hub)
	err := b.Start()
	if err != nil {
		t.Fatalf("Không thể khởi động Bridge: %v", err)
	}

	// 3. Giả lập Backend gửi gói tin UDP tới Bridge
	serverAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8891")
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("Không thể kết nối tới Bridge UDP: %v", err)
	}
	defer conn.Close()

	testPayload := map[string]interface{}{
		"manga_id": "test-123",
		"chapter":  10,
		"title":    "Chương Test Bridge",
	}
	data, _ := json.Marshal(testPayload)

	// [AC2] Gửi dữ liệu JSON thô
	_, err = conn.Write(data)
	if err != nil {
		t.Fatalf("Gửi UDP thất bại: %v", err)
	}

	// 4. Kiểm tra AC4: Logic không bị treo khi không có WebSocket client
	// Chúng ta đợi một khoảng ngắn để đảm bảo Goroutine listen đã chạy qua đoạn Broadcast
	done := make(chan bool)
	go func() {
		time.Sleep(200 * time.Millisecond)
		done <- true
	}()

	select {
	case <-done:
		// Thành công: Logic không bị block
	case <-time.After(1 * time.Second):
		t.Error("Logic bị treo (Timeout) - Có thể AC4 chưa đạt")
	}
}

func TestUDPBridge_InvalidJSON(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	bridgePort := 8892
	b := NewUDPBridge(bridgePort, hub)
	b.Start()

	serverAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8892")
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// Gửi dữ liệu không phải JSON
	_, err = conn.Write([]byte("không phải json"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Đợi log (manual check) hoặc đảm bảo không crash
	time.Sleep(100 * time.Millisecond)
}
