package main

import (
	_ "TopicService/docs"
	"TopicService/internal/config"
	"TopicService/internal/interfaces/api"
	myHttp "TopicService/internal/interfaces/api/http"
	auth "TopicService/internal/interfaces/api/middleware"
	"TopicService/internal/interfaces/api/persistence/postgres"
	"TopicService/internal/usecases"
	pkgLogger "TopicService/pkg/logger"
	"TopicService/pkg/pg"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/x-t4m-cx/common-grpc-auth/client"
	"os"
)

func main() {
	// 1. Инициализация цветного логгера
	logger := pkgLogger.Init()

	// 2. Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}

	// 4. Подключение к БД
	db, err := pg.NewDB(cfg)
	if err != nil {
		logger.Error("Ошибка подключения к БД",
			"error", err,
			"хост", cfg.DBHost,
			"порт", cfg.DBPort,
		)
		os.Exit(1)
	}
	defer func() {
		if err := pg.CloseDB(db); err != nil {
			logger.Error("Ошибка закрытия соединения с БД", "error", err)
		}
	}()

	// 5. Инициализация репозиториев
	topicRepo := postgres.NewTopicRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	logger.Info("Репозитории инициализированы")

	// 6. Инициализация use cases
	topicUS := usecases.NewTopicUseCase(topicRepo, logger)
	commentUS := usecases.NewCommentUseCase(commentRepo, logger)
	logger.Info("Use cases инициализированы")

	// 7. Инициализация auth клиента
	authClient, err := client.New(cfg.AuthService)
	if err != nil {
		logger.Error("Ошибка создания auth клиента",
			"error", err,
			"адрес", cfg.AuthService,
		)
		os.Exit(1)
	}
	defer authClient.Close()

	// 8. Создание обработчиков
	topicHandler := myHttp.NewTopicHandler(topicUS, logger)
	commentHandler := myHttp.NewCommentHandler(commentUS, logger)
	authHandler := myHttp.NewAuthHandler(authClient, logger)
	middleware := auth.NewAuthMiddleware(authClient, logger)

	// 9. Настройка роутера
	router := gin.Default()

	// Middleware для логирования запросов
	router.Use(pkgLogger.HTTPLogMiddleware(logger))

	// Статические файлы
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Документация Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API endpoints
	api.SetupTopicRoutes(router, topicHandler, commentHandler, authHandler, middleware.Auth())

	// 10. Запуск сервера
	logger.Info("Сервер запускается", "порт", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		logger.Error("Ошибка запуска сервера", "error", err)
		os.Exit(1)
	}
}
