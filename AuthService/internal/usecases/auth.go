package usecases

import (
	"AuthService/internal/config"
	"AuthService/internal/domain"
	"AuthService/internal/domain/models"
	"AuthService/internal/domain/repositories"
	"AuthService/pkg/jwt"
	"AuthService/pkg/password"
	"context"
	"errors"
	"go.uber.org/zap"
	"time"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (*models.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*models.TokenPair, error)
	VerifyToken(token string) (*models.TokenClaims, error)
}

type AuthServiceStruct struct {
	repo          repositories.UserRepo
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
	logger        *zap.Logger
}

func NewAuthService(userRepo repositories.UserRepo, cfg *config.Config, logger *zap.Logger) AuthService {
	return &AuthServiceStruct{
		repo:          userRepo,
		accessSecret:  cfg.AccessSecret,
		refreshSecret: cfg.RefreshSecret,
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
		logger:        logger.With(zap.String("component", "auth_service")),
	}
}

func (s *AuthServiceStruct) Register(ctx context.Context, username string, plainPassword string) error {
	s.logger.Info("registering new user", zap.String("username", username))

	_, err := s.repo.FindByUsername(ctx, username)
	if err == nil {
		s.logger.Warn("user already exists", zap.String("username", username))
		return domain.UserAlreadyExists
	} else if !errors.Is(err, domain.UserNotFound) {
		s.logger.Error("failed to check user existence",
			zap.String("username", username),
			zap.Error(err))
		return err
	}

	hashedPassword, err := password.Hash(plainPassword)
	if err != nil {
		s.logger.Error("failed to hash password",
			zap.String("username", username),
			zap.Error(err))
		return err
	}

	user := &models.User{
		Username: username,
		Password: hashedPassword,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user",
			zap.String("username", username),
			zap.Error(err))
		return err
	}

	s.logger.Info("user registered successfully",
		zap.String("username", username),
		zap.Int("user_id", user.ID))
	return nil
}

func (s *AuthServiceStruct) Login(ctx context.Context, username string, plainPassword string) (*models.TokenPair, error) {
	s.logger.Info("user login attempt", zap.String("username", username))

	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.UserNotFound) {
			s.logger.Warn("user not found during login", zap.String("username", username))
		} else {
			s.logger.Error("failed to find user during login",
				zap.String("username", username),
				zap.Error(err))
		}
		return nil, err
	}

	if err := password.Verify(user.Password, plainPassword); err != nil {
		s.logger.Warn("invalid password provided",
			zap.String("username", username),
			zap.Error(err))
		return nil, domain.InvalidData
	}

	tokens, err := s.GenerateTokens(user.ID, username)
	if err != nil {
		s.logger.Error("failed to generate tokens",
			zap.Int("user_id", user.ID),
			zap.String("username", username),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("user logged in successfully",
		zap.Int("user_id", user.ID),
		zap.String("username", username))
	return tokens, nil
}

func (s *AuthServiceStruct) Refresh(ctx context.Context, token string) (*models.TokenPair, error) {
	s.logger.Debug("refreshing tokens")

	claims, err := jwt.ValidateToken(token, s.refreshSecret)
	if err != nil {
		s.logger.Warn("invalid refresh token provided",
			zap.Error(err))
		return nil, domain.InvalidToken
	}

	s.logger.Info("refresh token validated",
		zap.Int("user_id", claims.UserID),
		zap.String("username", claims.Username))

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		s.logger.Error("failed to find user by ID during refresh",
			zap.Int("user_id", claims.UserID),
			zap.Error(err))
		return nil, err
	}

	tokens, err := s.GenerateTokens(user.ID, user.Username)
	if err != nil {
		s.logger.Error("failed to generate new tokens during refresh",
			zap.Int("user_id", user.ID),
			zap.String("username", user.Username),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("tokens refreshed successfully",
		zap.Int("user_id", user.ID),
		zap.String("username", user.Username))
	return tokens, nil
}

func (s *AuthServiceStruct) GenerateTokens(id int, username string) (*models.TokenPair, error) {
	s.logger.Debug("generating new tokens",
		zap.Int("user_id", id),
		zap.String("username", username))

	accessToken, err := jwt.GenerateToken(
		models.TokenClaims{UserID: id, Username: username},
		s.accessSecret, s.accessTTL,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.GenerateToken(
		models.TokenClaims{UserID: id, Username: username},
		s.refreshSecret, s.refreshTTL,
	)
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthServiceStruct) VerifyToken(token string) (*models.TokenClaims, error) {
	s.logger.Debug("verifying token")

	tokenClaims, err := jwt.ValidateToken(token, s.accessSecret)
	if err != nil {
		s.logger.Warn("token verification failed",
			zap.Error(err))
		return nil, err
	}

	s.logger.Debug("token verified successfully",
		zap.Int("user_id", tokenClaims.UserID),
		zap.String("username", tokenClaims.Username))
	return tokenClaims, nil
}
