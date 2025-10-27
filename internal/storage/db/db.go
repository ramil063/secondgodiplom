package db

import (
	"context"
	"time"

	"github.com/ramil063/secondgodiplom/internal/logger"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

// Storage структура для работы с данными
type Storage struct {
	Repository *repository.Repository
}

func (s *Storage) SetRepository(repository *repository.Repository) {
	s.Repository = repository
}

func (s *Storage) GetRepository() repository.Repository {
	return *s.Repository
}

// Init инициализация таблиц бд и проверка соединения
func Init(repository repository.Repository) error {
	var err error

	if err = CheckPing(repository); err != nil {
		logger.WriteErrorLog(err.Error())
		return err
	}

	err = CreateTables(repository)
	return err
}

// CheckPing проверка соединения с бд
func CheckPing(repository repository.Repository) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return repository.PingContext(ctx)
}

// CreateTables создание основных таблиц
func CreateTables(repository repository.Repository) error {
	var err error

	createTablesSQL := `--USERS
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		login VARCHAR(64) UNIQUE NOT NULL,
		password_hash BYTEA NOT NULL,
		first_name VARCHAR(64),
		last_name VARCHAR(64),
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);
	COMMENT ON COLUMN public.users.id IS 'Идентификатор пользователя';
	COMMENT ON COLUMN public.users.login IS 'Логин пользователя';
	COMMENT ON COLUMN public.users.password_hash IS 'Хеш пароля (с солью)';
	COMMENT ON COLUMN public.users.first_name IS 'Имя пользователя';
	COMMENT ON COLUMN public.users.last_name IS 'Фамилия пользователя';
	COMMENT ON COLUMN public.users.is_active IS 'Флаг деактивации';
	COMMENT ON COLUMN public.users.created_at IS 'Дата создания';
	COMMENT ON COLUMN public.users.updated_at IS 'Дата обновления';

			--ITEM_TYPE
	CREATE TABLE IF NOT EXISTS item_type (
		id SERIAL PRIMARY KEY,
		alias VARCHAR(32) UNIQUE NOT NULL,
		name VARCHAR(64) NOT NULL
	);
	COMMENT ON COLUMN public.item_type.id IS 'Идентификатор типа';
	COMMENT ON COLUMN public.item_type.alias IS 'Псевдоним';
	COMMENT ON COLUMN public.item_type.name IS 'Название';

	-- Пароли (логин/пароль)
	INSERT INTO item_type (alias, name) 
	VALUES ('passwords', 'Пары логин/пароль')
	ON CONFLICT (alias) DO NOTHING;
	
	-- Произвольные текстовые данные
	INSERT INTO item_type (alias, name) 
	VALUES ('text', 'Произвольные текстовые данные')
	ON CONFLICT (alias) DO NOTHING;
	
	-- Произвольные бинарные данные
	INSERT INTO item_type (alias, name) 
	VALUES ('binary', 'Произвольные бинарные данные')
	ON CONFLICT (alias) DO NOTHING;
	
	-- Данные банковских карт
	INSERT INTO item_type (alias, name) 
	VALUES ('card', 'Данные банковских карт')
	ON CONFLICT (alias) DO NOTHING;

	        --ENCRYPTED_ITEM
	CREATE TABLE IF NOT EXISTS encrypted_item (
		id SERIAL PRIMARY KEY,
		encrypted_data BYTEA NOT NULL,
		description TEXT,
		is_deleted BOOLEAN DEFAULT FALSE,
		user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		item_type_id INT NOT NULL REFERENCES item_type(id),
		encryption_algorithm VARCHAR(32),
		iv BYTEA,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);
	COMMENT ON COLUMN public.encrypted_item.id IS 'Идентификатор записи';
	COMMENT ON COLUMN public.encrypted_item.encrypted_data IS 'Зашифрованные данные';
	COMMENT ON COLUMN public.encrypted_item.description IS 'Описание';
	COMMENT ON COLUMN public.encrypted_item.is_deleted IS 'Мягкое удаление';
	COMMENT ON COLUMN public.encrypted_item.user_id IS 'Пользователь';
	COMMENT ON COLUMN public.encrypted_item.item_type_id IS 'Тип';
	COMMENT ON COLUMN public.encrypted_item.encryption_algorithm IS 'Алгоритм шифрования';
	COMMENT ON COLUMN public.encrypted_item.iv IS 'Вектор инициализации';
	COMMENT ON COLUMN public.encrypted_item.created_at IS 'Дата создания';
	COMMENT ON COLUMN public.encrypted_item.updated_at IS 'Дата обновления';

	        --ITEM_METADATA
	CREATE TABLE IF NOT EXISTS item_metadata (
		id SERIAL PRIMARY KEY,
		item_id INT NOT NULL REFERENCES encrypted_item(id) ON DELETE CASCADE,
		name VARCHAR(128) NOT NULL,
		value TEXT,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);
	COMMENT ON COLUMN public.item_metadata.id IS 'Идентификатор записи';
	COMMENT ON COLUMN public.item_metadata.item_id IS 'Связь с данными';
	COMMENT ON COLUMN public.item_metadata.name IS 'Название метаданных';
	COMMENT ON COLUMN public.item_metadata.value IS 'Значение метаданных';
	COMMENT ON COLUMN public.item_metadata.created_at IS 'Дата создания';
	COMMENT ON COLUMN public.item_metadata.updated_at IS 'Дата обновления';

			--OAUTH_ACCESS_TOKEN
	CREATE TABLE IF NOT EXISTS oauth_access_token (
		id SERIAL PRIMARY KEY,
		token_hash BYTEA NOT NULL,
		user_id INT REFERENCES users(id),
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);
	COMMENT ON COLUMN public.oauth_access_token.id IS 'Идентификатор токена';
	COMMENT ON COLUMN public.oauth_access_token.token_hash IS 'Хеш токена';
	COMMENT ON COLUMN public.oauth_access_token.user_id IS 'Владелец кода';
	COMMENT ON COLUMN public.oauth_access_token.expires_at IS 'Срок действия';
	COMMENT ON COLUMN public.oauth_access_token.created_at IS 'Дата создания';

			--OAUTH_REFRESH_TOKEN
	CREATE TABLE IF NOT EXISTS oauth_refresh_token (
		id SERIAL PRIMARY KEY,
		token_hash BYTEA NOT NULL,          -- хеш refresh-токена
		access_token_id INT NOT NULL REFERENCES oauth_access_token(id) ON DELETE CASCADE,
		is_revoked BOOLEAN DEFAULT FALSE,
		expires_at TIMESTAMP NOT NULL,    -- срок действия
		created_at TIMESTAMP DEFAULT NOW()
	);
	COMMENT ON COLUMN public.oauth_refresh_token.id IS 'Идентификатор токена';
	COMMENT ON COLUMN public.oauth_refresh_token.token_hash IS 'Хеш токена';
	COMMENT ON COLUMN public.oauth_refresh_token.access_token_id IS 'Публичный идентификатор';
	COMMENT ON COLUMN public.oauth_refresh_token.is_revoked IS 'отозван ли токен';
	COMMENT ON COLUMN public.oauth_refresh_token.expires_at IS 'Срок действия';
	COMMENT ON COLUMN public.oauth_refresh_token.created_at IS 'Дата создания';

	-- BINARY_FILES
	CREATE TABLE IF NOT EXISTS binary_file (
		id SERIAL PRIMARY KEY,
		filename VARCHAR(255) NOT NULL,
		mime_type VARCHAR(100),
		original_size BIGINT NOT NULL,
		chunk_size INTEGER NOT NULL,
		total_chunks INTEGER NOT NULL,
		description TEXT,
		is_deleted BOOLEAN DEFAULT FALSE,
		is_complete BOOLEAN DEFAULT FALSE,
		user_id INTEGER NOT NULL REFERENCES users(id),
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);
	COMMENT ON COLUMN public.binary_file.id IS 'Идентификатор токена';
	COMMENT ON COLUMN public.binary_file.filename IS 'Название';
	COMMENT ON COLUMN public.binary_file.mime_type IS 'Тип';
	COMMENT ON COLUMN public.binary_file.original_size IS 'Размер';
	COMMENT ON COLUMN public.binary_file.chunk_size IS 'Размер части';
	COMMENT ON COLUMN public.binary_file.total_chunks IS 'Количество частей';
	COMMENT ON COLUMN public.binary_file.description IS 'Описание';
	COMMENT ON COLUMN public.binary_file.is_deleted IS 'Мягкое удаление';
	COMMENT ON COLUMN public.binary_file.is_complete IS 'Закачанный файл?';
	COMMENT ON COLUMN public.binary_file.user_id IS 'Пользователь';
	COMMENT ON COLUMN public.binary_file.created_at IS 'Дата создания';
	COMMENT ON COLUMN public.binary_file.updated_at IS 'Дата обновления';
	
	-- BINARY_CHUNKS
	CREATE TABLE IF NOT EXISTS binary_file_chunk (
		id SERIAL PRIMARY KEY,
		file_id INTEGER NOT NULL REFERENCES binary_file(id) ON DELETE CASCADE,
		chunk_index INTEGER NOT NULL,
		encrypted_data BYTEA NOT NULL,
		encryption_algorithm VARCHAR(32) NOT NULL,
		iv BYTEA NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		UNIQUE(file_id, chunk_index)
	);
	COMMENT ON COLUMN public.binary_file_chunk.id IS 'Идентификатор токена';
	COMMENT ON COLUMN public.binary_file_chunk.file_id IS 'Хеш токена';
	COMMENT ON COLUMN public.binary_file_chunk.chunk_index IS 'Публичный идентификатор';
	COMMENT ON COLUMN public.binary_file_chunk.encrypted_data IS 'Зашифрованные данные';
	COMMENT ON COLUMN public.binary_file_chunk.encryption_algorithm IS 'Алгоритм шифрования';
	COMMENT ON COLUMN public.binary_file_chunk.iv IS 'Вектор инициализации';
	COMMENT ON COLUMN public.binary_file_chunk.created_at IS 'Дата создания';

	        --ITEM_METADATA
	CREATE TABLE IF NOT EXISTS binary_file_metadata (
		id SERIAL PRIMARY KEY,
		file_id INT NOT NULL REFERENCES binary_file(id) ON DELETE CASCADE,
		name VARCHAR(128) NOT NULL,
		value TEXT,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);
	COMMENT ON COLUMN public.binary_file_metadata.id IS 'Идентификатор записи';
	COMMENT ON COLUMN public.binary_file_metadata.file_id IS 'Связь с данными';
	COMMENT ON COLUMN public.binary_file_metadata.name IS 'Название метаданных';
	COMMENT ON COLUMN public.binary_file_metadata.value IS 'Значение метаданных';
	COMMENT ON COLUMN public.binary_file_metadata.created_at IS 'Дата создания';
	COMMENT ON COLUMN public.binary_file_metadata.updated_at IS 'Дата обновления';
`
	_, err = repository.ExecContext(context.Background(), createTablesSQL)
	return err
}
