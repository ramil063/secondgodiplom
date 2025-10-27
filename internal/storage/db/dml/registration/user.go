package registration

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/user"
	internalErrors "github.com/ramil063/secondgodiplom/internal/errors"
	"github.com/ramil063/secondgodiplom/internal/logger"
)

func (s *Reg) RegisterUser(ctx context.Context, user *user.User) (int, error) {
	var userID int

	err := s.Repository.Pool.QueryRow(
		ctx,
		`INSERT INTO users (login, password_hash, first_name, last_name, is_active) 
         VALUES ($1, $2, $3, $4, $5) 
         RETURNING id`,
		user.Login,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		true,
	).Scan(&userID) // Сканируем возвращённый ID

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, internalErrors.ErrUniqueViolation
		}
		logger.WriteErrorLog("RegisterUser error" + err.Error())
		return 0, fmt.Errorf("failed to register user: %w", err)
	}

	return userID, nil
}
