package storage

import (
	"context"

	auth2 "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/auth"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/user"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/auth"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

type UserGetter interface {
	GetUserByLogin(ctx context.Context, login string) (*user.User, error)
}

type Loginer interface {
	LoginUser(ctx context.Context) (string, string, error)
}

type TokenSaver interface {
	SaveAccessToken(ctx context.Context, userID int, token string) (int, error)
	SaveRefreshToken(ctx context.Context, accessTokenId int, token string) error
}

type TokenGetter interface {
	GetRefreshToken(ctx context.Context, refreshToken string) (*auth2.RefreshToken, error)
}

type TokenRevoker interface {
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
}

type Authenticator interface {
	UserGetter
	TokenSaver
	TokenGetter
	TokenRevoker
}

func NewAuthStorage(rep repository.Repository) Authenticator {
	return &auth.Auth{
		Repository: &rep,
	}
}
