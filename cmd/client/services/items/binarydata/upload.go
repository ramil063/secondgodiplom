package binarydata

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
)

// UploadData загрузка файла на сервер
func (s *Service) UploadData(
	ctx context.Context,
	fileData []byte,
	fileInfo os.FileInfo,
	filePath,
	description string,
) (*binarydata.UploadFileResponse, int, error) {

	stream, err := s.client.UploadFile(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("❌ Возникла ошибка: %s\n", err.Error())
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
		return nil, 0, fmt.Errorf("❌ Возникла ошибка: %s\n", err.Error())
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
		return nil, 0, fmt.Errorf("❌ Возникла ошибка: %s\n", err.Error())
	default:
	}

	// Завершаем загрузку
	response, err := stream.CloseAndRecv()
	if err != nil {
		return nil, 0, err
	}
	return response, totalChunks, nil
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
			errors <- fmt.Errorf("❌ Возникла ошибка: %w", err)
			return
		}
	}

	// Все чанки успешно отправлены
	fmt.Println("Все части успешно отправлены")
}
