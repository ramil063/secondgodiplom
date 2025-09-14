package items

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	passwordQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/password"
	passwordService "github.com/ramil063/secondgodiplom/cmd/client/services/items/password"
)

// WorkWithPassword главное меню для работы с паролями
func WorkWithPassword(service passwordService.Servicer) dialog.AppState {
	for {
		err := dialog.ClearScreen()
		if err != nil {
			fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
		}
		showMenuPasswords()

		reader := bufio.NewReader(os.Stdin)
		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Ошибка считывания: %v\n", err)
			return dialog.StateMainMenu
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			err = createPassword()
			if err != nil {
				fmt.Printf("❌ Ошибка при создании данных пароля: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "2":
			err = showPassword(service)
			if err != nil {
				fmt.Printf("❌ Ошибка при показе данных пароля: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "3":
			err = listOfPasswords(service)
			if err != nil {
				fmt.Printf("❌ Ошибка при листинге данных пароля: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "4":
			err = changePassword()
			if err != nil {
				fmt.Printf("❌ Ошибка при изменении данных пароля: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "5":
			err = deletePassword()
			if err != nil {
				fmt.Printf("❌ Ошибка при удалении данных карты: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "6":
			return dialog.StateMainMenu // Выход в главное меню
		default:
			fmt.Println("❌ Неверный выбор!")
			err = dialog.PressEnterToContinue()
			if err != nil {
				fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
			}
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

func createPassword() error {
	var err error

	err = dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== СОЗДАНИЕ ДАННЫХ О ПАРОЛЕ ===")

	reader := bufio.NewReader(os.Stdin)

	var login, pwd, target, description, metaDataName, metaDataValue string

	// Сбор данных
	for {
		fmt.Print("Введите логин: ")
		login, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
		}
		login = strings.TrimSpace(login)

		if len(login) <= 0 {
			fmt.Println("❌ Логин должен содержать минимум 1 символ")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите пароль: ")
		pwd, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
		}
		pwd = strings.TrimSpace(pwd)

		if len(pwd) <= 0 {
			fmt.Println("❌ Пароль должен содержать минимум 1 символ")
			continue
		}
		break
	}

	fmt.Print("Введите систему или сайт: ")
	target, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}
	target = strings.TrimSpace(target)

	fmt.Print("Введите описание: ")
	description, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}
	description = strings.TrimSpace(description)

	fmt.Print("Введите название метаданных: ")
	metaDataName, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}
	metaDataName = strings.TrimSpace(metaDataName)

	fmt.Print("Введите значение метаданных: ")
	metaDataValue, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}
	metaDataValue = strings.TrimSpace(metaDataValue)

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := passwordQueue.SaveToCreateQueue(login, pwd, target, description, metaDataName, metaDataValue)
	if err != nil {
		return fmt.Errorf("❌ Ошибка сохранения в очередь: %v\n", err)
	}

	fmt.Printf("✅ Данные сохранены в очередь для отправки!\n")
	fmt.Printf("ID очереди: %s\n", queueID)
	fmt.Println("Данные будут отправлены на сервер в течение 30 секунд")
	fmt.Println("----------------------------------")

	err = dialog.PressEnterToContinue()
	if err != nil {
		return err
	}
	return nil
}

func showPassword(service passwordService.Servicer) error {
	err := dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== ИНФОРМАЦИЯ О ПАРОЛЕ ===")

	ctx := items.CreateAuthContext()

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	_, err = fmt.Scanln(&id)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}
	// Запрашиваем данные с сервера
	resp, err := service.GetPassword(ctx, id)

	if err != nil {
		return fmt.Errorf("❌ Ошибка получения данных\n")
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

	err = dialog.PressEnterToContinue()
	if err != nil {
		return err
	}
	return nil
}

func listOfPasswords(service passwordService.Servicer) error {
	currentPage := int32(1)
	filter := ""

	for {
		err := dialog.ClearScreen()
		if err != nil {
			fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
		}
		fmt.Println("=== УПРАВЛЕНИЕ ПАРОЛЯМИ ===")
		fmt.Printf("Страница: %d | Фильтр: '%s'\n", currentPage, filter)
		fmt.Println("===============================")

		// Получение данных с сервера
		resp, err := service.ListPasswords(items.CreateAuthContext(), currentPage, filter)
		if err != nil {
			return fmt.Errorf("❌ Ошибка получения данных: %v\n", err)
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
		choice, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1": // Следующая страница
			if currentPage < resp.TotalPages {
				currentPage++
			} else {
				fmt.Println("Это последняя страница")
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
			}

		case "2": // Предыдущая страница
			if currentPage > 1 {
				currentPage--
			} else {
				fmt.Println("Это первая страница")
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
			}

		case "3": // Ввод номера страницы
			fmt.Print("Введите номер страницы: ")
			var newPage int32
			_, err = fmt.Scanln(&newPage)
			if err != nil {
				return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
			}
			if newPage >= 1 && newPage <= resp.TotalPages {
				currentPage = newPage
			} else {
				fmt.Println("❌ Неверный номер страницы")
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
			}

		case "4": // Установить фильтр
			fmt.Print("Введите текст для фильтрации: ")
			newFilter, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
			}
			filter = strings.TrimSpace(newFilter)
			currentPage = 1 // Сброс на первую страницу при новом фильтре

		case "5": // Сбросить фильтр
			filter = ""
			currentPage = 1

		case "0": // Выход
			return nil

		default:
			fmt.Println("❌ Неверный выбор")
			err = dialog.PressEnterToContinue()
			if err != nil {
				fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
			}
		}
	}
}

func changePassword() error {
	var err error

	err = dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== ОБНОВЛЕНИЕ ДАННЫХ О ПАРОЛЕ ===")

	reader := bufio.NewReader(os.Stdin)

	var login, pwd, target, description string

	// Сбор данных

	var id int64
	for {
		fmt.Print("Введите идентификатор записи: ")
		_, err = fmt.Scanln(&id)
		if err != nil {
			return fmt.Errorf("❌ Ошибка ввода идентификатора")
		}

		if id <= 0 {
			fmt.Println("❌ Ошибка ввода идентификатора")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите логин: ")
		login, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
		}
		login = strings.TrimSpace(login)

		if len(login) <= 0 {
			fmt.Println("❌ Логин должен содержать минимум 1 символ")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите пароль: ")
		pwd, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
		}
		pwd = strings.TrimSpace(pwd)

		if len(pwd) <= 0 {
			fmt.Println("❌ Пароль должен содержать минимум 1 символ")
			continue
		}
		break
	}

	fmt.Print("Введите систему или сайт: ")
	target, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}
	target = strings.TrimSpace(target)

	fmt.Print("Введите описание: ")
	description, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}
	description = strings.TrimSpace(description)

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := passwordQueue.SaveToUpdateQueue(id, login, pwd, target, description)
	if err != nil {
		return fmt.Errorf("❌ Ошибка сохранения в очередь: %v\n", err)
	}

	fmt.Printf("✅ Данные сохранены в очередь для отправки!\n")
	fmt.Printf("ID очереди: %s\n", queueID)
	fmt.Println("Данные будут отправлены на сервер в течение 30 секунд")
	fmt.Println("----------------------------------")

	err = dialog.PressEnterToContinue()
	if err != nil {
		return err
	}
	return nil
}

func deletePassword() error {
	err := dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== ИНФОРМАЦИЯ О ПАРОЛЕ ===")

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	_, err = fmt.Scanln(&id)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := passwordQueue.SaveToDeleteQueue(id)
	if err != nil {
		return fmt.Errorf("❌ Ошибка сохранения в очередь: %v\n", err)
	}

	fmt.Printf("✅ Данные сохранены в очередь для отправки!\n")
	fmt.Printf("ID очереди: %s\n", queueID)
	fmt.Println("Данные будут отправлены на сервер в течение 30 секунд")
	fmt.Println("----------------------------------")

	err = dialog.PressEnterToContinue()
	if err != nil {
		return err
	}
	return nil
}
