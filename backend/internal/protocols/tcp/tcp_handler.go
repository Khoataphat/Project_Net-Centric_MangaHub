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
	var currentMangaID int

	defer func() {
		if currentUserID != 0 {
			removeConnection(currentUserID, currentMangaID, conn)
			// Broadcast presence sau khi user ngắt kết nối
			if currentMangaID > 0 {
				broadcastPresence(currentMangaID)
			}
		}
		conn.Close()
		log.Printf("[TCP] Connection closed: %v", conn.RemoteAddr())
	}()

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("[TCP] Client disconnected: %v", conn.RemoteAddr())
			return
		}

		var payload models.SyncPayload
		if err := json.Unmarshal([]byte(message), &payload); err != nil {
			log.Printf("[TCP] Lỗi giải mã gói tin: %v", err)
			continue
		}

		switch payload.Type {

		case "AUTH":
			if currentUserID != 0 {
				break
			}

			currentUserID = payload.UserID
			currentMangaID = payload.MangaID
			addConnection(currentUserID, currentMangaID, conn)
			log.Printf("[TCP Auth] User %d kết nối (MangaID=%d)", payload.UserID, payload.MangaID)

			if payload.MangaID <= 0 {
				log.Printf("[TCP Sync] Bỏ qua Resume cho User %d: MangaID không hợp lệ", payload.UserID)
				break
			}

			// ── Bước 1: Gửi SYNC_RESUME nếu có tiến độ cũ ──
			var lastChapter int
			query := "SELECT last_chapter FROM user_progress WHERE user_id = ? AND manga_id = ?"
			dbErr := database.DB.QueryRow(query, payload.UserID, payload.MangaID).Scan(&lastChapter)
			if dbErr == nil && lastChapter > 0 {
				log.Printf("[TCP Sync] Found progress: User %d, Manga %d, Chapter %d",
					payload.UserID, payload.MangaID, lastChapter)
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
				log.Printf("[TCP Sync] No progress found for User %d, Manga %d", payload.UserID, payload.MangaID)
			}

			// ── Bước 2: Broadcast PRESENCE_UPDATE cho tất cả người đọc manga đó ──
			broadcastPresence(payload.MangaID)

		case "UPDATE_PROGRESS":
			if payload.UserID <= 0 || payload.MangaID <= 0 || payload.Chapter <= 0 {
				log.Printf("[TCP DB] Bỏ qua UPDATE_PROGRESS không hợp lệ: %+v", payload)
				continue
			}

			// 1. Lưu vào Database
			upsertSQL := `
			INSERT INTO user_progress (user_id, manga_id, last_chapter, updated_at)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(user_id, manga_id) 
			DO UPDATE SET 
				last_chapter = excluded.last_chapter,
				updated_at   = CURRENT_TIMESTAMP;`

			res, err := database.DB.Exec(upsertSQL, payload.UserID, payload.MangaID, payload.Chapter)
			if err != nil {
				log.Printf("[TCP DB Error] User %d Manga %d: %v", payload.UserID, payload.MangaID, err)
			} else {
				rows, _ := res.RowsAffected()
				log.Printf("[TCP DB OK] User %d | Manga %d | Chapter %d | rows=%d",
					payload.UserID, payload.MangaID, payload.Chapter, rows)
			}

			// 2. Broadcast
			broadcast(payload.UserID, payload.MangaID, []byte(message), conn)
		}
	}
}
