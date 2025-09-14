package textdata

import (
	"context"
	"fmt"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/textdata"
)

// Servicer интерфейс по работе с текстовыми данными
type Servicer interface {
	GetTextData(ctx context.Context, id int64) (*textdata.TextDataItem, error)
	ListTextDataItems(ctx context.Context, page int32, filter string) (*textdata.ListTextDataResponse, error)
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

// ListTextDataItems получение списка текстовых данных
func (s *Service) ListTextDataItems(ctx context.Context, page int32, filter string) (*textdata.ListTextDataResponse, error) {
	resp, err := s.client.ListTextDataItems(ctx, &textdata.ListTextDataRequest{
		Page:   page,
		Filter: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка получения данных: %v\n", err)
	}
	return resp, nil
}
