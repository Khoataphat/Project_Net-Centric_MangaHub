package tcp

import (
	"bufio"
	"encoding/json"
	"log"
	"mangahub/internal/database"
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

				// Sau khi AUTH thành công, kiểm tra tiến độ cũ trong DB
				if payload.MangaID <= 0 {
					log.Printf("[TCP Sync] Bỏ qua Resume cho User %d: MangaID không hợp lệ (%d)", payload.UserID, payload.MangaID)
					return
				}

				var lastChapter int
				query := "SELECT last_chapter FROM user_progress WHERE user_id = ? AND manga_id = ?"
				err := database.DB.QueryRow(query, payload.UserID, payload.MangaID).Scan(&lastChapter)

				if err == nil {
					log.Printf("[TCP Sync] DB Found: User %d, Manga %d, Chapter %d", payload.UserID, payload.MangaID, lastChapter)
					// Nếu tìm thấy tiến độ cũ (>=1), gửi gói tin SYNC_RESUME để hỏi người dùng
					syncMsg := models.SyncPayload{
						Type:    "SYNC_RESUME",
						UserID:  payload.UserID,
						MangaID: payload.MangaID,
						Chapter: lastChapter,
						Message: "Old progress found",
					}
					jsonMsg, _ := json.Marshal(syncMsg)
					conn.Write(append(jsonMsg, '\n'))
					log.Printf("[TCP Sync] Sent SYNC_RESUME to User %d for Chapter %d", payload.UserID, lastChapter)
				} else {
					log.Printf("[TCP Sync] No progress found in DB for User %d, Manga %d", payload.UserID, payload.MangaID)
				}
			}


		case "UPDATE_PROGRESS":
			// 1. Lưu vào Database SQLite trước (Persistence)
			upsertSQL := `
			INSERT INTO user_progress (user_id, manga_id, last_chapter, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(user_id, manga_id) 
			DO UPDATE SET 
				last_chapter = excluded.last_chapter,
				updated_at = CURRENT_TIMESTAMP;`

			res, err := database.DB.Exec(upsertSQL, payload.UserID, payload.MangaID, payload.Chapter)
			if err != nil {
				log.Printf("[TCP DB Error] LỖI GHI DB cho User %d: %v", payload.UserID, err)
			} else {
				rows, _ := res.RowsAffected()
				log.Printf("[TCP DB Success] User %d - Manga %d - Chương %d (RowsAffected: %d)",
					payload.UserID, payload.MangaID, payload.Chapter, rows)
			}

			// 2. Broadcast gói tin (bao gồm cả \n) cho các thiết bị khác của cùng User
			broadcast(payload.UserID, []byte(message), conn)
		}
	}
}

