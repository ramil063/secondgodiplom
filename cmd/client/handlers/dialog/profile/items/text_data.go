package items

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	textdataQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/textdata"
	textdataService "github.com/ramil063/secondgodiplom/cmd/client/services/items/textdata"
)

// WorkWithTextData главное меню для работы с текстовыми данными
func WorkWithTextData(service textdataService.Servicer) dialog.AppState {
	for {
		err := dialog.ClearScreen()
		if err != nil {
			fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
		}
		showMenuTextData()

		reader := bufio.NewReader(os.Stdin)
		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Ошибка считывания: %v\n", err)
			return dialog.StateMainMenu
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			err = createTextData()
			if err != nil {
				fmt.Printf("❌ Ошибка при создании данных: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "2":
			err = showTextData(service)
			if err != nil {
				fmt.Printf("❌ Ошибка при показе данных: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "3":
			err = listOfTextData(service)
			if err != nil {
				fmt.Printf("❌ Ошибка при листинге данных: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "4":
			err = changeTextData()
			if err != nil {
				fmt.Printf("❌ Ошибка при изменении данных: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "5":
			err = deleteTextData()
			if err != nil {
				fmt.Printf("❌ Ошибка при удалении данных: %v\n", err)
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

func createTextData() error {
	var err error

	err = dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== СОЗДАНИЕ ДАННЫХ ===")

	reader := bufio.NewReader(os.Stdin)

	var text, description, metaDataName, metaDataValue string

	// Сбор данных с валидацией
	fmt.Print("Введите текст: ")
	text, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}
	text = strings.TrimSpace(text)

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
	queueID, err := textdataQueue.SaveToCreateQueue(text, description, metaDataName, metaDataValue)
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

func showTextData(service textdataService.Servicer) error {
	err := dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== ИНФОРМАЦИЯ ===")

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	_, err = fmt.Scanln(&id)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}

	resp, err := service.GetTextData(items.CreateAuthContext(), id)
	if err != nil {
		return fmt.Errorf("❌ Ошибка получения данных\n")
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

	err = dialog.PressEnterToContinue()
	if err != nil {
		return err
	}
	return nil
}

func listOfTextData(service textdataService.Servicer) error {
	currentPage := int32(1)
	filter := ""

	for {
		err := dialog.ClearScreen()
		if err != nil {
			fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
		}
		fmt.Println("=== СПИСОК ===")
		fmt.Printf("Страница: %d | Фильтр: '%s'\n", currentPage, filter)
		fmt.Println("===============================")

		// Получение данных с сервера
		resp, err := service.ListTextDataItems(items.CreateAuthContext(), currentPage, filter)
		if err != nil {
			return fmt.Errorf("❌ Ошибка получения данных: %v\n", err)
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
				fmt.Println("ℹ️ Это последняя страница")
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
			}

		case "2": // Предыдущая страница
			if currentPage > 1 {
				currentPage--
			} else {
				fmt.Println("ℹ️ Это первая страница")
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

func changeTextData() error {
	err := dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== ОБНОВЛЕНИЕ ДАННЫХ ===")

	reader := bufio.NewReader(os.Stdin)

	var text, description string

	// Сбор данных
	fmt.Print("Введите идентификатор записи: ")
	var id int64
	_, err = fmt.Scanln(&id)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}

	fmt.Print("Введите новый текст: ")
	text, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}
	text = strings.TrimSpace(text)

	fmt.Print("Введите описание: ")
	description, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}
	description = strings.TrimSpace(description)

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := textdataQueue.SaveToUpdateQueue(id, text, description)
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

func deleteTextData() error {
	err := dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== УДАЛЕНИЕ ===")

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	_, err = fmt.Scanln(&id)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := textdataQueue.SaveToDeleteQueue(id)
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
