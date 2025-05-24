package api

import (
	"TopicService/internal/interfaces/api/http"
	"github.com/gin-gonic/gin"
)

// @title Topic Service API
// @version 1.0
// @description This is a topic service with authentication
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func SetupTopicRoutes(router *gin.Engine,
	th *http.TopicHandler,
	ch *http.CommentHandler,
	ah *http.AuthHandler,
	authMiddleware gin.HandlerFunc) {

	// Auth routes
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", ah.Register)
		authGroup.POST("/login", ah.Login)
		authGroup.POST("/refresh", ah.Refresh)
		authGroup.POST("/logout", ah.Logout)
	}

	// Topic routes
	topicGroup := router.Group("/topics")
	{
		// Public routes
		topicGroup.GET("/", th.GetAll)
		topicGroup.GET("/:id", th.GetTopic)
		topicGroup.GET("/comments/:topic_id", ch.GetAll)

		// Protected routes
		protected := topicGroup.Use(authMiddleware)
		{
			protected.POST("/", th.CreateTopic)
			protected.PUT("/:id", th.UpdateTopic)
			protected.DELETE("/:id", th.DeleteTopic)
		}
	}

	// Comment routes
	commentGroup := router.Group("/comments")
	{
		commentGroup.GET("/:id", ch.GetComment)

		protected := commentGroup.Use(authMiddleware)
		{
			protected.POST("/", ch.CreateComment)
			protected.PUT("/:id", ch.UpdateComment)
			protected.DELETE("/:id", ch.DeleteComment)
		}
	}

}
