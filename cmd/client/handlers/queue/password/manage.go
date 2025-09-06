package password

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/ramil063/secondgodiplom/internal/constants/queue"
)

func SaveToCreateQueue(login, password, target, description, metaDataName, metaDataValue string) (string, error) {
	request := Request{
		ID:            uuid.New().String(),
		Login:         login,
		Password:      password,
		Target:        target,
		Description:   description,
		MetaDataName:  metaDataName,
		MetaDataValue: metaDataValue,
		CreatedAt:     time.Now(),
		RetryCount:    0,
		Status:        queue.RequestStatusPending,
	}

	// Сохраняем в JSON файл
	filename := fmt.Sprintf("%s/create_%s.json", queue.DirPassword, request.ID)
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

func SaveToUpdateQueue(id int64, login, password, target, description string) (string, error) {
	request := Request{
		ID:          strconv.Itoa(int(id)),
		Login:       login,
		Password:    password,
		Target:      target,
		Description: description,
		CreatedAt:   time.Now(),
		RetryCount:  0,
		Status:      queue.RequestStatusPending,
	}

	// Сохраняем в JSON файл
	filename := fmt.Sprintf("%s/update_%s.json", queue.DirPassword, request.ID)
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

func SaveToDeleteQueue(id int64) (string, error) {
	request := Request{
		ID:         strconv.Itoa(int(id)),
		RetryCount: 0,
		Status:     queue.RequestStatusPending,
	}

	// Сохраняем в JSON файл
	filename := fmt.Sprintf("%s/delete_%s.json", queue.DirPassword, request.ID)
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
