package auth

import "time"

// TokenExpiredSeconds через сколько сгорит токен авторизации
var TokenExpiredSeconds int64 = 60 * 60 * 24 * 30

// Token описывает токен авторизации
type Token struct {
	Token string `json:"token"` // Токен
}

// AccessTokenData описывает данные токена пользователя
type AccessTokenData struct {
	Login                string `json:"login"`                   // Логин
	AccessToken          string `json:"access_token"`            // Токен авторизации
	AccessTokenExpiredAt int64  `json:"access_token_expired_at"` // Время истечения токена
}

// RefreshToken описывает данные токена пользователя
type RefreshToken struct {
	TokenHash       string    `json:"token_hash"`        // Хеш токена
	AccessTokenHash string    `json:"access_token_hash"` // Хеш токена авторизации
	UserID          int       `json:"user_id"`           // Пользователь
	ExpiresAt       time.Time `json:"expires_at"`        // Время истечения токена
}
