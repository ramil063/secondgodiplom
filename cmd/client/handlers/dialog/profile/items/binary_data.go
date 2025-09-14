package items

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	binaryQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/binary"
	binarydataService "github.com/ramil063/secondgodiplom/cmd/client/services/items/binarydata"
)

const defaultDownloadDir = "tmp/downloads"

// WorkWithFile главное меню для работы с файлами
func WorkWithFile(service binarydataService.Servicer) dialog.AppState {
	for {
		err := dialog.ClearScreen()
		if err != nil {
			fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
		}
		showMenuFile()

		reader := bufio.NewReader(os.Stdin)
		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Ошибка считывания: %v\n", err)
			return dialog.StateMainMenu
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			err = uploadFileFromConsole(service)
			if err != nil {
				fmt.Printf("❌ Ошибка при загрузке: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "2":
			err = downloadFileFromConsole(service)
			if err != nil {
				fmt.Printf("❌ Ошибка при скачивании: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "3":
			err = listOfFilesData(service)
			if err != nil {
				fmt.Printf("❌ Ошибка при листинге данных файлов: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "4":
			err = getFileInfo(service)
			if err != nil {
				fmt.Printf("❌ Ошибка при показе данных файла: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "5":
			err = deleteFile()
			if err != nil {
				fmt.Printf("❌ Ошибка при удалении данных файла: %v\n", err)
				err = dialog.PressEnterToContinue()
				if err != nil {
					fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
				}
				continue
			}
		case "6":
			return dialog.StateMainMenu // Выход в главное меню
		default:
			fmt.Println("Неверный выбор!")
			err = dialog.PressEnterToContinue()
			if err != nil {
				fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
			}
		}
	}
}

func showMenuFile() {
	fmt.Printf("=== РАБОТА С ФАЙЛАМИ ===\n")
	fmt.Println("========================")
	fmt.Println("1. Загрузка")
	fmt.Println("2. Скачивание")
	fmt.Println("3. Получение всего списка файлов")
	fmt.Println("4. Получение по идентификатору")
	fmt.Println("5. Удаление по идентификатору")
	fmt.Println("6. Назад")
	fmt.Println("========================")
	fmt.Print("Выберите действие: ")
}

func uploadFileFromConsole(service binarydataService.Servicer) error {
	// Запрашиваем путь к файлу
	fmt.Print("Введите полный путь к файлу: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	filePath := strings.TrimSpace(scanner.Text())

	// Запрашиваем описание
	fmt.Print("Введите описание: ")
	scanner.Scan()
	description := strings.TrimSpace(scanner.Text())

	// Читаем файл
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("❌ Возникла ошибка чтения файла: %s\n", err.Error())
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("❌ Возникла ошибка получения данных из файла: %s\n", err.Error())
	}

	ctx := items.CreateAuthContext()

	response, totalChunks, err := service.UploadData(ctx, fileData, fileInfo, filePath, description)
	if err != nil {
		return fmt.Errorf("❌ Возникла ошибка: %s\n", err.Error())
	}

	fmt.Printf("✅ Файл загружен успешно!\n")
	fmt.Printf("   ID: %d\n", response.FileId)
	fmt.Printf("   Размер: %d байт\n", response.BytesReceived)
	fmt.Printf("   Количество частей: %d\n", totalChunks)
	return nil
}

func downloadFileFromConsole(service binarydataService.Servicer) error {
	// 1. Запрашиваем ID файла
	fmt.Print("Введите идентификатор файла для загрузки: ")
	var fileID int64
	_, err := fmt.Scanln(&fileID)
	if err != nil {
		return err
	}

	// 2. Запрашиваем папку для сохранения
	fmt.Print("Введите для сохранения (нажмите Enter для директории по умолчанию): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	downloadDir := strings.TrimSpace(scanner.Text())
	if downloadDir == "" {
		downloadDir = defaultDownloadDir
	}

	// 3. Создаем stream для скачивания
	ctx := items.CreateAuthContext()
	filePath, err := service.DownloadData(ctx, fileID, downloadDir)
	if err != nil {
		return err
	}
	fmt.Printf("\n✅ Файл загружен успешно: %s\n", filePath)
	return nil
}

func deleteFile() error {
	fmt.Print("Введите идентификатор файла для удаления: ")
	var fileID int64
	_, err := fmt.Scanln(&fileID)
	if err != nil {
		return fmt.Errorf("❌ Возникла ошибка: %w", err)
	}

	// Сохраняем в очередь вместо немедленной отправки
	queueID, err := binaryQueue.SaveToDeleteQueue(fileID)
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

func listOfFilesData(service binarydataService.Servicer) error {
	currentPage := int32(1)
	filter := ""

	for {
		err := dialog.ClearScreen()
		if err != nil {
			fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
		}
		fmt.Println("=== УПРАВЛЕНИЕ ФАЙЛАМИ ===")
		fmt.Printf("Страница: %d | Фильтр: '%s'\n", currentPage, filter)
		fmt.Println("===============================")

		// Получение данных с сервера
		ctx := items.CreateAuthContext()
		resp, err := service.ListFiles(ctx, currentPage, filter)
		if err != nil {
			return fmt.Errorf("❌ Ошибка получения данных: %v\n", err)
		}

		// Вывод паролей
		if len(resp.Files) == 0 {
			fmt.Println("Записей не найдено")
		} else {
			for _, val := range resp.Files {
				fmt.Printf("ID: %d\n", val.Id)
				fmt.Printf("   Имя: %s\n", val.Filename)
				fmt.Printf("   Тип: %s\n", val.MimeType)
				fmt.Printf("   Размер: %d\n", val.Size)
				fmt.Printf("   Описание: %s\n", val.Description)
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
				fmt.Println("Неверный номер страницы")
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
			fmt.Println("Неверный выбор")
			err = dialog.PressEnterToContinue()
			if err != nil {
				fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
			}
		}
	}
}

func getFileInfo(service binarydataService.Servicer) error {
	fmt.Print("Введите идентификатор файла для просмотра: ")
	var fileID int64
	_, err := fmt.Scanln(&fileID)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}

	ctx := items.CreateAuthContext()
	resp, err := service.GetFileInfo(ctx, fileID)
	if err != nil {
		return fmt.Errorf("\nОшибка получения данных по файлу: %s\n", err.Error())
	}

	fmt.Printf("Id: %d\n", resp.Id)
	fmt.Printf("   Название: %s\n", resp.Filename)
	fmt.Printf("   Тип: %s\n", resp.MimeType)
	fmt.Printf("   Размер(байт): %d\n", resp.Size)
	fmt.Printf("   Описание: %s\n", resp.Description)
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
