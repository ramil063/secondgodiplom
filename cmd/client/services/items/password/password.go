package password

import (
	"context"
	"fmt"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/password"
)

// Servicer интерфейс по работе с данными паролей
type Servicer interface {
	GetPassword(ctx context.Context, id int64) (*password.PasswordItem, error)
	ListPasswords(ctx context.Context, page int32, filter string) (*password.ListPasswordsResponse, error)
}

// Service сервис по работе с данными паролей
type Service struct {
	client password.ServiceClient
}

// NewService инициализация сервиса по работе с данными паролей
// в сервисе находится gRPC клиент для отправки данных на сервер
func NewService(client password.ServiceClient) *Service {
	return &Service{
		client: client,
	}
}

// GetPassword получение данных пароля
func (s *Service) GetPassword(ctx context.Context, id int64) (*password.PasswordItem, error) {
	resp, err := s.client.GetPassword(ctx, &password.GetPasswordRequest{
		Id: id,
	})

	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка получения данных\n")
	}
	return resp, nil
}

// ListPasswords получение списка данных паролей
func (s *Service) ListPasswords(ctx context.Context, page int32, filter string) (*password.ListPasswordsResponse, error) {
	// Получение данных с сервера
	resp, err := s.client.ListPasswords(ctx, &password.ListPasswordsRequest{
		Page:   page,
		Filter: filter,
	})
	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка получения данных: %v\n", err)
	}
	return resp, nil
}
