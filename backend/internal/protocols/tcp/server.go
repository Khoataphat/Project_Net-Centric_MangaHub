package tcp

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
)

// 1. CONNECTION POOL
var (
	clientsMap = make(map[int][]net.Conn)
	mu         sync.Mutex
)

// 2. LISTENER
// StartTCPServer khởi chạy TCP Server và hỗ trợ đóng an toàn qua Context (AC1, Defensive Rules)
func StartTCPServer(ctx context.Context, addr string) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("[TCP] Lỗi khởi tạo Server: %v", err)
	}

	// Đảm bảo listener đóng khi hàm thoát
	defer listener.Close()

	log.Printf("[TCP] Sync Server đang chạy tại %s", addr)

	// Goroutine để đóng listener khi Context bị hủy (Graceful Shutdown)
	go func() {
		<-ctx.Done()
		log.Println("[TCP] Đang dừng TCP Server...")
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil // Thoát bình thường khi Shutdown
			default:
				log.Printf("[TCP] Lỗi kết nối: %v", err)
				continue
			}
		}
		go handleConnection(conn)
	}
}

// --- CÁC HÀM TIỆN ÍCH (HELPER) ---

// thêm một kết nối mới vào Pool một cách an toàn
func addConnection(userID int, conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	clientsMap[userID] = append(clientsMap[userID], conn)
}

// xóa kết nối khi Client tắt trình duyệt
func removeConnection(userID int, connToRemove net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	conns := clientsMap[userID]
	for i, conn := range conns {
		if conn == connToRemove {
			clientsMap[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

// broadcast gửi gói tin đến tất cả thiết bị của một User (TRỪ thiết bị vừa gửi)
func broadcast(userID int, message []byte, senderConn net.Conn) {
	mu.Lock()
	conns, ok := clientsMap[userID]
	if !ok || len(conns) == 0 {
		mu.Unlock()
		return
	}
	// Copy connections to avoid holding lock during Write (I/O)
	localConns := make([]net.Conn, len(conns))
	copy(localConns, conns)
	mu.Unlock()

	for _, conn := range localConns {
		// Không gửi ngược lại cho chính người vừa lật trang
		if conn != senderConn {
			_, err := conn.Write(message)
			if err != nil {
				log.Printf("[TCP] Lỗi gửi dữ liệu tới Client: %v", err)
				// Note: We don't remove here because handleConnection's defer will handle it
				// or the next heartbeat/read will fail.
			}
		}
	}
}
