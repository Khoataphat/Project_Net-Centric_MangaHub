package websocket

import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Cấu hình Upgrader để cho phép mọi nguồn (tránh lỗi CORS)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// ServeWS là API endpoint để Client gọi vào
func ServeWS(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Lỗi upgrade WS:", err)
		return
	}
	
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Chạy 2 Goroutine: 1 để đọc tin, 1 để ghi tin
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// Có tin nhắn mới -> Gửi vào Hub để phát cho mọi người
		c.hub.broadcast <- message 
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		c.conn.WriteMessage(websocket.TextMessage, message)
	}
}
