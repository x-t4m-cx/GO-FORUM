package http

import (
	"ChatService/internal/delivery/websocket"
	"ChatService/internal/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RegisterRoutes(router *gin.Engine, hub *websocket.Hub, chatUsecase usecase.ChatUsecase) {
	handler := NewHandler(chatUsecase, hub)

	api := router.Group("/api")
	{
		api.GET("/messages", handler.GetRecentMessages)
		api.GET("/ws", func(c *gin.Context) {
			username := c.Query("username")
			if username == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "username parameter is required"})
				return
			}
			c.Request.Header.Set("username", username)
			websocket.ServeWS(hub, c.Writer, c.Request)
		})
	}
}
