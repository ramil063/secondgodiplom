package bankcard

import (
	"context"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/bankcard"
)

// Servicer интерфейс по работе с данными банковских карт
type Servicer interface {
	GetCardData(ctx context.Context, id int64) (bankcard.CardDataItem, error)
	ListCardsData(ctx context.Context, page int32, filter string) (*bankcard.ListCardsDataResponse, error)
}

// Service сервис по работе с данными банковских карт
type Service struct {
	client bankcard.ServiceClient
}

// NewService инициализация сервиса по работе с данными банковских карт
// в сервисе находится gRPC клиент для отправки данных на сервер
func NewService(client bankcard.ServiceClient) *Service {
	return &Service{
		client: client,
	}
}

// GetCardData получение данных банковской карты
func (s *Service) GetCardData(ctx context.Context, id int64) (bankcard.CardDataItem, error) {
	resp, err := s.client.GetCardData(ctx, &bankcard.GetCardDataRequest{
		Id: id,
	})
	if err != nil {
		return bankcard.CardDataItem{}, err
	}
	return bankcard.CardDataItem{
		Id:              resp.Id,
		Number:          resp.Number,
		ValidUntilYear:  resp.ValidUntilYear,
		ValidUntilMonth: resp.ValidUntilMonth,
		Cvv:             resp.Cvv,
		Holder:          resp.Holder,
		Description:     resp.Description,
		CreatedAt:       resp.CreatedAt,
		MetaData:        resp.MetaData,
	}, nil
}

// ListCardsData получение списка данных банковских карт
func (s *Service) ListCardsData(ctx context.Context, page int32, filter string) (*bankcard.ListCardsDataResponse, error) {
	resp, err := s.client.ListCardsData(ctx, &bankcard.ListCardsDataRequest{
		Page:   page,
		Filter: filter,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
