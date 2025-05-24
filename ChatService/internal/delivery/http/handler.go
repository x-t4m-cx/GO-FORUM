package http

import (
	"ChatService/internal/delivery/websocket"
	"ChatService/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	chatUsecase usecase.ChatUsecase
	hub         *websocket.Hub
}

func NewHandler(chatUsecase usecase.ChatUsecase, hub *websocket.Hub) *Handler {
	return &Handler{
		chatUsecase: chatUsecase,
		hub:         hub,
	}
}

func (h *Handler) GetRecentMessages(c *gin.Context) {
	messages, err := h.chatUsecase.GetRecentMessages(50)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
