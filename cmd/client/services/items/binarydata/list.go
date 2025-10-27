package binarydata

import (
	"context"
	"fmt"

	"github.com/ramil063/secondgodiplom/cmd/client/generics/list"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/binarydata"
)

// ListItems получение списка данных по файлам
func (s *Service) ListItems(ctx context.Context, page int32, filter string) (*list.Response[binarydata.FileListItem], error) {
	resp, err := s.client.ListFiles(ctx, &binarydata.ListFilesRequest{
		Page:   page,
		Filter: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка получения данных: %v\n", err)
	}

	return &list.Response[binarydata.FileListItem]{
		Items:       resp.Files,
		TotalPages:  resp.TotalPages,
		TotalCount:  resp.TotalCount,
		CurrentPage: resp.CurrentPage,
	}, nil
}
