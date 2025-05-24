package grpc

import (
	"AuthService/internal/usecases"
	"AuthService/pkg/grpc/auth"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	auth.UnimplementedAuthServiceServer
	authService usecases.AuthService
}

func New(authService usecases.AuthService) AuthServer {
	return AuthServer{
		authService: authService,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	err := s.authService.Register(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &auth.RegisterResponse{}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	tokens, err := s.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &auth.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *auth.RefreshRequest) (*auth.RefreshResponse, error) {
	tokens, err := s.authService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &auth.RefreshResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthServer) VerifyToken(ctx context.Context, req *auth.VerifyTokenRequest) (*auth.VerifyTokenResponse, error) {
	claims, err := s.authService.VerifyToken(req.Token)
	if err != nil {
		return &auth.VerifyTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}
	return &auth.VerifyTokenResponse{
		Valid:    true,
		Username: claims.Username,
	}, nil
}
