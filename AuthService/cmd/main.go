package main

import (
	"AuthService/internal/config"
	"AuthService/internal/postgres"
	"AuthService/internal/usecases"
	"AuthService/pkg/grpc/auth"
	"AuthService/pkg/pg"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Warn("failed to sync logger", zap.Error(err))
		}
	}()
	zap.ReplaceGlobals(logger)

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// Подключение к базе данных
	db, err := pg.NewDB(cfg)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database connection", zap.Error(err))
		}
	}()

	// Инициализация репозиториев и сервисов
	userRepo := postgres.NewUserRepository(db, logger)
	authService := usecases.NewAuthService(userRepo, cfg, logger)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer()
	auth.RegisterAuthServiceServer(grpcServer, &auth.Server{
		AuthService: authService,
	})

	// Запуск сервера
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err), zap.String("port", cfg.GRPCPort))
	}

	logger.Info("gRPC server started", zap.String("port", cfg.GRPCPort))
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve gRPC server", zap.Error(err))
	}
}
