package storage

import (
	"context"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/user"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/registration"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

// Registerer интерфейс описывающий логику регистрации пользователя
type Registerer interface {
	RegisterUser(ctx context.Context, user *user.User) (int, error)
}

// NewRegistrationStorage инициализация структуры для регистрации пользователя
// в структуре есть указатель на репозиторий
func NewRegistrationStorage(rep repository.Repository) Registerer {
	return &registration.Reg{
		Repository: &rep,
	}
}
