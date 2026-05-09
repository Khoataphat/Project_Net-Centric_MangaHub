package models

//TCP
type SyncPayload struct {
	Type    string `json:"type"` // "AUTH", "UPDATE_PROGRESS"
	UserID  int    `json:"user_id"`
	MangaID int    `json:"manga_id"`
	Chapter int    `json:"chapter"`
	Message string `json:"message"` // Thông tin bổ sung nếu cần
}
