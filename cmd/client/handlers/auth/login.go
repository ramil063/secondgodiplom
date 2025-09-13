package auth

import (
	"context"
	"fmt"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	cookieContants "github.com/ramil063/secondgodiplom/internal/constants/cookie"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
	"github.com/ramil063/secondgodiplom/internal/security/cookie"
)

// Login функция для авторизации и сохранения токенов авторизации
// токены сохраняются в специальный файл
func Login(client auth.AuthServiceClient, login, password string) (dialog.UserSession, error) {
	resp, err := client.Login(context.Background(), &auth.LoginRequest{
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

// Refresh функция для обновления токенов при истечении срока жизни токена авторизации
func Refresh(client auth.AuthServiceClient) error {
	_, refreshToken, _, err := cookie.LoadTokens(cookieContants.FileToSaveCookie)
	if err != nil {
		fmt.Println("Load tokens failed:", err)
		return err
	}
	resp, err := client.Refresh(context.Background(), &auth.RefreshRequest{
		RefreshToken: refreshToken,
	})

	if err != nil {
		fmt.Println("Refresh failed:", err)
		return err
	}

	// Сохраняем токены (например, в файл)
	err = cookie.SaveTokens(resp.AccessToken, resp.RefreshToken, cookieContants.FileToSaveCookie, resp.ExpiresIn)
	if err != nil {
		fmt.Println("Cookie save failed:", err)
		return err
	}
	fmt.Println("Refresh successful! Tokens saved.")
	return nil
}
