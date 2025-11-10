package main

import (
	"fmt"
	"log"
	"net/http"
	"realtime-chat/internal/hub"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handler(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			username = fmt.Sprintf("anon-%d", time.Now().Unix()%1000)
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}

		client := hub.NewClient(conn, h, username)
		h.Register <- client

		// Notify join
		h.Broadcast <- hub.Message{
			Type:    "join",
			Sender:  username,
			Content: fmt.Sprintf("%s joined", username),
		}

		go client.WriteMessage()
		go client.ReadMessage()
	}
}

func main() {
	h := hub.New()
	go h.Run()

	// rute websocket
	http.HandleFunc("/ws", handler(h))

	fmt.Println("WebSocket server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Println("error starting server:", err)
	}
}
