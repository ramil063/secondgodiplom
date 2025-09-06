package binary

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ramil063/secondgodiplom/internal/constants/queue"
)

func SaveToDeleteQueue(id int64) (string, error) {
	request := Request{
		ID:         id,
		RetryCount: 0,
		Status:     queue.RequestStatusPending,
	}

	// Сохраняем в JSON файл
	filename := fmt.Sprintf("%s/delete_%s.json", queue.DirBinary, request.ID)
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
