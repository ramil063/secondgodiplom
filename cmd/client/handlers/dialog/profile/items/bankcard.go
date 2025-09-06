package items

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	bankcardQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/bankcard"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/bankcard"
)

func WorkWithBankCardData(client bankcard.ServiceClient) dialog.AppState {
	for {
		dialog.ClearScreen()
		showMenuBankCardData()

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			createBankCardData()
		case "2":
			showBankCardData(client)
		case "3":
			listOfBankCardData(client)
		case "4":
			changeBankCardData()
		case "5":
			deleteBankCardData()
		case "6":
			return dialog.StateMainMenu // Выход в главное меню
		default:
			fmt.Println("❌ Неверный выбор!")
			dialog.PressEnterToContinue()
		}
	}
}

func showMenuBankCardData() {
	fmt.Printf("=== РАБОТА С ДАННЫМИ БАНКОВСКИХ КАРТ ===\n")
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

func createBankCardData() {
	dialog.ClearScreen()
	fmt.Println("=== СОЗДАНИЕ ДАННЫХ ===")

	reader := bufio.NewReader(os.Stdin)

	var number, holder, description, metaDataName, metaDataValue string
	var year, month, cvv int32

	// Сбор данных
	for {
		fmt.Print("Введите номер: ")
		number, _ = reader.ReadString('\n')
		number = strings.TrimSpace(number)

		if len(number) <= 3 {
			fmt.Println("❌ Номер должен содержать минимум 4 символа")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите год: ")
		fmt.Scanln(&year)

		if year > 2000 && year < 3000 {
			fmt.Println("❌ Неверно введен год")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите месяц: ")
		fmt.Scanln(&month)

		if month > 0 && month < 13 {
			fmt.Println("❌ Неверно введен месяц")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите CVV код: ")
		fmt.Scanln(&cvv)

		if cvv > 99 && cvv < 1000 {
			fmt.Println("❌ Неверно введен код")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите фамилию и имя держателя: ")
		holder, _ = reader.ReadString('\n')
		holder = strings.TrimSpace(holder)

		if len(holder) <= 3 {
			fmt.Println("❌ Поле должно содержать минимум 3 символа")
			continue
		}
		if !strings.Contains(holder, " ") {
			fmt.Println("❌ Введите фамилию и имя через пробел")
			continue
		}
		parts := strings.Split(holder, " ")
		if len(parts) < 2 {
			fmt.Println("❌ Требуется только 2 слова")
			continue
		}
		break
	}

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
	queueID, err := bankcardQueue.SaveToCreateQueue(
		number,
		year,
		month,
		cvv,
		holder,
		description,
		metaDataName,
		metaDataValue)
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

func showBankCardData(client bankcard.ServiceClient) {
	dialog.ClearScreen()
	fmt.Println("=== ИНФОРМАЦИЯ ===")

	ctx := items.CreateAuthContext()

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	fmt.Scanln(&id)
	// Запрашиваем данные с сервера
	resp, err := client.GetCardData(ctx, &bankcard.GetCardDataRequest{
		Id: id,
	})

	if err != nil {
		fmt.Printf("❌ Ошибка получения данных\n")
		dialog.PressEnterToContinue()
		return
	}

	fmt.Printf("Id: %d\n", resp.Id)
	fmt.Printf("   Текст: %s\n", resp.Number)
	fmt.Printf("   Годен до год/месяц: %s\n", strconv.Itoa(int(resp.ValidUntilYear))+"/"+strconv.Itoa(int(resp.ValidUntilMonth)))
	fmt.Printf("   CVV код: %d\n", resp.Cvv)
	fmt.Printf("   Держатель карты: %s\n", resp.Holder)
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

func listOfBankCardData(client bankcard.ServiceClient) {
	currentPage := int32(1)
	filter := ""

	for {
		dialog.ClearScreen()
		fmt.Println("=== СПИСОК ===")
		fmt.Printf("Страница: %d | Фильтр: '%s'\n", currentPage, filter)
		fmt.Println("===============================")

		// Получение данных с сервера
		resp, err := client.ListCardsData(items.CreateAuthContext(), &bankcard.ListCardsDataRequest{
			Page:   currentPage,
			Filter: filter,
		})
		if err != nil {
			fmt.Printf("❌ Ошибка получения данных: %v\n", err)
			dialog.PressEnterToContinue()
			return
		}

		// Вывод паролей
		if len(resp.Cards) == 0 {
			fmt.Println("Записей не найдено")
		} else {
			for _, val := range resp.Cards {
				fmt.Printf("ID: %d\n", val.Id)
				fmt.Printf("   Текст: %s\n", val.Number)
				fmt.Printf("   Годен до год/месяц: %s\n", strconv.Itoa(int(val.ValidUntilYear))+"/"+strconv.Itoa(int(val.ValidUntilMonth)))
				fmt.Printf("   CVV код: %d\n", val.Cvv)
				fmt.Printf("   Держатель карты: %s\n", val.Holder)
				fmt.Printf("   Создано: %s\n", val.CreatedAt)
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
				fmt.Println("Это последняя страница")
				dialog.PressEnterToContinue()
			}

		case "2": // Предыдущая страница
			if currentPage > 1 {
				currentPage--
			} else {
				fmt.Println("Это первая страница")
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

func changeBankCardData() {
	dialog.ClearScreen()
	fmt.Println("=== ОБНОВЛЕНИЕ ДАННЫХ ===")

	reader := bufio.NewReader(os.Stdin)

	var number, holder, description string

	// Сбор данных
	fmt.Print("Введите идентификатор записи: ")
	var id int64
	_, err := fmt.Scanln(&id)

	// Сбор данных
	fmt.Print("Введите номер: ")
	number, _ = reader.ReadString('\n')
	number = strings.TrimSpace(number)

	fmt.Print("Введите год: ")
	var year int32
	fmt.Scanln(&year)

	fmt.Print("Введите месяц: ")
	var month int32
	fmt.Scanln(&month)

	fmt.Print("Введите CVV код: ")
	var cvv int32
	fmt.Scanln(&cvv)

	fmt.Print("Введите фамилию и имя держателя: ")
	holder, _ = reader.ReadString('\n')
	holder = strings.TrimSpace(holder)

	fmt.Print("Введите описание: ")
	description, _ = reader.ReadString('\n')
	description = strings.TrimSpace(description)

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := bankcardQueue.SaveToUpdateQueue(
		id,
		number,
		year,
		month,
		cvv,
		holder,
		description)
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

func deleteBankCardData() {
	dialog.ClearScreen()
	fmt.Println("=== УДАЛЕНИЕ ===")

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	fmt.Scanln(&id)

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := bankcardQueue.SaveToDeleteQueue(id)
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
