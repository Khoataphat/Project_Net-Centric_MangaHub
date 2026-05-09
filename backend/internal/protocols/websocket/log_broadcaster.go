package websocket

import (
	"log"

	"github.com/gin-gonic/gin"
	gws "github.com/gorilla/websocket"
)

type LogBroadcaster struct {
	clients    map[*LogClient]bool
	broadcast  chan []byte
	register   chan *LogClient
	unregister chan *LogClient
}

type LogClient struct {
	broadcaster *LogBroadcaster
	conn        *gws.Conn
	send        chan []byte
}

var ServerLogs = NewLogBroadcaster()

func NewLogBroadcaster() *LogBroadcaster {
	return &LogBroadcaster{
		clients:    make(map[*LogClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *LogClient),
		unregister: make(chan *LogClient),
	}
}

func (b *LogBroadcaster) Run() {
	for {
		select {
		case client := <-b.register:
			b.clients[client] = true
		case client := <-b.unregister:
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client.send)
			}
		case message := <-b.broadcast:
			for client := range b.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(b.clients, client)
				}
			}
		}
	}
}

type WsLogWriter struct{}

func (w WsLogWriter) Write(p []byte) (n int, err error) {
	msg := append([]byte(nil), p...)
	select {
	case ServerLogs.broadcast <- msg:
	default:
	}
	return len(p), nil
}

func ServeLogWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Lỗi upgrade WS logs:", err)
		return
	}

	client := &LogClient{
		broadcaster: ServerLogs,
		conn:        conn,
		send:        make(chan []byte, 256),
	}
	client.broadcaster.register <- client

	go client.writePump()
	client.readPump()
}

func (c *LogClient) readPump() {
	defer func() {
		c.broadcaster.unregister <- c
		_ = c.conn.Close()
	}()

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (c *LogClient) writePump() {
	defer c.conn.Close()
	for {
		message, ok := <-c.send
		if !ok {
			_ = c.conn.WriteMessage(gws.CloseMessage, []byte{})
			return
		}
		if err := c.conn.WriteMessage(gws.TextMessage, message); err != nil {
			return
		}
	}
}
