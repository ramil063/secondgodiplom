package items

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	passwordQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/password"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/password"
)

// WorkWithPassword главное меню для работы с паролями
func WorkWithPassword(client password.ServiceClient) dialog.AppState {
	for {
		dialog.ClearScreen()
		showMenuPasswords()

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			createPassword()
		case "2":
			showPassword(client)
		case "3":
			listOfPasswords(client)
		case "4":
			changePassword()
		case "5":
			deletePassword()
		case "6":
			return dialog.StateMainMenu // Выход в главное меню
		default:
			fmt.Println("❌ Неверный выбор!")
			dialog.PressEnterToContinue()
		}
	}
}

func showMenuPasswords() {
	fmt.Printf("=== РАБОТА С ДАННЫМИ ПО ЛОГИНУ И ПАРОЛЮ ===\n")
	fmt.Println("========================")
	fmt.Println("1. Создание")
	fmt.Println("2. Получение по идентификатору")
	fmt.Println("3. Получение всего списка")
	fmt.Println("4. Обновление данных")
	fmt.Println("5. Удаление по идентификатору")
	fmt.Println("6. Назад")
	fmt.Println("========================")
	fmt.Print("Выберите действие: ")
}

func createPassword() {
	dialog.ClearScreen()
	fmt.Println("=== СОЗДАНИЕ ДАННЫХ О ПАРОЛЕ ===")

	//ctx := items.CreateAuthContext()

	reader := bufio.NewReader(os.Stdin)

	var login, pwd, target, description, metaDataName, metaDataValue string

	// Сбор данных
	for {
		fmt.Print("Введите логин: ")
		login, _ = reader.ReadString('\n')
		login = strings.TrimSpace(login)

		if len(login) <= 0 {
			fmt.Println("❌ Логин должен содержать минимум 1 символ")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите пароль: ")
		pwd, _ = reader.ReadString('\n')
		pwd = strings.TrimSpace(pwd)

		if len(pwd) <= 0 {
			fmt.Println("❌ Пароль должен содержать минимум 1 символ")
			continue
		}
		break
	}

	fmt.Print("Введите систему или сайт: ")
	target, _ = reader.ReadString('\n')
	target = strings.TrimSpace(target)

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
	queueID, err := passwordQueue.SaveToCreateQueue(login, pwd, target, description, metaDataName, metaDataValue)
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

func showPassword(client password.ServiceClient) {
	dialog.ClearScreen()
	fmt.Println("=== ИНФОРМАЦИЯ О ПАРОЛЕ ===")

	ctx := items.CreateAuthContext()

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	fmt.Scanln(&id)
	// Запрашиваем данные с сервера
	resp, err := client.GetPassword(ctx, &password.GetPasswordRequest{
		Id: id,
	})

	if err != nil {
		fmt.Printf("❌ Ошибка получения данных\n")
		dialog.PressEnterToContinue()
		return
	}

	fmt.Printf("Id: %d\n", resp.Id)
	fmt.Printf("   Логин: %s\n", resp.Login)
	fmt.Printf("   Пароль: %s\n", resp.Password)
	fmt.Printf("   Система: %s\n", resp.Target)
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

func listOfPasswords(client password.ServiceClient) {
	currentPage := int32(1)
	filter := ""

	for {
		dialog.ClearScreen()
		fmt.Println("=== УПРАВЛЕНИЕ ПАРОЛЯМИ ===")
		fmt.Printf("Страница: %d | Фильтр: '%s'\n", currentPage, filter)
		fmt.Println("===============================")

		// Получение данных с сервера
		resp, err := client.ListPasswords(items.CreateAuthContext(), &password.ListPasswordsRequest{
			Page:   currentPage,
			Filter: filter,
		})
		if err != nil {
			fmt.Printf("❌ Ошибка получения данных: %v\n", err)
			dialog.PressEnterToContinue()
			return
		}

		// Вывод паролей
		if len(resp.Passwords) == 0 {
			fmt.Println("Записей не найдено")
		} else {
			for _, val := range resp.Passwords {
				fmt.Printf("ID: %d\n", val.Id)
				fmt.Printf("   Цель: %s\n", val.Target)
				fmt.Printf("   Логин: %s\n", val.Login)
				fmt.Printf("   Пароль: %s\n", val.Password)
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

func changePassword() {
	dialog.ClearScreen()
	fmt.Println("=== ОБНОВЛЕНИЕ ДАННЫХ О ПАРОЛЕ ===")

	reader := bufio.NewReader(os.Stdin)

	var login, pwd, target, description string

	// Сбор данных

	var id int64
	for {
		fmt.Print("Введите идентификатор записи: ")
		_, err := fmt.Scanln(&id)
		if err != nil {
			fmt.Println("❌ Ошибка ввода идентификатора")
			continue
		}

		if id <= 0 {
			fmt.Println("❌ Ошибка ввода идентификатора")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите логин: ")
		login, _ = reader.ReadString('\n')
		login = strings.TrimSpace(login)

		if len(login) <= 0 {
			fmt.Println("❌ Логин должен содержать минимум 1 символ")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите пароль: ")
		pwd, _ = reader.ReadString('\n')
		pwd = strings.TrimSpace(pwd)

		if len(pwd) <= 0 {
			fmt.Println("❌ Пароль должен содержать минимум 1 символ")
			continue
		}
		break
	}

	fmt.Print("Введите систему или сайт: ")
	target, _ = reader.ReadString('\n')
	target = strings.TrimSpace(target)

	fmt.Print("Введите описание: ")
	description, _ = reader.ReadString('\n')
	description = strings.TrimSpace(description)

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := passwordQueue.SaveToUpdateQueue(id, login, pwd, target, description)
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

func deletePassword() {
	dialog.ClearScreen()
	fmt.Println("=== ИНФОРМАЦИЯ О ПАРОЛЕ ===")

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	fmt.Scanln(&id)

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := passwordQueue.SaveToDeleteQueue(id)
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
