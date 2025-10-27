package textdata

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/ramil063/secondgodiplom/internal/constants/queue"
)

// SaveToCreateQueue создание файла с текстом для отправки на сервер и создания новой записи
func SaveToCreateQueue(textData, description, metaDataName, metaDataValue string) (string, error) {
	request := Request{
		ID:            uuid.New().String(),
		TextData:      textData,
		Description:   description,
		MetaDataName:  metaDataName,
		MetaDataValue: metaDataValue,
		CreatedAt:     time.Now(),
		RetryCount:    0,
		Status:        queue.RequestStatusPending,
	}

	// Сохраняем в JSON файл
	filename := fmt.Sprintf("%s/create_%s.json", queue.DirTextData, request.ID)
	data, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return "", err
	}

	return request.ID, nil
}

// SaveToUpdateQueue создание файла с текстом для отправки на сервер и обновления
func SaveToUpdateQueue(id int64, textData, description string) (string, error) {
	request := Request{
		ID:          strconv.Itoa(int(id)),
		TextData:    textData,
		Description: description,
		CreatedAt:   time.Now(),
		RetryCount:  0,
		Status:      queue.RequestStatusPending,
	}

	// Сохраняем в JSON файл
	filename := fmt.Sprintf("%s/update_%s.json", queue.DirTextData, request.ID)
	data, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return "", err
	}

	return request.ID, nil
}

// SaveToDeleteQueue создание файла с текстом для отправки на сервер для удаления
func SaveToDeleteQueue(id int64) (string, error) {
	request := Request{
		ID:         strconv.Itoa(int(id)),
		RetryCount: 0,
		Status:     queue.RequestStatusPending,
	}

	// Сохраняем в JSON файл
	filename := fmt.Sprintf("%s/delete_%s.json", queue.DirTextData, request.ID)
	data, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return "", err
	}

	return request.ID, nil
}
