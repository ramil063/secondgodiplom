package binary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/items"
	"github.com/ramil063/secondgodiplom/internal/logger"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

type Item struct {
	Repository *repository.Repository
}

func (i *Item) CreateFileRecord(
	ctx context.Context,
	userID int,
	metadata *binarydata.FileMetadata,
) (int64, error) {
	var fileID int64

	err := i.Repository.Pool.QueryRow(ctx, `
        INSERT INTO binary_file (
            user_id, filename, mime_type, original_size, 
            description, chunk_size, total_chunks
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id`,
		userID,
		metadata.Filename,
		metadata.MimeType,
		metadata.OriginalSize,
		metadata.Description,
		metadata.ChunkSize,
		metadata.TotalChunks,
	).Scan(&fileID)

	return fileID, err
}

func (i *Item) SaveChunk(
	ctx context.Context,
	fileID int64,
	chunkIndex int32,
	encryptedData []byte,
	algorithm string,
	iv []byte,
) error {
	result, err := i.Repository.Pool.Exec(ctx, `
        INSERT INTO binary_file_chunk (file_id, chunk_index, encrypted_data, encryption_algorithm, iv)
        VALUES ($1, $2, $3, $4, $5)`,
		fileID, chunkIndex, encryptedData, algorithm, iv,
	)

	if err != nil {
		return fmt.Errorf("failed to mark file as complete: %w", err)
	}

	// Проверяем, что файл был найден и обновлен
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("file not found or already deleted: ID %s", fileID)
	}
	return err
}

func (i *Item) MarkFileComplete(ctx context.Context, fileID int64, totalBytes int64) error {

	// Выполняем запрос на обновление
	result, err := i.Repository.Pool.Exec(ctx, `
        UPDATE binary_file 
        SET 
            is_complete = TRUE,
            original_size = $1,
            updated_at = NOW()
        WHERE id = $2 
        AND is_deleted = FALSE`,
		totalBytes, // Обновляем actual size
		fileID,
	)

	if err != nil {
		return fmt.Errorf("failed to mark file as complete: %w", err)
	}

	// Проверяем, что файл был найден и обновлен
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("file not found or already deleted: ID %s", fileID)
	}

	return nil
}

func (i *Item) GetFileInfo(ctx context.Context, fileID int64, userID int64) (*items.FileInfo, error) {
	query := `SELECT
				bf.id,
				bf.filename,
				bf.mime_type,
				bf.original_size,
				bf.description,
				bf.chunk_size,
				bf.total_chunks,
				bf.created_at,
				COALESCE(
					json_agg(
						json_build_object(
							'id', bfm.id,
							'name', bfm.name,
							'value', bfm.value,
							'created_at', bfm.created_at
						) 
						ORDER BY bfm.created_at
					) FILTER (WHERE bfm.id IS NOT NULL),
					'[]'
				) as metadata
			FROM binary_file bf
			LEFT JOIN binary_file_metadata bfm on bf.id = bfm.file_id
			WHERE bf.is_deleted = FALSE 
			  AND bf.id = $1 
			  AND bf.user_id = $2
			GROUP BY bf.id
			`

	row := i.Repository.Pool.QueryRow(ctx, query, fileID, userID)
	var fileInfo items.FileInfo
	var metadataJSON []byte

	err := row.Scan(
		&fileInfo.ID,
		&fileInfo.Filename,
		&fileInfo.MimeType,
		&fileInfo.OriginalSize,
		&fileInfo.Description,
		&fileInfo.ChunkSize,
		&fileInfo.TotalChunks,
		&fileInfo.CreatedAt,
		&metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to count items: %w", err)
	}
	// Парсим JSON с метаданными
	if err = json.Unmarshal(metadataJSON, &fileInfo.MetaDataItems); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	return &fileInfo, nil
}

