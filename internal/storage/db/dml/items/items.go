package items

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	itemModel "github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/items"
	"github.com/ramil063/secondgodiplom/internal/logger"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

type Item struct {
	Repository *repository.Repository
}

func (pi *Item) SaveEncryptedData(
	ctx context.Context,
	encryptedItem *itemModel.EncryptedItem,
) (int64, error) {

	typeRow := pi.Repository.Pool.QueryRow(
		ctx,
		`SELECT id FROM item_type WHERE alias = $1`,
		encryptedItem.Type)

	var typeId int64
	if err := typeRow.Scan(&typeId); err != nil {
		return 0, err
	}

	row := pi.Repository.Pool.QueryRow(
		ctx,
		`INSERT INTO encrypted_item (encrypted_data, description, user_id, item_type_id, encryption_algorithm, iv)
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		encryptedItem.Data,
		encryptedItem.Description,
		encryptedItem.UserID,
		typeId,
		encryptedItem.EncryptionAlgorithm,
		encryptedItem.Iv)

	var itemId int64
	err := row.Scan(&itemId)

	if err != nil {
		return 0, errors.New("SaveAccessToken error in sql empty result")
	}

	return itemId, nil
}

func (pi *Item) SaveMetadata(ctx context.Context, metadata *itemModel.MetaData) error {
	exec, err := pi.Repository.Pool.Exec(
		ctx,
		`INSERT INTO item_metadata (item_id, name, value) VALUES ($1, $2, $3)`,
		metadata.ItemID,
		metadata.Name,
		metadata.Value)

	if err != nil {
		return err
	}
	if exec == nil {
		logger.WriteErrorLog("save metadata error in sql empty result")
		return errors.New("error in sql empty result")
	}

	rows := exec.RowsAffected()
	if rows != 1 {
		logger.WriteErrorLog("save metadata error expected to affect 1 row")
		return errors.New("expected to affect 1 row")
	}
	return nil
}

func (pi *Item) GetListItems(
	ctx context.Context,
	userID int64,
	page int32,
	perPage int32,
	itemType string,
	filter string,
) ([]*itemModel.ItemData, int32, error) {
	offset := (page - 1) * perPage
	args := []interface{}{userID}
	args = append(args, itemType)
	paramCounter := 3

	mainSelectFields := `ei.id,
            ei.encrypted_data,
            ei.description,
            ei.created_at,
            ei.encryption_algorithm,
            ei.iv,
            COALESCE(
                json_agg(
                    json_build_object(
                        'id', im.id,
                        'name', im.name,
                        'value', im.value,
                        'created_at', im.created_at
                    ) 
                    ORDER BY im.created_at
                ) FILTER (WHERE im.id IS NOT NULL),
                '[]'
            ) as metadata`

	countSelectFields := `COUNT(*)`

	// Базовый запрос
	commonQuery := `
		SELECT 
        {{select}}
        FROM encrypted_item ei
        LEFT JOIN item_metadata im ON ei.id = im.item_id
        LEFT JOIN item_type it ON ei.item_type_id = it.id
        WHERE ei.user_id = $1
        AND ei.is_deleted = FALSE
        AND it.alias = $2`

	// Добавляем фильтр
	if filter != "" {
		commonQuery += fmt.Sprintf(` AND (
            ei.description ILIKE $%d
            OR im.name ILIKE $%d
            OR im.value ILIKE $%d
        )`, paramCounter, paramCounter, paramCounter)
		args = append(args, "%"+filter+"%")
		paramCounter++
	}

	countCommonQuery := strings.Replace(commonQuery, "{{select}}", countSelectFields, 1)

	// Группируем и добавляем пагинацию
	limitOffset := fmt.Sprintf(`
		GROUP BY ei.id
        ORDER BY ei.created_at DESC
		LIMIT $%d OFFSET $%d`, paramCounter, paramCounter+1)

	args = append(args, perPage, offset)

	commonQuery = strings.Replace(commonQuery, "{{select}}", mainSelectFields, 1)
	commonQuery += limitOffset

	// Выполняем запрос
	rows, err := pi.Repository.Pool.Query(ctx, commonQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*itemModel.ItemData
	for rows.Next() {
		var pwd itemModel.ItemData
		var metadataJSON []byte

		err = rows.Scan(
			&pwd.ID,
			&pwd.Data,
			&pwd.Description,
			&pwd.CreatedAt,
			&pwd.EncryptionAlgorithm,
			&pwd.IV,
			&metadataJSON,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan items: %w", err)
		}

		// Парсим JSON с метаданными
		if err = json.Unmarshal(metadataJSON, &pwd.MetaDataItems); err != nil {
			return nil, 0, fmt.Errorf("failed to parse metadata: %w", err)
		}

		items = append(items, &pwd)
	}

	// Получаем общее количество
	totalCount, err := pi.GetTotalCount(ctx, countCommonQuery, userID, itemType, filter)
	if err != nil {
		logger.WriteErrorLog("GetTotalCount error: " + err.Error())
		totalCount = 0
	}

	return items, totalCount, nil
}

func (pi *Item) GetTotalCount(ctx context.Context, query string, userID int64, itemType, filter string) (int32, error) {
	// Получаем общее количество
	var totalCount int32
	filterCondition := ""

	args := []interface{}{userID} // $1 = userID
	args = append(args, itemType)
	if filter != "" {
		args = append(args, "%"+filter+"%") // $2 = perPage
	}
	query = strings.Replace(query, "{{filterConditions}}", filterCondition, 1)

	err := pi.Repository.Pool.QueryRow(ctx, query, args...).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to count items: %w", err)
	}
	return totalCount, nil
}

func (pi *Item) GetItem(ctx context.Context, itemID int64) (*itemModel.ItemData, error) {

	query := `SELECT
				ei.id,
				ei.encrypted_data,
				ei.description,
				ei.created_at,
				ei.encryption_algorithm,
				ei.iv,
				COALESCE(
					json_agg(
						json_build_object(
							'id', im.id,
							'name', im.name,
							'value', im.value,
							'created_at', im.created_at
						) 
						ORDER BY im.created_at
					) FILTER (WHERE im.id IS NOT NULL),
					'[]'
				) as metadata
			FROM encrypted_item ei
			LEFT JOIN public.item_metadata im on ei.id = im.item_id
			WHERE ei.is_deleted = FALSE AND ei.id = $1
			GROUP BY ei.id
			`

	row := pi.Repository.Pool.QueryRow(ctx, query, itemID)
	var pwd itemModel.ItemData
	var metadataJSON []byte

	err := row.Scan(
		&pwd.ID,
		&pwd.Data,
		&pwd.Description,
		&pwd.CreatedAt,
		&pwd.EncryptionAlgorithm,
		&pwd.IV,
		&metadataJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to count items: %w", err)
	}
	// Парсим JSON с метаданными
	if err := json.Unmarshal(metadataJSON, &pwd.MetaDataItems); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	return &itemModel.ItemData{
		ID:                  pwd.ID,
		Data:                pwd.Data,
		Description:         pwd.Description,
		CreatedAt:           pwd.CreatedAt,
		EncryptionAlgorithm: pwd.EncryptionAlgorithm,
		IV:                  pwd.IV,
		MetaDataItems:       pwd.MetaDataItems,
	}, nil
}

func (pi *Item) DeleteItem(ctx context.Context, itemID int64) error {
	exec, err := pi.Repository.Pool.Exec(
		ctx,
		`UPDATE encrypted_item SET is_deleted=TRUE WHERE id = $1`,
		itemID)

	if err != nil {
		return errors.New("DeleteItem error in sql empty result")
	}
	if exec == nil {
		logger.WriteErrorLog("DeleteItem error in sql empty result")
		return errors.New("DeleteItem error in sql empty result")
	}

	rows := exec.RowsAffected()
	if rows != 1 {
		logger.WriteErrorLog("DeleteItem error expected to affect 1 row")
		return errors.New("DeleteItem expected to affect 1 row")
	}
	return nil
}

func (pi *Item) UpdateItem(
	ctx context.Context,
	itemId int64,
	encryptedItem *itemModel.EncryptedItem,
) (int64, error) {

	setQuery := ``
	args := make([]interface{}, 0)
	args = append(args, itemId)

	if len(encryptedItem.Data) > 0 {
		num := len(args) + 1
		setQuery += `encrypted_data=$` + strconv.Itoa(num) + `,`
		args = append(args, encryptedItem.Data)
	}
	if encryptedItem.Description != "" {
		num := len(args) + 1
		setQuery += `description=$` + strconv.Itoa(num) + `,`
		args = append(args, encryptedItem.Description)
	}
	if encryptedItem.EncryptionAlgorithm != "" {
		num := len(args) + 1
		setQuery += `encryption_algorithm=$` + strconv.Itoa(num) + `,`
		args = append(args, encryptedItem.EncryptionAlgorithm)
	}
	if len(encryptedItem.Iv) > 0 {
		num := len(args) + 1
		setQuery += `iv=$` + strconv.Itoa(num) + `,`
		args = append(args, encryptedItem.Iv)
	}
	setQuery = strings.Trim(setQuery, ",")

	if setQuery == "" {
		return 0, errors.New("UpdateItem empty data in update")
	}

	row := pi.Repository.Pool.QueryRow(
		ctx,
		`UPDATE encrypted_item 
				SET `+setQuery+` 
				WHERE is_deleted = FALSE AND id = $1
		`,
		args...)

	if row == nil {
		logger.WriteErrorLog("UpdateItem error in sql empty result")
		return 0, errors.New("UpdateItem error in sql empty result")
	}

	return itemId, nil
}

// GetMetaDataList убрать нигде не используется
func (pi *Item) GetMetaDataList(ctx context.Context, itemId int64) ([]*itemModel.MetaData, error) {
	var metadataItem itemModel.MetaData
	var metadata []*itemModel.MetaData

	rowsMetaData, err := pi.Repository.Pool.Query(
		ctx,
		`SELECT
				id,
				name,
				value,
				created_at
			FROM item_metadata
			WHERE item_id=$1;`,
		itemId,
	)

	if err != nil {
		return nil, errors.New("GetMetaDataList error on sql empty result")
	}
	rowsMetaData.Close()
	for rowsMetaData.Next() {
		rowsMetaData.Scan(&metadataItem.ID, &metadataItem.Name, &metadataItem.Value, &metadataItem.CreatedAt)
		metadata = append(metadata, &metadataItem)
	}
	return metadata, nil
}
