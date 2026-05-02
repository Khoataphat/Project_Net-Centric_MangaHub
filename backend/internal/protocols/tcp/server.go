package tcp

import (
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
func StartTCPServer(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("[TCP] Lỗi khởi tạo Server: %v", err)
	}
	defer listener.Close()
	log.Printf("[TCP] Sync Server đang chạy tại port %s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[TCP] Lỗi kết nối: %v", err)
			continue
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
	defer mu.Unlock()

	for _, conn := range clientsMap[userID] {
		// Không gửi ngược lại cho chính người vừa lật trang
		if conn != senderConn {
			_, err := conn.Write(message)
			if err != nil {
				log.Printf("[TCP] Lỗi gửi dữ liệu tới Client: %v", err)
			}
		}
	}
}
