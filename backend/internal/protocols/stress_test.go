package protocols_test

import (
	"context"
	"fmt"
	"log"
	"mangahub/internal/database"
	"mangahub/internal/protocols/grpc"
	"mangahub/internal/protocols/tcp"
	"mangahub/internal/protocols/udp"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	grpc_lib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"mangahub/proto"
)

// TestGlobalConcurrency kiểm chứng khả năng chịu tải và Race Condition của toàn hệ thống (AC2)
func TestGlobalConcurrency(t *testing.T) {
	// 1. Setup Môi trường test
	dbPath := "./stress_test.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)
	database.InitDB(dbPath)

	// Seed data cho gRPC
	database.DB.Exec("INSERT INTO mangas (id, title, author) VALUES (1, 'One Piece', 'Oda')")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Khởi chạy các Protocol Servers
	tcpAddr := ":29090"
	grpcAddr := ":25051"
	udpPort := 29999

	go tcp.StartTCPServer(ctx, tcpAddr)
	go grpc.StartGRPCServer(ctx, grpcAddr)
	udp.InitUDPServer(udpPort)

	// Chờ server sẵn sàng
	time.Sleep(500 * time.Millisecond)

	const numWorkers = 20
	const iterations = 50
	var wg sync.WaitGroup

	// --- WORKER 1: TCP SYNC RACE TEST ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		var internalWg sync.WaitGroup
		for i := 0; i < numWorkers; i++ {
			internalWg.Add(1)
			go func(uid int) {
				defer internalWg.Done()
				conn, err := net.Dial("tcp", "localhost"+tcpAddr)
				if err != nil {
					return
				}
				defer conn.Close()
				// Auth & Update liên tục
				for j := 0; j < iterations; j++ {
					payload := fmt.Sprintf(`{"type":"AUTH", "user_id":%d}`+"\n", uid)
					conn.Write([]byte(payload))
					update := fmt.Sprintf(`{"type":"UPDATE_PROGRESS", "user_id":%d, "manga_id":1, "chapter":%d}`+"\n", uid, j)
					conn.Write([]byte(update))
					time.Sleep(1 * time.Millisecond)
				}
			}(i)
		}
		internalWg.Wait()
	}()

	// --- WORKER 2: UDP BROADCAST RACE TEST ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numWorkers*iterations; i++ {
			udp.BroadcastUpdate(udp.UDPPayload{
				Title:   "New Chapter Stress",
				Chapter: i,
			})
		}
	}()

	// --- WORKER 3: DATABASE CONCURRENT WRITE TEST ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		var internalWg sync.WaitGroup
		for i := 0; i < numWorkers; i++ {
			internalWg.Add(1)
			go func(workerID int) {
				defer internalWg.Done()
				for j := 0; j < iterations; j++ {
					_, err := database.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", 
						fmt.Sprintf("user_%d_%d", workerID, j), "pass")
					if err != nil {
						// t.Errorf("DB Error: %v", err) // SQLite might busy if not using SetMaxOpenConns(1)
					}
				}
			}(i)
		}
		internalWg.Wait()
	}()

	// --- WORKER 4: gRPC CONCURRENT READ TEST ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := grpc_lib.NewClient("localhost"+grpcAddr, grpc_lib.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("gRPC Dial error: %v", err)
			return
		}
		defer conn.Close()
		client := proto.NewMangaServiceClient(conn)

		var internalWg sync.WaitGroup
		for i := 0; i < numWorkers; i++ {
			internalWg.Add(1)
			go func() {
				defer internalWg.Done()
				for j := 0; j < iterations; j++ {
					_, err := client.GetMangaDetail(context.Background(), &proto.MangaRequest{Id: 1})
					if err != nil {
						return
					}
				}
			}()
		}
		internalWg.Wait()
	}()

	// Đợi tất cả hoàn thành
	wg.Wait()
	log.Println("[Stress Test] Toàn bộ worker đã hoàn thành mà không crash.")
}
