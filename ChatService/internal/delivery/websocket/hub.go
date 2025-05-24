package websocket

import (
	"ChatService/internal/domain"
	"ChatService/internal/usecase"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Hub struct {
	clients     map[*Client]bool
	broadcast   chan domain.ChatMessage
	register    chan *Client
	unregister  chan *Client
	chatUsecase usecase.ChatUsecase
}

func NewHub(chatUsecase usecase.ChatUsecase) *Hub {
	return &Hub{
		broadcast:   make(chan domain.ChatMessage),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		chatUsecase: chatUsecase,
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
			msgBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			for client := range h.clients {
				select {
				case client.send <- msgBytes:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("username")
	if username == "" {
		http.Error(w, "X-Username header is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := NewClient(hub, conn, username)
	hub.register <- client

	go client.writePump()
	go client.readPump()
}
