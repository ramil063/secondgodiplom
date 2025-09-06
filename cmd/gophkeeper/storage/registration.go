package storage

import (
	"context"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/user"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/registration"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

type Registerer interface {
	RegisterUser(ctx context.Context, user *user.User) (int, error)
}

func NewRegistrationStorage(rep repository.Repository) Registerer {
	return &registration.Reg{
		Repository: &rep,
	}
}
