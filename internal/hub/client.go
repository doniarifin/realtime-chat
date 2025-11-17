package hub

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	conn     *websocket.Conn
	hub      *Hub
	send     chan Message
	Username string
}

func NewClient(conn *websocket.Conn, h *Hub, username string) *Client {
	return &Client{
		conn:     conn,
		hub:      h,
		send:     make(chan Message, 256),
		Username: username,
	}
}

func (c *Client) ReadMessage() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, p, err := c.conn.ReadMessage()

		if err != nil {
			log.Println(err)
			break
		}

		var msg Message
		if err := json.Unmarshal(p, &msg); err != nil {
			continue
		}
		if msg.Sender == "" {
			msg.Sender = c.Username
		}
		c.hub.Broadcast <- msg
	}
}

func (c *Client) WriteMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
