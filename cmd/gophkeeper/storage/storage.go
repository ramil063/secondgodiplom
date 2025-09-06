package storage

import (
	"github.com/ramil063/secondgodiplom/internal/storage/db"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

type Repositorier interface {
	SetRepository(repository *repository.Repository)
	GetRepository() repository.Repository
}

type Storager interface {
	Repositorier
}

func NewDBStorage() Storager {
	return &db.Storage{}
}
