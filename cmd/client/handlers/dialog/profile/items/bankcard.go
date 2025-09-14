package items

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	bankcardQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/bankcard"
	bankcardService "github.com/ramil063/secondgodiplom/cmd/client/services/items/bankcard"
)

// WorkWithBankCardData главное меню для работы с банковской картой
func WorkWithBankCardData(service bankcardService.Servicer) dialog.AppState {
	for {
		err := dialog.ClearScreen()
		if err != nil {
			fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
		}
		showMenuBankCardData()

		reader := bufio.NewReader(os.Stdin)
		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Ошибка считывания: %v\n", err)
			return dialog.StateMainMenu
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			err = createBankCardData()
			if err != nil {
				fmt.Printf("❌ Ошибка при создании данных карты: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "2":
			err = showBankCardData(service)
			if err != nil {
				fmt.Printf("❌ Ошибка при показе данных карты: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "3":
			err = listOfBankCardData(service)
			if err != nil {
				fmt.Printf("❌ Ошибка при листинге данных карты: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "4":
			err = changeBankCardData()
			if err != nil {
				fmt.Printf("❌ Ошибка при изменении данных карты: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "5":
			err = deleteBankCardData()
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

func createBankCardData() error {
	var err error

	err = dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== СОЗДАНИЕ ДАННЫХ ===")

	reader := bufio.NewReader(os.Stdin)

	var number, holder, description, metaDataName, metaDataValue string
	var year, month, cvv int32

	// Сбор данных
	for {
		fmt.Print("Введите номер: ")
		number, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания номера: %s\n", err)
		}
		number = strings.TrimSpace(number)

		if len(number) <= 3 {
			fmt.Println("❌ Номер должен содержать минимум 4 символа")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите год: ")
		_, err = fmt.Scanln(&year)
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания года: %v\n", err)
		}

		if year < 2000 || year > 3000 {
			fmt.Println("❌ Неверно введен год")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите месяц: ")
		_, err = fmt.Scanln(&month)
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания месяца: %v\n", err)
		}

		if month < 1 || month > 12 {
			fmt.Println("❌ Неверно введен месяц")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите CVV код: ")
		_, err = fmt.Scanln(&cvv)
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания кода: %v\n", err)
		}

		if cvv < 100 || cvv > 999 {
			fmt.Println("❌ Неверно введен код")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите фамилию и имя держателя: ")
		holder, err = reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания данных держателя: %v\n", err)
		}
		holder = strings.TrimSpace(holder)

		if len(holder) < 3 {
			fmt.Println("❌ Поле должно содержать минимум 3 символа")
			continue
		}
		if !strings.Contains(holder, " ") {
			fmt.Println("❌ Введите фамилию и имя через пробел")
			continue
		}
		parts := strings.Split(holder, " ")
		if len(parts) > 2 {
			fmt.Println("❌ Требуется только 2 слова")
			continue
		}
		break
	}

	fmt.Print("Введите описание: ")
	description, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания описания: %v\n", err)
	}
	description = strings.TrimSpace(description)

	fmt.Print("Введите название метаданных: ")
	metaDataName, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания названия: %v\n", err)
	}
	metaDataName = strings.TrimSpace(metaDataName)

	fmt.Print("Введите значение метаданных: ")
	metaDataValue, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания значения: %v\n", err)
	}
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

func showBankCardData(service bankcardService.Servicer) error {
	err := dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== ИНФОРМАЦИЯ ===")

	ctx := items.CreateAuthContext()

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	_, err = fmt.Scanln(&id)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания идентификатора: %v\n", err)
	}
	// Запрашиваем данные с сервера
	resp, err := service.GetCardData(ctx, id)

	if err != nil {
		return fmt.Errorf("❌ Ошибка получения данных\n")
	}

	fmt.Printf("Id: %d\n", resp.Id)
	fmt.Printf("   Текст: %s\n", resp.Number)
	fmt.Printf("   Годен до год/месяц: %d/%d\n", resp.ValidUntilYear, resp.ValidUntilMonth)
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

	err = dialog.PressEnterToContinue()
	if err != nil {
		return err
	}
	return nil
}

func listOfBankCardData(service bankcardService.Servicer) error {
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
		ctx := items.CreateAuthContext()
		resp, err := service.ListCardsData(ctx, currentPage, filter)
		if err != nil {
			return fmt.Errorf("❌ Ошибка получения данных: %v\n", err)
		}

		// Вывод паролей
		if len(resp.Cards) == 0 {
			fmt.Println("Записей не найдено")
		} else {
			for _, val := range resp.Cards {
				fmt.Printf("ID: %d\n", val.Id)
				fmt.Printf("   Текст: %s\n", val.Number)
				fmt.Printf("   Годен до год/месяц: %d/%d\n", val.ValidUntilYear, val.ValidUntilMonth)
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
		choice, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
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
				return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
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
				return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
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

func changeBankCardData() error {
	err := dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== ОБНОВЛЕНИЕ ДАННЫХ ===")

	reader := bufio.NewReader(os.Stdin)

	var number, holder, description string

	// Сбор данных
	fmt.Print("Введите идентификатор записи: ")
	var id int64
	_, err = fmt.Scanln(&id)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
	}

	// Сбор данных
	fmt.Print("Введите номер: ")
	number, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
	}
	number = strings.TrimSpace(number)

	fmt.Print("Введите год: ")
	var year int32
	_, err = fmt.Scanln(&year)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
	}

	fmt.Print("Введите месяц: ")
	var month int32
	_, err = fmt.Scanln(&month)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
	}

	fmt.Print("Введите CVV код: ")
	var cvv int32
	_, err = fmt.Scanln(&cvv)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
	}

	fmt.Print("Введите фамилию и имя держателя: ")
	holder, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
	}
	holder = strings.TrimSpace(holder)

	fmt.Print("Введите описание: ")
	description, err = reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
	}
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

func deleteBankCardData() error {
	err := dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
	}
	fmt.Println("=== УДАЛЕНИЕ ===")

	fmt.Print("Введите идентификатор записи: ")
	var id int64
	_, err = fmt.Scanln(&id)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %v\n", err)
	}

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := bankcardQueue.SaveToDeleteQueue(id)
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
