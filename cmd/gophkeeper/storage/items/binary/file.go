package binary

import (
	"context"

	"github.com/ramil063/secondgodiplom/cmd/gophkeeper/storage/models/items"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/items/binary"
	"github.com/ramil063/secondgodiplom/internal/storage/db/dml/repository"
)

type Filer interface {
	CreateFileRecord(ctx context.Context, userID int, metadata *binarydata.FileMetadata) (int64, error)
	SaveChunk(ctx context.Context, fileID int64, chunkIndex int32, encryptedData []byte, algorithm string, iv []byte) error
	MarkFileComplete(ctx context.Context, fileID int64, totalBytes int64) error
	GetFileInfo(ctx context.Context, fileID int64, userID int64) (*items.FileInfo, error)
	GetChunksInRange(ctx context.Context, fileID int64, start, end int32) ([]*items.ChunkData, error)
	DeleteFile(ctx context.Context, userID, fileID int64) error
	GetListFiles(ctx context.Context, userID int64, page int32, perPage int32, filter string) ([]*items.FileInfo, int32, error)
	GetTotalCount(ctx context.Context, query string, userID int64, filter string) (int32, error)
}

func NewStorage(rep repository.Repository) Filer {
	return &binary.Item{
		Repository: &rep,
	}
}
