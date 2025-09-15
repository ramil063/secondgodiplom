package textdata

import (
	"context"
	"fmt"

	"github.com/ramil063/secondgodiplom/cmd/client/generics/list"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/textdata"
)

// Servicer интерфейс по работе с текстовыми данными
type Servicer interface {
	GetTextData(ctx context.Context, id int64) (*textdata.TextDataItem, error)
	ListItems(ctx context.Context, page int32, filter string) (*list.Response[textdata.TextDataItem], error)
}

// Service сервис по работе с текстовыми данными
type Service struct {
	client textdata.ServiceClient
}

// NewService инициализация сервиса по работе с текстовыми данными
// в сервисе находится gRPC клиент для отправки данных на сервер
func NewService(client textdata.ServiceClient) *Service {
	return &Service{
		client: client,
	}
}

// GetTextData получение текстовых данных
func (s *Service) GetTextData(ctx context.Context, id int64) (*textdata.TextDataItem, error) {
	resp, err := s.client.GetTextData(ctx, &textdata.GetTextDataRequest{
		Id: id,
	})
	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка получения данных\n")
	}
	return resp, nil
}

// ListItems получение списка текстовых данных
func (s *Service) ListItems(ctx context.Context, page int32, filter string) (*list.Response[textdata.TextDataItem], error) {
	resp, err := s.client.ListTextDataItems(ctx, &textdata.ListTextDataRequest{
		Page:   page,
		Filter: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка получения данных: %v\n", err)
	}
	return &list.Response[textdata.TextDataItem]{
		Items:       resp.TextDataItems,
		TotalPages:  resp.TotalPages,
		TotalCount:  resp.TotalCount,
		CurrentPage: resp.CurrentPage,
	}, nil
}
