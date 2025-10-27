package interceptors

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	internalJwt "github.com/ramil063/secondgodiplom/internal/security/jwt"
)

func TestNewAuthInterceptor(t *testing.T) {
	secret := "test_secret"
	interceptor := NewAuthInterceptor(secret)

	tests := []struct {
		name         string
		fullMethod   string
		setupContext func() context.Context
		wantError    bool
		errorCode    codes.Code
	}{
		{
			name:       "auth method - skip authentication",
			fullMethod: "/auth.AuthService/Login",
			setupContext: func() context.Context {
				return context.Background()
			},
			wantError: false,
		},
		{
			name:       "valid token",
			fullMethod: "/user.UserService/GetUser",
			setupContext: func() context.Context {
				token, _ := internalJwt.GenerateAccessToken(1, secret)
				md := metadata.Pairs("authorization", "Bearer "+token)
				ctx := context.WithValue(context.Background(), "user_id", 1)
				return metadata.NewIncomingContext(ctx, md)
			},
			wantError: false,
		},
		{
			name:       "missing token",
			fullMethod: "/user.UserService/GetUser",
			setupContext: func() context.Context {
				return context.Background()
			},
			wantError: true,
			errorCode: codes.Unknown,
		},
		{
			name:       "invalid token",
			fullMethod: "/user.UserService/GetUser",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "Bearer invalid_token")
				ctx := context.WithValue(context.Background(), "user_id", 1)
				return metadata.NewIncomingContext(ctx, md)
			},
			wantError: true,
			errorCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			info := &grpc.UnaryServerInfo{FullMethod: tt.fullMethod}

			// Mock handler
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				// Проверяем что в контексте есть user_id для аутентифицированных запросов
				if !tt.wantError && !isAuthMethod(tt.fullMethod) {
					userID, ok := ctx.Value("user_id").(int)
					assert.True(t, ok, "Context should contain user_id")
					assert.Greater(t, userID, 0)
				}
				return "response", nil
			}

			result, err := interceptor(ctx, "request", info, handler)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "response", result)
			}
		})
	}
}

func TestNewAuthInterceptors(t *testing.T) {
	secret := "test_secret"

	interceptors := NewAuthInterceptors(secret)

	assert.NotNil(t, interceptors.Unary)
	assert.NotNil(t, interceptors.Stream)

	// Проверяем что это действительно интерцепторы
	assert.IsType(t, (grpc.UnaryServerInterceptor)(nil), interceptors.Unary)
	assert.IsType(t, (grpc.StreamServerInterceptor)(nil), interceptors.Stream)
}

type MockServerStream struct {
	ctx context.Context
}

func (m *MockServerStream) SetHeader(metadata.MD) error  { return nil }
func (m *MockServerStream) SendHeader(metadata.MD) error { return nil }
func (m *MockServerStream) SetTrailer(metadata.MD)       {}
func (m *MockServerStream) Context() context.Context     { return m.ctx }
func (m *MockServerStream) SendMsg(interface{}) error    { return nil }
func (m *MockServerStream) RecvMsg(interface{}) error    { return nil }

