package interceptors

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptors обработка авторизации
type AuthInterceptors struct {
	Unary  grpc.UnaryServerInterceptor
	Stream grpc.StreamServerInterceptor
}

// NewAuthInterceptors инициализация основной структуры авторизации
func NewAuthInterceptors(secret string) *AuthInterceptors {
	return &AuthInterceptors{
		Unary:  NewAuthInterceptor(secret),
		Stream: NewStreamAuthInterceptor(secret),
	}
}

// Обертка для ServerStream с измененным контекстом
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context возврат контекста
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// NewAuthInterceptor инициализация простого интерсептора авторизации
func NewAuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Пропускаем аутентификационные методы
		if isAuthMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		newCtx, err := authenticateRequest(ctx, secret)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// NewStreamAuthInterceptor инициализация стримингового интерсептора авторизации
func NewStreamAuthInterceptor(secret string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Пропускаем аутентификационные методы
		if isAuthMethod(info.FullMethod) {
			return handler(srv, ss)
		}
		newCtx, err := authenticateRequest(ss.Context(), secret)
		if err != nil {
			return err
		}

		// Создаем обертку для stream с новым контекстом
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}
		return handler(srv, wrappedStream)
	}
}

func authenticateRequest(ctx context.Context, secret string) (context.Context, error) {
	token, err := extractTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := validateAccessToken(token, secret)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, "userID", userID), nil
}

// Пропускаем методы аутентификации
func isAuthMethod(fullMethod string) bool {
	authMethods := map[string]bool{
		"/auth.AuthService/Login":            true,
		"/auth.AuthService/StreamLogin":      true,
		"/auth.RegistrationService/Register": true,
		"/auth.AuthService/Refresh":          true,
	}
	return authMethods[fullMethod]
}

// Извлечение токена из метаданных
func extractTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("metadata not provided")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", fmt.Errorf("authorization header not provided")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid authorization format")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

func validateAccessToken(tokenString, secret string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id not found in token")
	}

	return int(userID), nil
}
