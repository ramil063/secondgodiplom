package binary

import (
	"context"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/config"
	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/items/binary"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
	"github.com/ramil063/secondgodiplom/internal/security/crypto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	binarydata.UnimplementedServiceServer

	storage      binary.Filer
	Encryptor    crypto.Encryptor
	Decryptor    crypto.Decryptor
	workersCount int
}

func NewServer(storage binary.Filer, encryptor crypto.Encryptor, decryptor crypto.Decryptor, config *config.ServerConfig) *Server {
	return &Server{
		storage:      storage,
		Encryptor:    encryptor,
		Decryptor:    decryptor,
		workersCount: config.WorkersCount,
	}
}

type chunkTask struct {
	fileID     int64
	chunk      *binarydata.FileChunk
	chunkIndex int32
}

type chunkResult struct {
	chunkIndex     int32
	bytesProcessed int64
	err            error
}

func (s *Server) UploadFile(stream binarydata.Service_UploadFileServer) error {
	ctx := stream.Context()
	// Безопасное извлечение userID из контекста
	userIDValue := ctx.Value("userID")
	if userIDValue == nil {
		return status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, ok := userIDValue.(int)
	if !ok {
		return status.Error(codes.Internal, "invalid user ID format")
	}

	var metadata *binarydata.FileMetadata
	var fileID int64

	// Создаем каналы для многопоточной обработки
	chunks := make(chan *chunkTask, 100) // Буферизированный канал
	results := make(chan *chunkResult, 100)
	errors := make(chan error, 10) // Канал для ошибок
	var wg sync.WaitGroup

	// 1. Запускаем workers для обработки чанков (многопоточность!)
	for i := 0; i < s.workersCount; i++ {
		wg.Add(1)
		go s.chunkProcessorWorker(ctx, chunks, results, &wg)
	}

	// 2. Запускаем worker для сохранения в БД (отдельный поток)
	var totalBytes int64
	var totalChunks int32
	go func() {
		for result := range results {
			if result.err != nil {
				errors <- result.err
				return
			}
			totalBytes += result.bytesProcessed
		}
	}()

	// 3. Читаем stream и распределяем задачи
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			close(chunks)
			return status.Error(codes.Internal, "failed to receive data")
		}

		switch data := request.Data.(type) {
		case *binarydata.UploadFileRequest_Metadata:
			metadata = data.Metadata

			// Создаем запись о файле
			fileID, err = s.storage.CreateFileRecord(ctx, userID, metadata)
			if err != nil {
				close(chunks)
				return status.Error(codes.Internal, "failed to create file record")
			}

		case *binarydata.UploadFileRequest_Chunk:
			if metadata == nil {
				close(chunks)
				return status.Error(codes.InvalidArgument, "metadata must be sent first")
			}

			// Отправляем чанк в канал для обработки
			chunks <- &chunkTask{
				fileID:     fileID,
				chunk:      data.Chunk,
				chunkIndex: data.Chunk.ChunkIndex,
			}
			totalChunks++
		}
	}

	// 4. Завершаем обработку
	close(chunks)  // Сигнализируем workers о завершении
	wg.Wait()      // Ждем завершения всех workers
	close(results) // Закрываем results после завершения workers

	// 5. Проверяем ошибки
	select {
	case err := <-errors:
		return status.Error(codes.Internal, fmt.Sprintf("processing failed: %v", err))
	default:
	}

	// 6. Валидация - все ли чанки получены?
	if totalChunks != metadata.TotalChunks {
		return status.Error(codes.Internal,
			fmt.Sprintf("missing chunks: received %d, expected %d",
				totalChunks, metadata.TotalChunks))
	}

	// 7. Обновляем статус файла
	err := s.storage.MarkFileComplete(ctx, fileID, totalBytes)
	if err != nil {
		return status.Error(codes.Internal, "failed to mark file complete")
	}

	return stream.SendAndClose(&binarydata.UploadFileResponse{
		FileId:        fileID,
		BytesReceived: totalBytes,
		Status:        "success",
	})
}

func (s *Server) chunkProcessorWorker(
	ctx context.Context,
	tasks <-chan *chunkTask,
	results chan<- *chunkResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for task := range tasks {
		// Шифруем чанк
		encryptedData, algorithm, iv, err := s.Encryptor.Encrypt(task.chunk.Data)
		if err != nil {
			results <- &chunkResult{err: fmt.Errorf("chunk %d encryption failed: %w", task.chunkIndex, err)}
			continue
		}

		// Сохраняем чанк в БД (каждый worker имеет свое соединение)
		err = s.storage.SaveChunk(ctx, task.fileID, task.chunkIndex, encryptedData, algorithm, iv)
		if err != nil {
			results <- &chunkResult{err: fmt.Errorf("chunk %d save failed: %w", task.chunkIndex, err)}
			continue
		}

		results <- &chunkResult{
			chunkIndex:     task.chunkIndex,
			bytesProcessed: int64(len(task.chunk.Data)),
		}
	}
}

