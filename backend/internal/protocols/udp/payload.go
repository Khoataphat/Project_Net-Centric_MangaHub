package udp

type UDPPayload struct {
	MangaID   string `json:"manga_id"`
	Chapter   int    `json:"chapter"`
	Title     string `json:"title"`
	Timestamp int64  `json:"timestamp"`
}