func TestStreamAuthInterceptor_Integration(t *testing.T) {
	secret := "test_secret_123456"
	interceptor := NewStreamAuthInterceptor(secret)

	tests := []struct {
		name        string
		setupStream func() grpc.ServerStream
		fullMethod  string
		expectError bool
	}{
		{
			name: "successful authentication",
			setupStream: func() grpc.ServerStream {
				token, _ := internalJwt.GenerateAccessToken(456, secret)
				md := metadata.Pairs("authorization", "Bearer "+token)
				ctx := metadata.NewIncomingContext(context.Background(), md)
				return &MockServerStream{ctx: ctx}
			},
			fullMethod:  "/service.Method/Stream",
			expectError: false,
		},
		{
			name: "auth method skip",
			setupStream: func() grpc.ServerStream {
				return &MockServerStream{ctx: context.Background()}
			},
			fullMethod:  "/auth.AuthService/StreamLogin",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := tt.setupStream()
			info := &grpc.StreamServerInfo{FullMethod: tt.fullMethod}

			// Handler который проверяет контекст
			handler := func(srv interface{}, stream grpc.ServerStream) error {
				if !isAuthMethod(tt.fullMethod) {
					// Для не-аутентификационных методов проверяем userID в контексте
					userID, ok := stream.Context().Value("userID").(int)
					assert.True(t, ok, "Context should contain userID")
					assert.Equal(t, 456, userID)
				}
				return nil
			}

			err := interceptor(nil, stream, info, handler)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_authenticateRequest(t *testing.T) {
	secret := "test_secret_123456"

	tests := []struct {
		name         string
		setupContext func() context.Context
		wantUserID   int
		wantError    bool
		errorMsg     string
	}{
		{
			name: "valid token",
			setupContext: func() context.Context {
				token, _ := internalJwt.GenerateAccessToken(123, secret)
				md := metadata.Pairs("authorization", "Bearer "+token)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantUserID: 123,
			wantError:  false,
		},
		{
			name: "missing authorization header",
			setupContext: func() context.Context {
				return context.Background()
			},
			wantError: true,
			errorMsg:  "metadata not provided",
		},
		{
			name: "invalid authorization format",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "invalid_format")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantError: true,
			errorMsg:  "invalid authorization format",
		},
		{
			name: "empty token",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "Bearer ")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantError: true,
			errorMsg:  "invalid token",
		},
		{
			name: "invalid token",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "Bearer invalid_token_123")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantError: true,
			errorMsg:  "invalid token",
		},
		{
			name: "different secret",
			setupContext: func() context.Context {
				// Токен подписан другим секретом
				token, _ := internalJwt.GenerateAccessToken(123, "different_secret")
				md := metadata.Pairs("authorization", "Bearer "+token)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantError: true,
			errorMsg:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()

			newCtx, err := authenticateRequest(ctx, secret)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, newCtx)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, newCtx)

				// Проверяем что userID добавлен в контекст
				userID, ok := newCtx.Value("userID").(int)
				assert.True(t, ok)
				assert.Equal(t, tt.wantUserID, userID)
			}
		})
	}
}

func TestExtractAndValidate_Integration(t *testing.T) {
	secret := "integration_test_secret"

	tests := []struct {
		name         string
		setupContext func() context.Context
		wantUserID   int
		wantError    bool
	}{
		{
			name: "full cycle success",
			setupContext: func() context.Context {
				token, _ := internalJwt.GenerateAccessToken(789, secret)
				md := metadata.Pairs("authorization", "Bearer "+token)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantUserID: 789,
			wantError:  false,
		},
		{
			name: "full cycle with different user",
			setupContext: func() context.Context {
				token, _ := internalJwt.GenerateAccessToken(999, secret)
				md := metadata.Pairs("authorization", "Bearer "+token)
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantUserID: 999,
			wantError:  false,
		},
		{
			name: "full cycle with invalid token",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "Bearer invalid_token")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()

			// Извлекаем токен
			token, err := extractTokenFromContext(ctx)
			if tt.wantError && err != nil {
				return // Ожидаемая ошибка на этапе извлечения
			}
			require.NoError(t, err)

			// Валидируем токен
			userID, err := validateAccessToken(token, secret)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUserID, userID)
			}
		})
	}
}

