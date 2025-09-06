package binary

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	"github.com/ramil063/secondgodiplom/internal/constants/queue"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
)

func Init() {
	// Создаем директорию для очереди
	os.MkdirAll(queue.DirBinary, 0755)
}

func Process(client binarydata.ServiceClient) {
	processDelete(client)
}

func processDelete(client binarydata.ServiceClient) {
	files, err := filepath.Glob(filepath.Join(queue.DirBinary, "delete_*.json"))
	if err != nil {
		fmt.Printf("❌ Ошибка поиска файлов очереди: %v\n", err)
		return
	}

	if len(files) == 0 {
		return
	}

	fmt.Printf("Найдено %d запросов в очереди\n", len(files))

	for _, file := range files {
		sendDelete(client, file)
	}
}

func sendDelete(client binarydata.ServiceClient, filename string) {
	// Читаем файл
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("❌ Ошибка чтения файла %s: %v\n", filename, err)
		return
	}

	var request Request
	if err := json.Unmarshal(data, &request); err != nil {
		fmt.Printf("❌ Ошибка парсинга JSON %s: %v\n", filename, err)
		return
	}

	// Помечаем как обрабатывается
	request.Status = queue.RequestStatusProcessing
	saveRequest(filename, request)

	// Отправляем на сервер
	ctx := items.CreateAuthContext()
	_, err = client.DeleteFile(ctx, &binarydata.DeleteFileRequest{
		FileId: request.ID,
	})

	if err != nil {
		fmt.Printf("❌ Ошибка отправки %s: %v\n", request.GeneratedID, err)
		request.RetryCount++
		request.Status = queue.RequestStatusFailed
		saveRequest(filename, request)

		// Удаляем после 3 попыток
		if request.RetryCount >= 3 {
			os.Remove(filename)
			fmt.Printf("Удален файл после 3 неудачных попыток: %s\n", request.GeneratedID)
		}
		return
	}

	// Успешная отправка - удаляем файл
	os.Remove(filename)
	fmt.Printf("✅ Запрос успешно отправлен и удален из хранилища\n")
}

func saveRequest(filename string, request Request) {
	data, _ := json.MarshalIndent(request, "", "  ")
	os.WriteFile(filename, data, 0644)
}
