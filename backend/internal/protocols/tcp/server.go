package tcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
)

// connKey nhóm kết nối theo (userID, mangaID) để broadcast progress đúng manga
type connKey struct {
	userID  int
	mangaID int
}

// PresencePayload là gói tin gửi về client khi số người đọc thay đổi
type PresencePayload struct {
	Type    string `json:"type"`
	MangaID int    `json:"manga_id"`
	Count   int    `json:"count"`
}

var (
	mu sync.Mutex

	// clientsMap: nhóm connection theo (userID, mangaID) — dùng để sync progress
	clientsMap = make(map[connKey][]net.Conn)

	// presenceMap: manga_id → set of all connections đang đọc manga đó
	// (từ bất kỳ user nào) — dùng để đếm Live Presence
	presenceMap = make(map[int]map[net.Conn]struct{})
)

// StartTCPServer khởi chạy TCP Server
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

// ─────────────────────────────────────────────
// Progress Sync helpers (theo userID + mangaID)
// ─────────────────────────────────────────────

func addConnection(userID int, mangaID int, conn net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	// 1. Progress map
	key := connKey{userID, mangaID}
	clientsMap[key] = append(clientsMap[key], conn)

	// 2. Presence map
	if presenceMap[mangaID] == nil {
		presenceMap[mangaID] = make(map[net.Conn]struct{})
	}
	presenceMap[mangaID][conn] = struct{}{}

	log.Printf("[TCP Pool] +Conn: User %d | Manga %d | ProgressPeers=%d | Readers=%d",
		userID, mangaID, len(clientsMap[key]), len(presenceMap[mangaID]))
}

func removeConnection(userID int, mangaID int, connToRemove net.Conn) {
	mu.Lock()
	defer mu.Unlock()

	// 1. Progress map
	key := connKey{userID, mangaID}
	conns := clientsMap[key]
	for i, c := range conns {
		if c == connToRemove {
			clientsMap[key] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	// 2. Presence map
	if presenceMap[mangaID] != nil {
		delete(presenceMap[mangaID], connToRemove)
		if len(presenceMap[mangaID]) == 0 {
			delete(presenceMap, mangaID)
		}
	}

	readers := len(presenceMap[mangaID])
	log.Printf("[TCP Pool] -Conn: User %d | Manga %d | Readers=%d", userID, mangaID, readers)
}

// broadcast gửi gói tin đến các thiết bị khác của cùng (userID, mangaID)
func broadcast(userID int, mangaID int, message []byte, senderConn net.Conn) {
	mu.Lock()
	key := connKey{userID, mangaID}
	conns := clientsMap[key]
	localConns := make([]net.Conn, len(conns))
	copy(localConns, conns)
	mu.Unlock()

	sent := 0
	for _, c := range localConns {
		if c != senderConn {
			if _, err := c.Write(message); err != nil {
				log.Printf("[TCP Broadcast] Lỗi gửi tới client: %v", err)
			} else {
				sent++
			}
		}
	}
	log.Printf("[TCP Broadcast] User %d | Manga %d | Gửi đến %d thiết bị khác", userID, mangaID, sent)
}

// ─────────────────────────────────────────────
// Live Presence helpers
// ─────────────────────────────────────────────

// countReaders đếm số người đang đọc manga (gọi trong khi đang giữ lock)
func countReaders(mangaID int) int {
	return len(presenceMap[mangaID])
}

// broadcastPresence gửi PRESENCE_UPDATE tới tất cả client đang đọc mangaID
func broadcastPresence(mangaID int) {
	mu.Lock()
	count := countReaders(mangaID)
	// Lấy danh sách conn để gửi (copy ra ngoài trước khi unlock)
	readers := make([]net.Conn, 0, count)
	for c := range presenceMap[mangaID] {
		readers = append(readers, c)
	}
	mu.Unlock()

	payload := PresencePayload{
		Type:    "PRESENCE_UPDATE",
		MangaID: mangaID,
		Count:   count,
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Presence] Lỗi marshal JSON: %v", err)
		return
	}
	msg := append(jsonBytes, '\n')

	log.Printf("[Presence] Manga %d | %d người đang đọc", mangaID, count)
	for _, c := range readers {
		if _, err := c.Write(msg); err != nil {
			log.Printf("[Presence] Lỗi gửi tới client: %v", err)
		}
	}
}
