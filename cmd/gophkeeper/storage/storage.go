package storage

import (
	"github.com/ramil063/secondgodiplom/internal/storage/db"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

// Repositorier интерфейс описывающий функции для получения и возврата репозитория
type Repositorier interface {
	SetRepository(repository *repository.Repository)
	GetRepository() repository.Repository
}

// Storager интерфейс центрального хранилища
type Storager interface {
	Repositorier
}

// NewDBStorage инициализация центрального хранилища
func NewDBStorage() Storager {
	return &db.Storage{}
}
