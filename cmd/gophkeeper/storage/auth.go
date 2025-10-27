package storage

import (
	"context"

	authModel "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/auth"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/user"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/auth"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

// UserGetter интерфейс описывающий работу с получением пользователя
type UserGetter interface {
	GetUserByLogin(ctx context.Context, login string) (*user.User, error)
}

// Loginer интерфейс описывающий работу с авторизацией пользователя
type Loginer interface {
	LoginUser(ctx context.Context) (string, string, error)
}

// TokenSaver интерфейс описывающий работу с сохранением токенов
type TokenSaver interface {
	SaveAccessToken(ctx context.Context, userID int, token string) (int, error)
	SaveRefreshToken(ctx context.Context, accessTokenId int, token string) error
}

// TokenGetter интерфейс описывающий работу с получением токена
type TokenGetter interface {
	GetRefreshToken(ctx context.Context, refreshToken string) (*authModel.RefreshToken, error)
}

// TokenRevoker интерфейс описывающий работу с отзыванием токена
type TokenRevoker interface {
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
}

// Authenticator интерфейс описывающий полный спектр работ по авторизации
type Authenticator interface {
	UserGetter
	TokenSaver
	TokenGetter
	TokenRevoker
}

// NewAuthStorage инициализация структуры для работы авторизации
// в структуре присутствует репозиторий для работы с репозиторием
func NewAuthStorage(rep repository.Repository) Authenticator {
	return &auth.Auth{
		Repository: &rep,
	}
}
