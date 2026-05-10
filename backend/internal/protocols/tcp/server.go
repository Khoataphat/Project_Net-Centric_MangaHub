package tcp

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
)

// connKey nhóm kết nối theo (userID, mangaID) để broadcast đúng manga
type connKey struct {
	userID  int
	mangaID int
}

// 1. CONNECTION POOL - key theo (userID, mangaID)
var (
	clientsMap = make(map[connKey][]net.Conn)
	mu         sync.Mutex
)

// 2. LISTENER
func StartTCPServer(ctx context.Context, addr string) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("[TCP] Lỗi khởi tạo Server: %v", err)
	}
	defer listener.Close()

	log.Printf("[TCP] Sync Server đang chạy tại %s", addr)

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
				return nil
			default:
				log.Printf("[TCP] Lỗi kết nối: %v", err)
				continue
			}
		}
		go handleConnection(conn)
	}
}

// addConnection thêm kết nối theo key (userID, mangaID)
func addConnection(userID int, mangaID int, conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	key := connKey{userID, mangaID}
	clientsMap[key] = append(clientsMap[key], conn)
	log.Printf("[TCP Pool] +Conn: User %d | Manga %d | Total=%d", userID, mangaID, len(clientsMap[key]))
}

// removeConnection xóa kết nối khi client ngắt kết nối
func removeConnection(userID int, mangaID int, connToRemove net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	key := connKey{userID, mangaID}
	conns := clientsMap[key]
	for i, conn := range conns {
		if conn == connToRemove {
			clientsMap[key] = append(conns[:i], conns[i+1:]...)
			log.Printf("[TCP Pool] -Conn: User %d | Manga %d | Total=%d", userID, mangaID, len(clientsMap[key]))
			break
		}
	}
}

// broadcast gửi gói tin đến tất cả thiết bị của cùng (userID, mangaID) - TRỪ thiết bị gửi
func broadcast(userID int, mangaID int, message []byte, senderConn net.Conn) {
	mu.Lock()
	key := connKey{userID, mangaID}
	conns, ok := clientsMap[key]
	if !ok || len(conns) == 0 {
		mu.Unlock()
		return
	}
	localConns := make([]net.Conn, len(conns))
	copy(localConns, conns)
	mu.Unlock()

	sent := 0
	for _, conn := range localConns {
		if conn != senderConn {
			_, err := conn.Write(message)
			if err != nil {
				log.Printf("[TCP] Lỗi gửi tới client: %v", err)
			} else {
				sent++
			}
		}
	}
	log.Printf("[TCP Broadcast] User %d | Manga %d | Gửi đến %d thiết bị khác", userID, mangaID, sent)
}
