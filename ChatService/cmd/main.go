package main

import (
	"ChatService/internal/config"
	"ChatService/internal/delivery/http"
	"ChatService/internal/delivery/websocket"
	mongo "ChatService/internal/repository"
	"ChatService/internal/usecase"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func main() {
	cfg := config.LoadConfig()

	mongoRepo, err := mongo.NewMongoRepository(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	chatUsecase := usecase.NewChatUsecase(mongoRepo, cfg.MessageLifetime)
	hub := websocket.NewHub(chatUsecase)
	go hub.Run()
	go startCleanupScheduler(chatUsecase, 10*time.Second)
	router := gin.Default()
	router.Use(cors.Default())
	router.Static("/static", "./static")
	http.RegisterRoutes(router, hub, chatUsecase)

	log.Printf("Starting server on %s", cfg.ServerAddress)
	if err := router.Run(":" + cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
func startCleanupScheduler(chatUsecase usecase.ChatUsecase, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		log.Println("Starting expired messages cleanup...")
		if err := chatUsecase.CleanupExpiredMessages(); err != nil {
			log.Printf("Failed to cleanup expired messages: %v", err)
		} else {
			log.Println("Expired messages cleanup completed")
		}
	}
}
