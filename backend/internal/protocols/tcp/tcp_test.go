package tcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mangahub/internal/database"
	"mangahub/internal/models"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Khởi tạo DB giả lập cho test
	dbFile := "test_tcp.db"
	database.InitDB(dbFile)
	defer os.Remove(dbFile)

	os.Exit(m.Run())
}

func TestTCPConcurrency(goTest *testing.T) {
	port := ":9091"
	go func() {
		if err := StartTCPServer(context.Background(), port); err != nil {
			log.Printf("Test server error: %v", err)
		}
	}()

	// Chờ server khởi động
	time.Sleep(100 * time.Millisecond)

	const numClients = 50
	var wg sync.WaitGroup
	wg.Add(numClients)

	for i := 1; i <= numClients; i++ {
		go func(userID int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", "localhost"+port)
			if err != nil {
				goTest.Errorf("User %d: Không thể kết nối: %v", userID, err)
				return
			}
			defer conn.Close()

			// Auth
			auth := models.SyncPayload{Type: "AUTH", UserID: userID}
			bytes, _ := json.Marshal(auth)
			fmt.Fprintf(conn, "%s\n", string(bytes))

			// Gửi update
			update := models.SyncPayload{Type: "UPDATE_PROGRESS", UserID: userID, MangaID: 1, Chapter: 5}
			bytes, _ = json.Marshal(update)
			fmt.Fprintf(conn, "%s\n", string(bytes))

			// Đọc phản hồi (nếu có broadcast từ thiết bị khác - ở đây ta chỉ có 1 thiết bị mỗi user nên broadcast sẽ không gửi cho chính nó)
			// Để test broadcast thực sự, ta cần 2 connection cho cùng 1 user
		}(i)
	}

	wg.Wait()
}

func TestTCPBroadcastStability(goTest *testing.T) {
	port := ":9092"
	go func() {
		if err := StartTCPServer(context.Background(), port); err != nil {
			log.Printf("Test server error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	userID := 999
	
	// Client 1
	c1, _ := net.Dial("tcp", "localhost"+port)
	defer c1.Close()
	auth1, _ := json.Marshal(models.SyncPayload{Type: "AUTH", UserID: userID})
	fmt.Fprintf(c1, "%s\n", string(auth1))

	// Client 2
	c2, _ := net.Dial("tcp", "localhost"+port)
	defer c2.Close()
	auth2, _ := json.Marshal(models.SyncPayload{Type: "AUTH", UserID: userID})
	fmt.Fprintf(c2, "%s\n", string(auth2))

	time.Sleep(50 * time.Millisecond)

	// Client 1 gửi update -> Client 2 nhận được
	update := models.SyncPayload{Type: "UPDATE_PROGRESS", UserID: userID, MangaID: 10, Chapter: 20}
	bytes, _ := json.Marshal(update)
	fmt.Fprintf(c1, "%s\n", string(bytes))

	// Đọc từ c2
	reader := bufio.NewReader(c2)
	msg, err := reader.ReadString('\n')
	if err != nil {
		goTest.Fatalf("Lỗi đọc từ Client 2: %v", err)
	}

	var received models.SyncPayload
	json.Unmarshal([]byte(msg), &received)

	if received.Chapter != 20 {
		goTest.Errorf("Mong đợi Chapter 20, nhận được %d", received.Chapter)
	}
}
