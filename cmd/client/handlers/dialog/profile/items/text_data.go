package items

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	textdataQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/textdata"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/textdata"
)

// WorkWithTextData главное меню для работы с текстовыми данными
func WorkWithTextData(client textdata.ServiceClient) dialog.AppState {
	for {
		dialog.ClearScreen()
		showMenuTextData()

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			createTextData()
		case "2":
			showTextData(client)
		case "3":
			listOfTextData(client)
		case "4":
			changeTextData()
		case "5":
			deleteTextData()
		case "6":
			return dialog.StateMainMenu // Выход в главное меню
		default:
			fmt.Println("❌ Неверный выбор!")
			dialog.PressEnterToContinue()
		}
	}
}

func showMenuTextData() {
	fmt.Printf("=== РАБОТА С ТЕКСТОВЫМИ ДАННЫМИ ===\n")
	fmt.Println("========================")
	fmt.Println("1. Создание")
	fmt.Println("2. Получение по идентификатору")
	fmt.Println("3. Получение всего списка")
	fmt.Println("4. Обновление данных")
	fmt.Println("5. Удаление по идентификатору")
	fmt.Println("6. Вернуться")
	fmt.Println("========================")
	fmt.Print("Выберите действие: ")
}

func createTextData() {
	dialog.ClearScreen()
	fmt.Println("=== СОЗДАНИЕ ДАННЫХ ===")

	reader := bufio.NewReader(os.Stdin)

	var text, description, metaDataName, metaDataValue string

	// Сбор данных с валидацией
	fmt.Print("Введите текст: ")
	text, _ = reader.ReadString('\n')
	text = strings.TrimSpace(text)

	fmt.Print("Введите описание: ")
	description, _ = reader.ReadString('\n')
	description = strings.TrimSpace(description)

	fmt.Print("Введите название метаданных: ")
	metaDataName, _ = reader.ReadString('\n')
	metaDataName = strings.TrimSpace(metaDataName)

	fmt.Print("Введите значение метаданных: ")
	metaDataValue, _ = reader.ReadString('\n')
	metaDataValue = strings.TrimSpace(metaDataValue)

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := textdataQueue.SaveToCreateQueue(text, description, metaDataName, metaDataValue)
	if err != nil {
		fmt.Printf("❌ Ошибка сохранения в очередь: %v\n", err)
		dialog.PressEnterToContinue()
		return
	}

	fmt.Printf("✅ Данные сохранены в очередь для отправки!\n")
	fmt.Printf("ID очереди: %s\n", queueID)
	fmt.Println("Данные будут отправлены на сервер в течение 30 секунд")
	fmt.Println("----------------------------------")

	dialog.PressEnterToContinue()
}

func showTextData(client textdata.ServiceClient) {
	dialog.ClearScreen()
	fmt.Println("=== ИНФОРМАЦИЯ ===")

	ctx := items.CreateAuthContext()

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	fmt.Scanln(&id)
	// Запрашиваем данные с сервера
	resp, err := client.GetTextData(ctx, &textdata.GetTextDataRequest{
		Id: id,
	})

	if err != nil {
		fmt.Printf("❌ Ошибка получения данных\n")
		dialog.PressEnterToContinue()
		return
	}

	fmt.Printf("Id: %d\n", resp.Id)
	fmt.Printf("   Текст: %s\n", resp.TextData)
	fmt.Printf("   Создано: %s\n", resp.CreatedAt)
	if len(resp.MetaData) > 0 {
		fmt.Println("   Метаданные ---")
	}
	for _, val := range resp.MetaData {
		fmt.Printf("        Идентификатор: %d\n", val.Id)
		fmt.Printf("        Название: %s\n", val.Name)
		fmt.Printf("        Значение: %s\n", val.Value)
		fmt.Println("        -----------")
	}
	fmt.Println("---")

	dialog.PressEnterToContinue()
}

