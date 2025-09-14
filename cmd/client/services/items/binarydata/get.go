package binarydata

import (
	"context"
	"fmt"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
)

// GetFileInfo получение информации по файлу
func (s *Service) GetFileInfo(ctx context.Context, fileID int64) (*binarydata.FileInfoItem, error) {
	resp, err := s.client.GetFileInfo(ctx, &binarydata.GetFileInfoRequest{
		FileId: fileID,
	})
	if err != nil {
		return nil, fmt.Errorf("\nОшибка получения данных по файлу: %s\n", err.Error())
	}
	return resp, err
}
