package hub

import (
	"sync"
	"time"
)

// Message is the structure broadcasted between clients
type Message struct {
	Type    string     `json:"type"`
	Sender  string     `json:"sender"`
	Room    string     `json:"room,omitempty"`
	Content string     `json:"content"`
	Time    *time.Time `json:"time"`
}

type Hub struct {
	clients    map[*Client]bool
	Register   chan *Client
	unregister chan *Client
	Broadcast  chan Message
	mu         sync.RWMutex
}

func New() *Hub {
	h := &Hub{
		clients:    make(map[*Client]bool),
		Register:   make(chan *Client),
		unregister: make(chan *Client),
		Broadcast:  make(chan Message, 256),
	}
	return h
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.mu.Lock()
			h.clients[c] = true
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
			h.mu.Unlock()
		case msg := <-h.Broadcast:
			h.mu.RLock()
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					close(c.send)
					delete(h.clients, c)
				}
			}
			h.mu.RUnlock()
		}
	}
}
