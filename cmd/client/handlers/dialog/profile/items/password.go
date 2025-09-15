package items

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/generics/list"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	passwordQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/password"
	passwordService "github.com/ramil063/secondgodiplom/cmd/client/services/items/password"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/password"
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
			err = list.ShowListData(service, displayPassword)
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

func displayPassword(val *password.PasswordItem) {
	fmt.Printf("ID: %d\n", val.Id)
	fmt.Printf("   Цель: %s\n", val.Target)
	fmt.Printf("   Логин: %s\n", val.Login)
	fmt.Printf("   Пароль: %s\n", val.Password)
	fmt.Printf("   Создано: %s\n", val.CreatedAt)
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
