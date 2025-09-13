package items

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/items"
	binaryQueue "github.com/ramil063/secondgodiplom/cmd/client/handlers/queue/binary"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
)

const chunkSize = 64 * 1024 // 64KB
const numberOfWorkers = 4
const defaultDownloadDir = "tmp/downloads"
const numberOfChunks = 20

var writeMutex sync.Mutex

// WorkWithFile главное меню для работы с файлами
func WorkWithFile(client binarydata.ServiceClient) dialog.AppState {
	for {
		dialog.ClearScreen()
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
			err = uploadFileFromConsole(client)
			if err != nil {
				fmt.Printf("❌ Ошибка при загрузке: %v\n", err)
				dialog.PressEnterToContinue()
				continue
			}
		case "2":
			err = downloadFileFromConsole(client)
			if err != nil {
				fmt.Printf("❌ Ошибка при скачивании: %v\n", err)
				dialog.PressEnterToContinue()
				continue
			}
		case "3":
			err = listOfFilesData(client)
			if err != nil {
				fmt.Printf("❌ Ошибка при листинге данных файлов: %v\n", err)
				dialog.PressEnterToContinue()
				continue
			}
		case "4":
			err = getFileInfo(client)
			if err != nil {
				fmt.Printf("❌ Ошибка при показе данных файла: %v\n", err)
				dialog.PressEnterToContinue()
				continue
			}
		case "5":
			err = deleteFile()
			if err != nil {
				fmt.Printf("❌ Ошибка при удалении данных файла: %v\n", err)
				dialog.PressEnterToContinue()
				continue
			}
		case "6":
			return dialog.StateMainMenu // Выход в главное меню
		default:
			fmt.Println("Неверный выбор!")
			dialog.PressEnterToContinue()
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

func uploadFileFromConsole(client binarydata.ServiceClient) error {
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
		return fmt.Errorf("Возникла ошибка чтения файла: %s\n", err.Error())
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("Возникла ошибка получения данных из файла: %s\n", err.Error())
	}

	// Создаем stream
	ctx := items.CreateAuthContext()
	stream, err := client.UploadFile(ctx)
	if err != nil {
		return fmt.Errorf("Возникла ошибка: %s\n", err.Error())
	}

	// 1. Отправляем метаданные
	totalChunks := (len(fileData) + chunkSize - 1) / chunkSize

	err = stream.Send(&binarydata.UploadFileRequest{
		Data: &binarydata.UploadFileRequest_Metadata{
			Metadata: &binarydata.FileMetadata{
				Filename:     fileInfo.Name(),
				MimeType:     getMimeType(filePath),
				OriginalSize: fileInfo.Size(),
				Description:  description,
				ChunkSize:    int32(chunkSize),
				TotalChunks:  int32(totalChunks),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("Возникла ошибка: %s\n", err.Error())
	}

	// 2. Разбиваем на чанки и отправляем в многопоточке
	var wg sync.WaitGroup
	chunks := make(chan *binarydata.UploadFileRequest, numberOfChunks)
	errors := make(chan error, totalChunks)

	// Запускаем workers для отправки
	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go sendChunkWorker(stream, chunks, errors, &wg)
	}

	// Создаем чанки
	for i := 0; i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(fileData) {
			end = len(fileData)
		}

		chunks <- &binarydata.UploadFileRequest{
			Data: &binarydata.UploadFileRequest_Chunk{
				Chunk: &binarydata.FileChunk{
					Data:       fileData[start:end],
					ChunkIndex: int32(i),
					IsLast:     i == totalChunks-1,
				},
			},
		}
	}
	close(chunks)

	wg.Wait()

	// Проверяем ошибки
	select {
	case err := <-errors:
		return fmt.Errorf("Возникла ошибка: %s\n", err.Error())
	default:
	}

	// Завершаем загрузку
	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("Возникла ошибка: %s\n", err.Error())
	}

	fmt.Printf("✅ Файл загружен успешно!\n")
	fmt.Printf("   ID: %d\n", response.FileId)
	fmt.Printf("   Размер: %d байт\n", response.BytesReceived)
	fmt.Printf("   Количество частей: %d\n", totalChunks)
	return nil
}

// Определение MIME типа по расширению файла
func getMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".pdf":  "application/pdf",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".zip":  "application/zip",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	return "application/octet-stream" // default
}

func sendChunkWorker(stream binarydata.Service_UploadFileClient, chunks <-chan *binarydata.UploadFileRequest, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for chunk := range chunks {
		if err := stream.Send(chunk); err != nil {
			// Проверяем, это EOF или реальная ошибка?
			if err == io.EOF {
				// Сервер закрыл соединение - это нормально
				fmt.Println("Сервер закрыл соединение (EOF)")
				return
			}

			// Это реальная ошибка
			errors <- fmt.Errorf("Возникла ошибка: %w", err)
			return
		}
	}

	// Все чанки успешно отправлены
	fmt.Println("Все части успешно отправлены")
}

