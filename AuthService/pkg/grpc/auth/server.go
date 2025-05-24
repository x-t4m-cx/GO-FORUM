package auth

import (
	"AuthService/internal/domain"
	"AuthService/internal/usecases"
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	UnimplementedAuthServiceServer
	AuthService usecases.AuthService
}

func (s *Server) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	err := s.AuthService.Register(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, domain.UserAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}
	return &RegisterResponse{Message: "user created successfully"}, nil
}

func (s *Server) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	tokens, err := s.AuthService.Login(ctx, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.UserNotFound), errors.Is(err, domain.InvalidData):
			return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
		default:
			return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
		}
	}

	return &LoginResponse{
		Message:      "login successful",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *RefreshRequest) (*RefreshResponse, error) {
	tokens, err := s.AuthService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, domain.InvalidToken):
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		case errors.Is(err, domain.UserNotFound):
			return nil, status.Errorf(codes.Unauthenticated, "user not found")
		default:
			return nil, status.Errorf(codes.Internal, "failed to refresh tokens: %v", err)
		}
	}

	return &RefreshResponse{
		Message:      "tokens refreshed successfully",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *LogoutRequest) (*LogoutResponse, error) {
	// В реальной реализации здесь должна быть логика инвалидации токена
	return &LogoutResponse{Message: "logout successful"}, nil
}

func (s *Server) VerifyToken(ctx context.Context, req *VerifyTokenRequest) (*VerifyTokenResponse, error) {
	claim, err := s.AuthService.VerifyToken(req.Token)
	if err != nil {
		return &VerifyTokenResponse{
			Valid: false,
			Error: "invalid token",
		}, nil
	}

	return &VerifyTokenResponse{
		Valid:    true,
		Username: claim.Username,
	}, nil
}
