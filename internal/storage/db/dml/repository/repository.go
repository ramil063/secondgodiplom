package repository

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	serverConfig "github.com/ramil063/secondgodiplom/cmd/gophkeeper/config"
	internalErrors "github.com/ramil063/secondgodiplom/internal/errors"
	"github.com/ramil063/secondgodiplom/internal/logger"
)

type Pooler interface {
	Close()
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Ping(ctx context.Context) error
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type Repository struct {
	Pool Pooler
}

func (dbr *Repository) ExecContext(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	result, err := dbr.Pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Open открыть соединение с бд
func (dbr *Repository) Open(config *serverConfig.ServerConfig) (*pgxpool.Pool, error) {
	conf, err := pgxpool.ParseConfig(config.DatabaseURI)
	if err != nil {
		return nil, internalErrors.NewDBError(err)
	}

	conf.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// do something with every new connection
		return nil
	}
	conf.MaxConns = config.DbMaxConnections
	conf.MinConns = config.DbMinConnections

	pool, err := pgxpool.ConnectConfig(context.Background(), conf)

	if err != nil {
		return nil, internalErrors.NewDBError(err)
	}
	return pool, nil
}

// PingContext проверить соединение с бд
func (dbr *Repository) PingContext(ctx context.Context) error {
	err := dbr.Pool.Ping(ctx)
	if err != nil {
		logger.WriteErrorLog(err.Error())
	}
	return err
}

// SetPool установить поле Pool для работы с бд
func (dbr *Repository) SetPool(config *serverConfig.ServerConfig) error {
	pool, err := dbr.Open(config)
	if err != nil {
		logger.WriteErrorLog(err.Error())
		return err
	}
	dbr.Pool = pool
	return nil
}

func NewRepository(config *serverConfig.ServerConfig) (*Repository, error) {
	rep := &Repository{}
	err := rep.SetPool(config)
	return rep, err
}