func listOfTextData(client textdata.ServiceClient) {
	currentPage := int32(1)
	filter := ""

	for {
		dialog.ClearScreen()
		fmt.Println("=== СПИСОК ===")
		fmt.Printf("Страница: %d | Фильтр: '%s'\n", currentPage, filter)
		fmt.Println("===============================")

		// Получение данных с сервера
		resp, err := client.ListTextDataItems(items.CreateAuthContext(), &textdata.ListTextDataRequest{
			Page:   currentPage,
			Filter: filter,
		})
		if err != nil {
			fmt.Printf("❌ Ошибка получения данных: %v\n", err)
			dialog.PressEnterToContinue()
			return
		}

		// Вывод паролей
		if len(resp.TextDataItems) == 0 {
			fmt.Println("Записей не найдено")
		} else {
			for _, val := range resp.TextDataItems {
				fmt.Printf("ID: %d\n", val.Id)
				fmt.Printf("    Текст: %s\n", val.TextData)
				fmt.Printf("    Создано: %s\n", val.CreatedAt)
				fmt.Println("---")
			}

			// Информация о пагинации
			fmt.Printf("Страница %d из %d | Всего записей: %d\n",
				currentPage, resp.TotalPages, resp.TotalCount)
		}

		// Меню навигации
		fmt.Println("\n===============================")
		fmt.Println("1. Следующая страница →")
		fmt.Println("2. Предыдущая страница ←")
		fmt.Println("3. Ввести номер страницы")
		fmt.Println("4. Установить фильтр")
		fmt.Println("5. Сбросить фильтр")
		fmt.Println("0. Вернуться")
		fmt.Println("===============================")
		fmt.Print("Выберите действие: ")

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1": // Следующая страница
			if currentPage < resp.TotalPages {
				currentPage++
			} else {
				fmt.Println("ℹ️ Это последняя страница")
				dialog.PressEnterToContinue()
			}

		case "2": // Предыдущая страница
			if currentPage > 1 {
				currentPage--
			} else {
				fmt.Println("ℹ️ Это первая страница")
				dialog.PressEnterToContinue()
			}

		case "3": // Ввод номера страницы
			fmt.Print("Введите номер страницы: ")
			var newPage int32
			_, err := fmt.Scanln(&newPage)
			if err == nil && newPage >= 1 && newPage <= resp.TotalPages {
				currentPage = newPage
			} else {
				fmt.Println("❌ Неверный номер страницы")
				dialog.PressEnterToContinue()
			}

		case "4": // Установить фильтр
			fmt.Print("Введите текст для фильтрации: ")
			newFilter, _ := reader.ReadString('\n')
			filter = strings.TrimSpace(newFilter)
			currentPage = 1 // Сброс на первую страницу при новом фильтре

		case "5": // Сбросить фильтр
			filter = ""
			currentPage = 1

		case "0": // Выход
			return

		default:
			fmt.Println("❌ Неверный выбор")
			dialog.PressEnterToContinue()
		}
	}
}

func changeTextData() {
	dialog.ClearScreen()
	fmt.Println("=== ОБНОВЛЕНИЕ ДАННЫХ ===")

	reader := bufio.NewReader(os.Stdin)

	var text, description string

	// Сбор данных
	fmt.Print("Введите идентификатор записи: ")
	var id int64
	_, err := fmt.Scanln(&id)

	fmt.Print("Введите новый текст: ")
	text, _ = reader.ReadString('\n')
	text = strings.TrimSpace(text)

	fmt.Print("Введите описание: ")
	description, _ = reader.ReadString('\n')
	description = strings.TrimSpace(description)

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := textdataQueue.SaveToUpdateQueue(id, text, description)
	if err != nil {
		fmt.Printf("❌ Ошибка сохранения в очередь: %v\n", err)
		dialog.PressEnterToContinue()
		return
	}

	fmt.Printf("✅ Данные сохранены в очередь для отправки!\n")
	fmt.Printf("ID очереди: %s\n", queueID)
	fmt.Println("Данные будут отправлены на сервер в течение 30 секунд")
	fmt.Println("----------------------------------")

	dialog.PressEnterToContinue()
}

func deleteTextData() {
	dialog.ClearScreen()
	fmt.Println("=== УДАЛЕНИЕ ===")

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	fmt.Scanln(&id)

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := textdataQueue.SaveToDeleteQueue(id)
	if err != nil {
		fmt.Printf("❌ Ошибка сохранения в очередь: %v\n", err)
		dialog.PressEnterToContinue()
		return
	}

	fmt.Printf("✅ Данные сохранены в очередь для отправки!\n")
	fmt.Printf("ID очереди: %s\n", queueID)
	fmt.Println("Данные будут отправлены на сервер в течение 30 секунд")
	fmt.Println("----------------------------------")

	dialog.PressEnterToContinue()
}