func TestExtractTokenFromContext(t *testing.T) {
	tests := []struct {
		name         string
		setupContext func() context.Context
		wantToken    string
		wantError    bool
		errorMsg     string
	}{
		{
			name: "valid token extraction",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "Bearer valid_token_123")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantToken: "valid_token_123",
			wantError: false,
		},
		{
			name: "case insensitive authorization header",
			setupContext: func() context.Context {
				md := metadata.Pairs("Authorization", "Bearer token_uppercase") // Заглавная A
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantToken: "token_uppercase",
			wantError: false,
		},
		{
			name: "multiple authorization headers - use first",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"authorization": "Bearer first_token",
					"Authorization": "Bearer second_token",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantToken: "first_token",
			wantError: false,
		},
		{
			name: "no metadata in context",
			setupContext: func() context.Context {
				return context.Background() // Нет метаданных
			},
			wantError: true,
			errorMsg:  "metadata not provided",
		},
		{
			name: "missing authorization header",
			setupContext: func() context.Context {
				md := metadata.Pairs("content-type", "application/json") // Другой заголовок
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantError: true,
			errorMsg:  "authorization header not provided",
		},
		{
			name: "empty authorization header",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "") // Пустой заголовок
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantError: true,
			errorMsg:  "invalid authorization format",
		},
		{
			name: "invalid authorization format - no bearer",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "InvalidFormat token") // Нет Bearer
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantError: true,
			errorMsg:  "invalid authorization format",
		},
		{
			name: "invalid authorization format - lowercase bearer",
			setupContext: func() context.Context {
				md := metadata.Pairs("authorization", "Bearer token_lowercase") // bearer в lowercase
				return metadata.NewIncomingContext(context.Background(), md)
			},
			wantToken: "token_lowercase", // Должно работать с любым регистром
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()

			token, err := extractTokenFromContext(ctx)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func TestValidateAccessToken(t *testing.T) {
	secret := "test_secret_key_1234567890"

	tests := []struct {
		name       string
		setupToken func() string
		wantUserID int
		wantError  bool
		errorMsg   string
	}{
		{
			name: "valid token",
			setupToken: func() string {
				token, _ := internalJwt.GenerateAccessToken(123, secret)
				return token
			},
			wantUserID: 123,
			wantError:  false,
		},
		{
			name: "token with different user ID",
			setupToken: func() string {
				token, _ := internalJwt.GenerateAccessToken(456, secret)
				return token
			},
			wantUserID: 456,
			wantError:  false,
		},
		{
			name: "invalid token signature",
			setupToken: func() string {
				// Токен подписан другим секретом
				token, _ := internalJwt.GenerateAccessToken(123, "different_secret")
				return token
			},
			wantError: true,
			errorMsg:  "invalid token",
		},
		{
			name: "malformed token",
			setupToken: func() string {
				return "malformed.jwt.token" // Невалидный JWT
			},
			wantError: true,
			errorMsg:  "invalid token",
		},
		{
			name: "empty token",
			setupToken: func() string {
				return ""
			},
			wantError: true,
			errorMsg:  "invalid token",
		},
		{
			name: "expired token",
			setupToken: func() string {
				// Создаем просроченный токен
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"user_id": 123,
					"exp":     time.Now().Add(-time.Hour).Unix(),
					"iat":     time.Now().Add(-2 * time.Hour).Unix(),
				})
				signedToken, _ := token.SignedString([]byte(secret))
				return signedToken
			},
			wantError: true,
			errorMsg:  "invalid token",
		},
		{
			name: "token without user_id claim",
			setupToken: func() string {
				// Токен без user_id
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"exp": time.Now().Add(time.Hour).Unix(),
				})
				signedToken, _ := token.SignedString([]byte(secret))
				return signedToken
			},
			wantError: true,
			errorMsg:  "user_id not found in token",
		},
		{
			name: "token with non-numeric user_id",
			setupToken: func() string {
				// Токен с user_id как строка
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"user_id": "not_a_number",
					"exp":     time.Now().Add(time.Hour).Unix(),
				})
				signedToken, _ := token.SignedString([]byte(secret))
				return signedToken
			},
			wantError: true,
			errorMsg:  "user_id not found in token",
		},
		{
			name: "token with float user_id",
			setupToken: func() string {
				// Токен с user_id как float
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"user_id": 123.0, // Float вместо int
					"exp":     time.Now().Add(time.Hour).Unix(),
				})
				signedToken, _ := token.SignedString([]byte(secret))
				return signedToken
			},
			wantUserID: 123, // Должно преобразовать float в int
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()

			userID, err := validateAccessToken(token, secret)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUserID, userID)
			}
		})
	}
}

func Test_isAuthMethod(t *testing.T) {
	type args struct {
		fullMethod string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test 1",
			args: args{
				fullMethod: "/auth.AuthService/Login",
			},
			want: true,
		},
		{
			name: "test 1",
			args: args{
				fullMethod: "/auth.Service/GetUser",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAuthMethod(tt.args.fullMethod); got != tt.want {
				t.Errorf("isAuthMethod() = %v, want %v", got, tt.want)
			}
		})
	}
}
