package binarydata

import (
	"context"
	"os"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
)

const chunkSize = 64 * 1024 // 64KB
const numberOfWorkers = 4
const numberOfChunks = 20

// Servicer интерфейс по работе с файлами
type Servicer interface {
	UploadData(ctx context.Context, fileData []byte, fileInfo os.FileInfo, filePath, description string) (*binarydata.UploadFileResponse, int, error)
	DownloadData(ctx context.Context, fileID int64, downloadDir string) (string, error)
	ListFiles(ctx context.Context, currentPage int32, filter string) (*binarydata.ListFilesResponse, error)
	GetFileInfo(ctx context.Context, fileID int64) (*binarydata.FileInfoItem, error)
}

// Service сервис по работе с файлами
type Service struct {
	client binarydata.ServiceClient
}

// NewService инициализация сервиса по работе с файлами
// в сервисе находится gRPC клиент для отправки данных на сервер
func NewService(client binarydata.ServiceClient) *Service {
	return &Service{
		client: client,
	}
}
