package auth

import (
	"context"
	"fmt"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	cookieContants "github.com/ramil063/secondgodiplom/internal/constants/cookie"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
	"github.com/ramil063/secondgodiplom/internal/security/cookie"
)

// Servicer интерфейс для работы с авторизацией
type Servicer interface {
	LoginProcess(login, password string) (dialog.UserSession, error)
	RefreshProcess() error
}

// Service сервис по работе с аторизацией
type Service struct {
	client auth.AuthServiceClient
}

// NewService инициализация сервиса по работе с авторизацией
// сервис является композицией с клиентом авторизации
func NewService(client auth.AuthServiceClient) *Service {
	return &Service{
		client: client,
	}
}

// LoginProcess функция для авторизации и сохранения токенов авторизации
// токены сохраняются в специальный файл
func (s *Service) LoginProcess(login, password string) (dialog.UserSession, error) {
	resp, err := s.client.Login(context.Background(), &auth.LoginRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		fmt.Println("❌ Ошибка авторизации:", err)
		return dialog.UserSession{}, err
	}

	// Сохраняем токены (например, в файл)
	err = cookie.SaveTokens(resp.AccessToken, resp.RefreshToken, cookieContants.FileToSaveCookie, resp.ExpiresIn)
	if err != nil {
		fmt.Println("❌ Ошибка сохранения токенов:", err)
		return dialog.UserSession{}, err
	}

	return dialog.UserSession{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		IsLoggedIn:   true,
	}, nil
}

// RefreshProcess функция для обновления токенов при истечении срока жизни токена авторизации
func (s *Service) RefreshProcess() error {
	_, refreshToken, _, err := cookie.LoadTokens(cookieContants.FileToSaveCookie)
	if err != nil {
		fmt.Println("❌ Ошибка загрузки токенов авторизации:", err)
		return err
	}
	resp, err := s.client.Refresh(context.Background(), &auth.RefreshRequest{
		RefreshToken: refreshToken,
	})

	if err != nil {
		fmt.Println("❌ Ошибка обновления токенов авторизации:", err)
		return err
	}

	// Сохраняем токены (например, в файл)
	err = cookie.SaveTokens(resp.AccessToken, resp.RefreshToken, cookieContants.FileToSaveCookie, resp.ExpiresIn)
	if err != nil {
		fmt.Println("❌ Ошибка сохранения токенов:", err)
		return err
	}
	fmt.Println("Обновление токенов авторизации прошло успешно. Токены сохранены.")
	return nil
}