func (r *Item) GetChunksInRange(ctx context.Context, fileID int64, start, end int32) ([]*items.ChunkData, error) {
	var chunks []*items.ChunkData

	rows, err := r.Repository.Pool.Query(ctx, `
        SELECT chunk_index, encrypted_data, encryption_algorithm, iv 
        FROM binary_file_chunk
        WHERE file_id = $1 AND chunk_index BETWEEN $2 AND $3
        ORDER BY chunk_index`,
		fileID, start, end-1, // end исключается
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var chunk items.ChunkData
		err = rows.Scan(&chunk.ChunkIndex, &chunk.EncryptedData, &chunk.EncryptionAlgorithm, &chunk.IV)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, &chunk)
	}

	return chunks, nil
}

func (i *Item) DeleteFile(ctx context.Context, userID, fileID int64) error {
	exec, err := i.Repository.Pool.Exec(
		ctx,
		`UPDATE binary_file 
				SET is_deleted=TRUE 
				WHERE id = $1 AND user_id = $2`,
		fileID,
		userID)

	if err != nil {
		return errors.New("DeleteFile error in sql empty result")
	}
	if exec == nil {
		logger.WriteErrorLog("DeleteFile error in sql empty result")
		return errors.New("DeleteFile error in sql empty result")
	}

	rows := exec.RowsAffected()
	if rows != 1 {
		logger.WriteErrorLog("DeleteFile error expected to affect 1 row")
		return errors.New("DeleteFile expected to affect 1 row")
	}
	return nil
}

func (i *Item) GetListFiles(ctx context.Context, userID int64, page int32, perPage int32, filter string) ([]*items.FileInfo, int32, error) {
	offset := (page - 1) * perPage
	args := []interface{}{userID}
	paramCounter := 2

	mainSelectFields := `bf.id,
            bf.filename,
            bf.mime_type,
			bf.original_size,
            bf.description,
            bf.created_at`

	countSelectFields := `COUNT(*)`

	// Базовый запрос
	commonQuery := `
		SELECT 
        {{select}}
        FROM binary_file bf
        LEFT JOIN binary_file_metadata bfm ON bf.id = bfm.file_id
        WHERE bf.user_id = $1
        AND bf.is_deleted = FALSE
        `

	// Добавляем фильтр
	if filter != "" {
		commonQuery += fmt.Sprintf(` AND (
            bf.description ILIKE $%d
            OR bf.filename ILIKE $%d
            OR bf.mime_type ILIKE $%d
            OR bfm.name ILIKE $%d
            OR bfm.value ILIKE $%d
        )`, paramCounter, paramCounter, paramCounter, paramCounter, paramCounter)
		args = append(args, "%"+filter+"%")
		paramCounter++
	}

	countCommonQuery := strings.Replace(commonQuery, "{{select}}", countSelectFields, 1)

	// Группируем и добавляем пагинацию
	limitOffset := fmt.Sprintf(`
		GROUP BY bf.id
        ORDER BY bf.created_at DESC
		LIMIT $%d OFFSET $%d`, paramCounter, paramCounter+1)

	args = append(args, perPage, offset)

	commonQuery = strings.Replace(commonQuery, "{{select}}", mainSelectFields, 1)
	commonQuery += limitOffset

	// Выполняем запрос
	rows, err := i.Repository.Pool.Query(ctx, commonQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var filesInfo []*items.FileInfo
	for rows.Next() {
		var fi items.FileInfo

		err := rows.Scan(
			&fi.ID,
			&fi.Filename,
			&fi.MimeType,
			&fi.OriginalSize,
			&fi.Description,
			&fi.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan items: %w", err)
		}

		filesInfo = append(filesInfo, &fi)
	}

	// Получаем общее количество
	totalCount, err := i.GetTotalCount(ctx, countCommonQuery, userID, filter)
	if err != nil {
		logger.WriteErrorLog("GetTotalCount error: " + err.Error())
		totalCount = 0
	}

	return filesInfo, totalCount, nil
}

func (i *Item) GetTotalCount(ctx context.Context, query string, userID int64, filter string) (int32, error) {
	// Получаем общее количество
	var totalCount int32
	filterCondition := ""

	args := []interface{}{userID} // $1 = userID
	if filter != "" {
		args = append(args, "%"+filter+"%") // $2 = filter
	}
	query = strings.Replace(query, "{{filterConditions}}", filterCondition, 1)

	err := i.Repository.Pool.QueryRow(ctx, query, args...).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to count items: %w", err)
	}
	return totalCount, nil
}