func (s *Server) DownloadFile(req *binarydata.DownloadFileRequest, stream binarydata.Service_DownloadFileServer) error {
	ctx := stream.Context()
	userID := ctx.Value("userID").(int)

	// 1. Получаем метаданные файла
	fileInfo, err := s.storage.GetFileInfo(ctx, req.FileId, int64(userID))
	if err != nil {
		return status.Error(codes.NotFound, "file not found")
	}

	// 2. Отправляем метаданные
	if err = stream.Send(&binarydata.DownloadFileResponse{
		Data: &binarydata.DownloadFileResponse_Metadata{
			Metadata: &binarydata.FileMetadata{
				Filename:     fileInfo.Filename,
				MimeType:     fileInfo.MimeType,
				OriginalSize: fileInfo.OriginalSize,
				Description:  fileInfo.Description,
				ChunkSize:    fileInfo.ChunkSize,
				TotalChunks:  fileInfo.TotalChunks,
			},
		},
	}); err != nil {
		return err
	}

	// 3. Многопоточное получение чанков
	chunks := make(chan *binarydata.FileChunk, 10)
	errors := make(chan error, 1)
	var wg sync.WaitGroup

	// Запускаем workers для получения чанков

	ranges := calculateChunkRanges(fileInfo.TotalChunks, int32(s.workersCount))
	for _, r := range ranges {
		wg.Add(1)
		go s.downloadChunkWorker(ctx, req.FileId, r.start, r.end, chunks, errors, &wg)
	}

	// 5. Важно: закрываем канал chunks после завершения всех воркеров
	go func() {
		wg.Wait()     // Ждем завершения всех воркеров
		close(chunks) // Закрываем канал - это разблокирует цикл range
	}()

	// 4. Отправляем чанки клиенту
	for chunk := range chunks {
		if err := stream.Send(&binarydata.DownloadFileResponse{
			Data: &binarydata.DownloadFileResponse_Chunk{
				Chunk: chunk,
			},
		}); err != nil {
			return err
		}
	}

	// Проверяем ошибки
	select {
	case err := <-errors:
		return err
	default:
		return nil
	}
}

func (s *Server) downloadChunkWorker(
	ctx context.Context,
	fileID int64,
	startChunk, endChunk int32,
	chunks chan<- *binarydata.FileChunk,
	errors chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// Получаем все чанки диапазона одним запросом
	chunkDataList, err := s.storage.GetChunksInRange(ctx, fileID, startChunk, endChunk)
	if err != nil {
		errors <- fmt.Errorf("failed to get chunks %d-%d: %w", startChunk, endChunk-1, err)
		return
	}

	// Обрабатываем каждый чанк
	for _, chunkData := range chunkDataList {
		decryptedData, err := s.Decryptor.Decrypt(chunkData.EncryptedData, chunkData.IV)
		if err != nil {
			errors <- fmt.Errorf("decryption failed for chunk: %w", err)
			return
		}

		chunks <- &binarydata.FileChunk{
			Data:       decryptedData,
			ChunkIndex: chunkData.ChunkIndex, // Нужно добавить index в ChunkData
			IsLast:     chunkData.ChunkIndex == endChunk-1,
		}
	}
}

// Динамическое распределение чанков по воркерам
func calculateChunkRanges(totalChunks, workersCount int32) []struct{ start, end int32 } {
	var ranges []struct{ start, end int32 }

	baseChunks := totalChunks / workersCount
	remainder := totalChunks % workersCount

	start := int32(0)
	for i := int32(0); i < workersCount; i++ {
		end := start + baseChunks
		if i < remainder {
			end++
		}
		ranges = append(ranges, struct{ start, end int32 }{start, end})
		start = end
	}

	return ranges
}

func (s *Server) DeleteFile(ctx context.Context, req *binarydata.DeleteFileRequest) (*binarydata.DeleteFileResponse, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}

	err := s.storage.DeleteFile(ctx, int64(userID), req.FileId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}
	return &binarydata.DeleteFileResponse{
		Success: true,
	}, nil
}

func (s *Server) ListFiles(ctx context.Context, req *binarydata.ListFilesRequest) (*binarydata.ListFilesResponse, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}

	// Валидация параметров
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 10 // Значение по умолчанию
	}
	listFiles, totalCount, err := s.storage.GetListFiles(ctx, int64(userID), req.Page, req.PerPage, req.Filter)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list files")
	}

	// Конвертируем в proto-сообщения
	var pbFileListItems []*binarydata.FileListItem

	for i := range listFiles {
		p := listFiles[i]

		pbFileListItems = append(pbFileListItems, &binarydata.FileListItem{
			Id:          p.ID,
			Filename:    p.Filename,
			MimeType:    p.MimeType,
			Size:        p.OriginalSize,
			Description: p.Description,
			CreatedAt:   p.CreatedAt.String(),
		})
	}
	totalPages := int32(math.Ceil(float64(totalCount) / float64(req.PerPage)))

	return &binarydata.ListFilesResponse{
		Files:       pbFileListItems,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		CurrentPage: req.Page,
	}, nil
}

func (s *Server) GetFileInfo(ctx context.Context, req *binarydata.GetFileInfoRequest) (*binarydata.FileInfoItem, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid authentication")
	}

	// 1. Получаем метаданные файла
	fileInfo, err := s.storage.GetFileInfo(ctx, req.FileId, int64(userID))
	if err != nil {
		return nil, status.Error(codes.NotFound, "file not found")
	}

	var metaDataList []*binarydata.MetaData

	for _, val := range fileInfo.MetaDataItems {
		metaDataList = append(metaDataList, &binarydata.MetaData{
			Id:    val.ID,
			Name:  val.Name,
			Value: val.Value,
		})
	}

	return &binarydata.FileInfoItem{
		Id:          fileInfo.ID,
		Filename:    fileInfo.Filename,
		MimeType:    fileInfo.MimeType,
		Size:        fileInfo.OriginalSize,
		Description: fileInfo.Description,
		CreatedAt:   fileInfo.CreatedAt.String(),
		MetaData:    metaDataList,
	}, nil
}
