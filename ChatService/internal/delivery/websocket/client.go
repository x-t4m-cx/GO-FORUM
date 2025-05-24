package websocket

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	username string
	send     chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn, username string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		username: username,
		send:     make(chan []byte),
	}
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

		msg, err := c.hub.chatUsecase.ProcessMessage(c.username, string(message))
		if err != nil {
			continue
		}

		c.hub.broadcast <- msg
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}
