package binarydata

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
)

var writeMutex sync.Mutex

// DownloadData скачивание файла с сервера
func (s *Service) DownloadData(ctx context.Context, fileID int64, downloadDir string) (string, error) {
	stream, err := s.client.DownloadFile(ctx, &binarydata.DownloadFileRequest{
		FileId: fileID,
	})
	if err != nil {
		return "", fmt.Errorf("❌ Возникла ошибка: %w", err)
	}

	// 4. Получаем метаданные
	firstResponse, err := stream.Recv()
	if err != nil {
		return "", fmt.Errorf("❌ Возникла ошибка: %w", err)
	}

	metadata := firstResponse.GetMetadata()
	if metadata == nil {
		return "", fmt.Errorf("❌ Возникла ошибка(метаданные должны быть отправлены первыми)")
	}

	// 5. Создаем файл для записи
	filePath := filepath.Join(downloadDir, metadata.Filename)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("❌ Возникла ошибка: %w", err)
	}
	defer file.Close()

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
		return "", fmt.Errorf("❌ Возникла ошибка: %w", err)
	default:
	}
	return filePath, nil
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
