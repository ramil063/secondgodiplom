package bankcard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	"github.com/ramil063/secondgodiplom/internal/constants/queue"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/bankcard"
)

// Init инициализация всего необходимого для очереди в частности создаем директорию
func Init() {
	os.MkdirAll(queue.DirBankcard, 0755)
}

// Process основной процесс работы очереди отложенных обращений на сервер
func Process(client bankcard.ServiceClient) {
	processCreate(client)
	processUpdate(client)
	processDelete(client)
}

func processCreate(client bankcard.ServiceClient) {
	files, err := filepath.Glob(filepath.Join(queue.DirBankcard, "create_*.json"))
	if err != nil {
		fmt.Printf("❌ Ошибка поиска файлов очереди: %v\n", err)
		return
	}

	if len(files) == 0 {
		return
	}

	fmt.Printf("\nНайдено %d запросов в очереди\n", len(files))

	for _, file := range files {
		sendCreate(client, file)
	}
}

func processUpdate(client bankcard.ServiceClient) {
	files, err := filepath.Glob(filepath.Join(queue.DirBankcard, "update_*.json"))
	if err != nil {
		fmt.Printf("❌ Ошибка поиска файлов очереди: %v\n", err)
		return
	}

	if len(files) == 0 {
		return
	}

	fmt.Printf("\nНайдено %d запросов в очереди\n", len(files))

	for _, file := range files {
		sendUpdate(client, file)
	}
}

func processDelete(client bankcard.ServiceClient) {
	files, err := filepath.Glob(filepath.Join(queue.DirBankcard, "delete_*.json"))
	if err != nil {
		fmt.Printf("❌ Ошибка поиска файлов очереди: %v\n", err)
		return
	}

	if len(files) == 0 {
		return
	}

	fmt.Printf("\nНайдено %d запросов в очереди\n", len(files))

	for _, file := range files {
		sendDelete(client, file)
	}
}

func sendCreate(client bankcard.ServiceClient, filename string) {
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
	resp, err := client.CreateCardData(ctx, &bankcard.CreateCardDataRequest{
		Number:          request.Number,
		ValidUntilYear:  request.ValidUntilYear,
		ValidUntilMonth: request.ValidUntilMonth,
		Cvv:             request.Cvv,
		Holder:          request.Holder,
		Description:     request.Description,
		MetaDataName:    request.MetaDataName,
		MetaDataValue:   request.MetaDataValue,
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
	fmt.Printf("✅ Запрос успешно отправлен и удален из хранилища: %s (ID: %d)\n", request.GeneratedID, resp.Id)
}

func sendUpdate(client bankcard.ServiceClient, filename string) {
	// Читаем файл
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("❌ Ошибка чтения файла %s: %v\n", filename, err)
		return
	}

	var request Request
	if err = json.Unmarshal(data, &request); err != nil {
		fmt.Printf("❌ Ошибка парсинга JSON %s: %v\n", filename, err)
		return
	}

	// Помечаем как обрабатывается
	request.Status = queue.RequestStatusProcessing
	saveRequest(filename, request)

	// Отправляем на сервер
	ctx := items.CreateAuthContext()
	resp, err := client.UpdateCardData(ctx, &bankcard.UpdateCardDataRequest{
		Id:              request.ID,
		Number:          request.Number,
		ValidUntilYear:  request.ValidUntilYear,
		ValidUntilMonth: request.ValidUntilMonth,
		Cvv:             request.Cvv,
		Holder:          request.Holder,
		Description:     request.Description,
	})

	if err != nil {
		fmt.Printf("❌ Ошибка отправки %d: %v\n", request.ID, err)
		request.RetryCount++
		request.Status = queue.RequestStatusFailed
		saveRequest(filename, request)

		// Удаляем после 3 попыток
		if request.RetryCount >= 3 {
			os.Remove(filename)
			fmt.Printf("Удален файл после 3 неудачных попыток: %d\n", request.ID)
		}
		return
	}

	// Успешная отправка - удаляем файл
	os.Remove(filename)
	fmt.Printf("✅ Запрос успешно отправлен и удален из хранилища: %s (ID: %d)\n", request.GeneratedID, resp.Id)
}

func sendDelete(client bankcard.ServiceClient, filename string) {
	// Читаем файл
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("❌ Ошибка чтения файла %s: %v\n", filename, err)
		return
	}

	var request Request
	if err = json.Unmarshal(data, &request); err != nil {
		fmt.Printf("❌ Ошибка парсинга JSON %s: %v\n", filename, err)
		return
	}

	// Помечаем как обрабатывается
	request.Status = queue.RequestStatusProcessing
	saveRequest(filename, request)

	// Отправляем на сервер
	ctx := items.CreateAuthContext()
	_, err = client.DeleteCardData(ctx, &bankcard.DeleteCardDataRequest{
		Id: request.ID,
	})

	if err != nil {
		fmt.Printf("❌ Ошибка отправки %d: %v\n", request.ID, err)
		request.RetryCount++
		request.Status = queue.RequestStatusFailed
		saveRequest(filename, request)

		// Удаляем после 3 попыток
		if request.RetryCount >= 3 {
			os.Remove(filename)
			fmt.Printf("Удален файл после 3 неудачных попыток: %d\n", request.ID)
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
