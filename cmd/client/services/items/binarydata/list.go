package binarydata

import (
	"context"
	"fmt"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
)

// ListFiles получение списка данных по файлам
func (s *Service) ListFiles(ctx context.Context, currentPage int32, filter string) (*binarydata.ListFilesResponse, error) {
	resp, err := s.client.ListFiles(ctx, &binarydata.ListFilesRequest{
		Page:   currentPage,
		Filter: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка получения данных: %v\n", err)
	}
	return resp, nil
}
