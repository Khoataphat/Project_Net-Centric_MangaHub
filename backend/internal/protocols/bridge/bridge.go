package bridge

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"mangahub/internal/protocols/websocket"
)

// UDPBridge quản lý việc nhận dữ liệu từ UDP và chuyển tiếp sang WebSocket Hub
type UDPBridge struct {
	hub  *websocket.Hub
	port int
	conn *net.UDPConn
}

// NewUDPBridge tạo một instance mới của UDPBridge
func NewUDPBridge(port int, hub *websocket.Hub) *UDPBridge {
	return &UDPBridge{
		port: port,
		hub:  hub,
	}
}

// Start khởi chạy UDP Bridge (AC1)
func (b *UDPBridge) Start() error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", b.port))
	if err != nil {
		return fmt.Errorf("không thể resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("lỗi khởi tạo UDP Bridge listener: %v", err)
	}
	b.conn = conn

	log.Printf("[Bridge] UDP Bridge đang lắng nghe tại port %d", b.port)

	// [AC1] Chạy Goroutine ngầm để lắng nghe liên tục
	go b.listen()

	return nil
}

// listen là loop lắng nghe gói tin (UDPReceiver)
func (b *UDPBridge) listen() {
	defer b.conn.Close()
	buf := make([]byte, 2048)

	for {
		n, addr, err := b.conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("[Bridge] Lỗi đọc từ UDP: %v", err)
			continue
		}

		// [AC2] Parse dữ liệu thô về định dạng JSON và thêm "type" để Frontend dễ xử lý
		var payload map[string]interface{}
		if err := json.Unmarshal(buf[:n], &payload); err != nil {
			log.Printf("[Bridge] Nhận gói tin không hợp lệ từ %s: %v", addr, err)
			continue
		}

		// Thêm type "NEW_CHAPTER" theo yêu cầu của Frontend UI Story
		payload["type"] = "NEW_CHAPTER"
		finalPayload, _ := json.Marshal(payload)

		// Log chi tiết gói tin nhận được (Task List)
		log.Printf("[Bridge] Nhận thông báo từ %s: %s", addr, string(finalPayload))

		// [AC3] Đẩy vào WebSocket Hub
		// Đảm bảo logic không bị treo (AC4) được đảm bảo bởi Hub.Broadcast sử dụng channel không block nếu Hub đang bận
		b.hub.Broadcast(finalPayload)
	}
}
