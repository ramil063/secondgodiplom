package auth

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/user"
)

func (s *Auth) GetUserByLogin(ctx context.Context, login string) (*user.User, error) {
	row := s.Repository.Pool.QueryRow(
		ctx,
		"SELECT id, password_hash FROM users WHERE login = $1",
		login)

	var id int
	var passwordHash []byte

	err := row.Scan(&id, &passwordHash)
	u := &user.User{
		ID:           id,
		PasswordHash: string(passwordHash),
	}

	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}
	return u, nil
}
