package bankcard

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ramil063/secondgodiplom/internal/constants/queue"
)

func SaveToCreateQueue(
	number string,
	validUntilYear int32,
	validUntilMonth int32,
	cvv int32,
	holder, description, metaDataName, metaDataValue string,
) (string, error) {
	request := Request{
		GeneratedID:     uuid.New().String(),
		Number:          number,
		ValidUntilYear:  validUntilYear,
		ValidUntilMonth: validUntilMonth,
		Cvv:             cvv,
		Holder:          holder,
		Description:     description,
		MetaDataName:    metaDataName,
		MetaDataValue:   metaDataValue,
		CreatedAt:       time.Now(),
		RetryCount:      0,
		Status:          queue.RequestStatusPending,
	}

	// Сохраняем в JSON файл
	filename := fmt.Sprintf("%s/create_%d.json", queue.DirBankcard, request.ID)
	data, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return "", err
	}

	return request.GeneratedID, nil
}

func SaveToUpdateQueue(
	id int64,
	number string,
	validUntilYear int32,
	validUntilMonth int32,
	cvv int32,
	holder,
	description string,
) (string, error) {
	request := Request{
		GeneratedID:     uuid.New().String(),
		ID:              id,
		Number:          number,
		ValidUntilYear:  validUntilYear,
		ValidUntilMonth: validUntilMonth,
		Cvv:             cvv,
		Holder:          holder,
		Description:     description,
		CreatedAt:       time.Now(),
		RetryCount:      0,
		Status:          queue.RequestStatusPending,
	}

	// Сохраняем в JSON файл
	filename := fmt.Sprintf("%s/update_%s.json", queue.DirBankcard, request.GeneratedID)
	data, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return "", err
	}

	return request.GeneratedID, nil
}

func SaveToDeleteQueue(id int64) (string, error) {
	request := Request{
		GeneratedID: uuid.New().String(),
		ID:          id,
		RetryCount:  0,
		Status:      queue.RequestStatusPending,
	}

	// Сохраняем в JSON файл
	filename := fmt.Sprintf("%s/delete_%s.json", queue.DirBankcard, request.GeneratedID)
	data, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return "", err
	}

	return request.GeneratedID, nil
}
