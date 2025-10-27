package auth

import (
	"context"
	"time"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage"
	modelAuth "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/auth"
	"github.com/ramil063/secondgodiplom/internal/hash"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
	"github.com/ramil063/secondgodiplom/internal/security/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server надстройка над стандартным gRPC сервером(авторизация)
type Server struct {
	auth.UnimplementedAuthServiceServer

	storage storage.Authenticator
	Secret  string
}

// NewAuthServer инициализация сервера авторизации, хранилища и секрета для шифрования данных
func NewAuthServer(storage storage.Authenticator, secret string) *Server {
	return &Server{
		storage: storage,
		Secret:  secret,
	}
}

// Login авторизация пользователя
// сохранение данных по токенам
func (s *Server) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	// Проверяем пользователя
	user, err := s.storage.GetUserByLogin(ctx, req.Login)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	// Проверяем пароль (bcrypt)
	if !hash.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	// Генерируем токены
	accessToken, err := jwt.GenerateAccessToken(user.ID, s.Secret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	// Сохраняем access token в БД
	accessTokenId, err := s.storage.SaveAccessToken(ctx, user.ID, accessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save refresh token")
	}

	refreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}

	// Сохраняем refresh token в БД
	err = s.storage.SaveRefreshToken(ctx, accessTokenId, refreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save refresh token")
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    modelAuth.TokenExpiredSeconds, // 30 минут
	}, nil
}

// Refresh обновление токенов на более свежие
func (s *Server) Refresh(ctx context.Context, req *auth.RefreshRequest) (*auth.RefreshResponse, error) {
	// 1. Валидируем старый refresh token
	oldTokenInfo, err := s.storage.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}

	// 2. Проверяем expiry
	if time.Now().After(oldTokenInfo.ExpiresAt) {
		return nil, status.Error(codes.Unauthenticated, "refresh token expired")
	}

	// 3. Отзываем старый refresh token
	err = s.storage.RevokeRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to revoke token")
	}

	// 5. Генерируем НОВУЮ пару токенов
	accessToken, err := jwt.GenerateAccessToken(oldTokenInfo.UserID, s.Secret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}
	accessTokenID, err := s.storage.SaveAccessToken(ctx, oldTokenInfo.UserID, accessToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save access token")
	}

	newRefreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}
	err = s.storage.SaveRefreshToken(ctx, accessTokenID, newRefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to save refresh token")
	}

	return &auth.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    modelAuth.TokenExpiredSeconds,
	}, nil
}
