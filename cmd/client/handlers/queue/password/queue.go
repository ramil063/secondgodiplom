package password

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	"github.com/ramil063/secondgodiplom/internal/constants/queue"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/password"
)

func Init() {
	// Создаем директорию для очереди
	os.MkdirAll(queue.DirPassword, 0755)
}

func Process(client password.ServiceClient) {
	processCreate(client)
	processUpdate(client)
	processDelete(client)
}

func processCreate(client password.ServiceClient) {
	files, err := filepath.Glob(filepath.Join(queue.DirPassword, "create_*.json"))
	if err != nil {
		fmt.Printf("❌ Ошибка поиска файлов очереди: %v\n", err)
		return
	}

	if len(files) == 0 {
		return
	}

	fmt.Printf("Найдено %d запросов в очереди\n", len(files))

	for _, file := range files {
		sendCreate(client, file)
	}
}

func processUpdate(client password.ServiceClient) {
	files, err := filepath.Glob(filepath.Join(queue.DirPassword, "update_*.json"))
	if err != nil {
		fmt.Printf("❌ Ошибка поиска файлов очереди: %v\n", err)
		return
	}

	if len(files) == 0 {
		return
	}

	fmt.Printf("Найдено %d запросов в очереди\n", len(files))

	for _, file := range files {
		sendUpdate(client, file)
	}
}

func processDelete(client password.ServiceClient) {
	files, err := filepath.Glob(filepath.Join(queue.DirPassword, "delete_*.json"))
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

func sendCreate(client password.ServiceClient, filename string) {
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
	resp, err := client.CreatePassword(ctx, &password.CreatePasswordRequest{
		Login:         request.Login,
		Password:      request.Password,
		Target:        request.Target,
		Description:   request.Description,
		MetaDataName:  request.MetaDataName,
		MetaDataValue: request.MetaDataValue,
	})

	if err != nil {
		fmt.Printf("❌ Ошибка отправки %s: %v\n", request.ID, err)
		request.RetryCount++
		request.Status = queue.RequestStatusFailed
		saveRequest(filename, request)

		// Удаляем после 3 попыток
		if request.RetryCount >= 3 {
			os.Remove(filename)
			fmt.Printf("Удален файл после 3 неудачных попыток: %s\n", request.ID)
		}
		return
	}

	// Успешная отправка - удаляем файл
	os.Remove(filename)
	fmt.Printf("✅ Запрос успешно отправлен и удален из хранилища: %s (ID: %d)\n", request.ID, resp.Id)
}

func sendUpdate(client password.ServiceClient, filename string) {
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
	intID, err := strconv.Atoi(request.ID)
	resp, err := client.UpdatePassword(ctx, &password.UpdatePasswordRequest{
		Id:          int64(intID),
		Login:       request.Login,
		Password:    request.Password,
		Target:      request.Target,
		Description: request.Description,
	})

	if err != nil {
		fmt.Printf("❌ Ошибка отправки %s: %v\n", request.ID, err)
		request.RetryCount++
		request.Status = queue.RequestStatusFailed
		saveRequest(filename, request)

		// Удаляем после 3 попыток
		if request.RetryCount >= 3 {
			os.Remove(filename)
			fmt.Printf("Удален файл после 3 неудачных попыток: %s\n", request.ID)
		}
		return
	}

	// Успешная отправка - удаляем файл
	os.Remove(filename)
	fmt.Printf("✅ Запрос успешно отправлен и удален из хранилища: %s (ID: %d)\n", request.ID, resp.Id)
}

func sendDelete(client password.ServiceClient, filename string) {
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
	intID, err := strconv.Atoi(request.ID)
	_, err = client.DeletePassword(ctx, &password.DeletePasswordRequest{
		Id: int64(intID),
	})

	if err != nil {
		fmt.Printf("❌ Ошибка отправки %s: %v\n", request.ID, err)
		request.RetryCount++
		request.Status = queue.RequestStatusFailed
		saveRequest(filename, request)

		// Удаляем после 3 попыток
		if request.RetryCount >= 3 {
			os.Remove(filename)
			fmt.Printf("Удален файл после 3 неудачных попыток: %s\n", request.ID)
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
