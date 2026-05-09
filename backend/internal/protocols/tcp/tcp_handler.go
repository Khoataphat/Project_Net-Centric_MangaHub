package tcp

import (
	"bufio"
	"encoding/json"
	"log"
	"mangahub/internal/models"
	"net"
)

func handleConnection(conn net.Conn) {
	var currentUserID int
	defer func() {
		if currentUserID != 0 {
			removeConnection(currentUserID, conn)
		}
		conn.Close()
	}()

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Client disconnected: %v", conn.RemoteAddr())
			return
		}

		var payload models.SyncPayload
		if err := json.Unmarshal([]byte(message), &payload); err != nil {
			log.Printf("Lỗi giải mã gói tin: %v", err)
			continue
		}

		switch payload.Type {
		case "AUTH":
			if currentUserID == 0 {
				currentUserID = payload.UserID
				addConnection(currentUserID, conn)
				log.Printf("User %d đã kết nối TCP", payload.UserID)
			}


		case "UPDATE_PROGRESS":
			// Broadcast gói tin (bao gồm cả \n) cho các thiết bị khác của cùng User
			broadcast(payload.UserID, []byte(message), conn)
			log.Printf("User %d đổi chương: Manga %d -> Chapter %d",
				payload.UserID, payload.MangaID, payload.Chapter)
		}
	}
}

