package registration

import (
	"context"
	"fmt"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
)

// Servicer интерфейс описывающий методы работы с регистрацией пользователя
type Servicer interface {
	RegisterUser(login, password, firstName, lastName string) (*auth.RegisterResponse, error)
}

// Service сервис по работе с регистрацией пользователя
type Service struct {
	client auth.RegistrationServiceClient
}

// NewService инициализация сервиса по работе с регистрацией пользователя
// в сервисе находится gRPC клиент для отправки данных на сервер
func NewService(client auth.RegistrationServiceClient) *Service {
	return &Service{
		client: client,
	}
}

// RegisterUser функция регистрации пользователя
func (s *Service) RegisterUser(login, password, firstName, lastName string) (*auth.RegisterResponse, error) {
	resp, err := s.client.Register(context.Background(), &auth.RegisterRequest{
		Login:     login,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	})

	if err != nil {
		return nil, fmt.Errorf("❌ Ошибка регистрации: %w\n", err)
	}

	if resp.UserId == "" {
		return nil, fmt.Errorf("❌ Ошибка регистрации: пустой идентификатор пользователя")
	}

	return resp, nil
}