func downloadFileFromConsole(client binarydata.ServiceClient) error {
	// 1. Запрашиваем ID файла
	fmt.Print("Введите идентификатор файла для загрузки: ")
	var fileID int64
	_, err := fmt.Scanln(&fileID)
	if err != nil {
		return fmt.Errorf("Возникла ошибка: %s\n", err.Error())
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
	stream, err := client.DownloadFile(ctx, &binarydata.DownloadFileRequest{
		FileId: fileID,
	})
	if err != nil {
		return fmt.Errorf("❌ Возникла ошибка: %w", err)
	}

	// 4. Получаем метаданные
	firstResponse, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("❌ Возникла ошибка: %w", err)
	}

	metadata := firstResponse.GetMetadata()
	if metadata == nil {
		fmt.Println("Возникла ошибка(метаданные должны быть отправлены первыми)")
	}

	// 5. Создаем файл для записи
	filePath := filepath.Join(downloadDir, metadata.Filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("❌ Возникла ошибка: %w", err)
	}
	defer file.Close()

	fmt.Printf("Загрузка: %s (%d байт)\n", metadata.Filename, metadata.OriginalSize)

	// 6. Многопоточная обработка чанков
	chunks := make(chan *binarydata.FileChunk, numberOfChunks)
	errors := make(chan error, 1)
	var wg sync.WaitGroup

	// Запускаем workers для записи
	for i := 0; i < numberOfWorkers; i++ {
		wg.Add(1)
		go writeChunkWorker(file, chunks, errors, &wg)
	}

	// 7. Получаем и обрабатываем чанки
	var receivedBytes int64
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			close(chunks)
			fmt.Println("Возникла ошибка: %w", err)
		}

		if chunk := response.GetChunk(); chunk != nil {
			chunks <- chunk
			receivedBytes += int64(len(chunk.Data))
		}
	}

	close(chunks)
	wg.Wait()

	// Проверяем ошибки
	select {
	case err = <-errors:
		return fmt.Errorf("❌ Возникла ошибка: %w", err)
	default:
	}

	fmt.Printf("\n✅ Файл загружен успешно: %s\n", filePath)
	return nil
}

func writeChunkWorker(file *os.File, chunks <-chan *binarydata.FileChunk, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for chunk := range chunks {
		// Используем mutex для thread-safe записи в файл
		err := func() error {
			writeMutex.Lock()
			defer writeMutex.Unlock()

			// Ищем позицию для записи (если чанки могут приходить не по порядку)
			offset := chunk.ChunkIndex * int32(chunkSize) // Предполагаем фиксированный размер чанка
			_, err := file.Seek(int64(offset), io.SeekStart)
			if err != nil {
				return err
			}

			_, err = file.Write(chunk.Data)
			return err
		}()

		if err != nil {
			errors <- fmt.Errorf("\n Возникла ошибка %d: %w", chunk.ChunkIndex, err)
			return
		}
	}
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

	dialog.PressEnterToContinue()
	return nil
}

func listOfFilesData(client binarydata.ServiceClient) error {
	currentPage := int32(1)
	filter := ""

	for {
		dialog.ClearScreen()
		fmt.Println("=== УПРАВЛЕНИЕ ФАЙЛАМИ ===")
		fmt.Printf("Страница: %d | Фильтр: '%s'\n", currentPage, filter)
		fmt.Println("===============================")

		// Получение данных с сервера
		resp, err := client.ListFiles(items.CreateAuthContext(), &binarydata.ListFilesRequest{
			Page:   currentPage,
			Filter: filter,
		})
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
			_, err = fmt.Scanln(&newPage)
			if err != nil {
				return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
			}
			if newPage >= 1 && newPage <= resp.TotalPages {
				currentPage = newPage
			} else {
				fmt.Println("Неверный номер страницы")
				dialog.PressEnterToContinue()
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
			dialog.PressEnterToContinue()
		}
	}
}

func getFileInfo(client binarydata.ServiceClient) error {
	fmt.Print("Введите идентификатор файла для просмотра: ")
	var fileID int64
	_, err := fmt.Scanln(&fileID)
	if err != nil {
		return fmt.Errorf("❌ Ошибка считывания: %s\n", err)
	}

	ctx := items.CreateAuthContext()
	resp, err := client.GetFileInfo(ctx, &binarydata.GetFileInfoRequest{
		FileId: fileID,
	})
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

	dialog.PressEnterToContinue()
	return nil
}
