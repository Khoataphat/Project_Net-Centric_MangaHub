package websocket

type Hub struct {
	// Danh sách các client đang kết nối
	clients map[*Client]bool

	// Kênh (channel) nhận tin nhắn từ client để phát đi (Broadcast)
	broadcast chan []byte

	// Kênh xử lý khi có client mới tham gia
	register chan *Client

	// Kênh xử lý khi có client thoát
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			// Lặp qua tất cả client đang online và gửi tin nhắn
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
