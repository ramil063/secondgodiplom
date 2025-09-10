package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	auth2 "github.com/ramil063/secondgodiplom/cmd/client/handlers/auth"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
)

// Login основная функция авторизации пользователя
func Login(client auth.AuthServiceClient) (dialog.AppState, dialog.UserSession) {
	dialog.ClearScreen()
	fmt.Println("\n=== АВТОРИЗАЦИЯ ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Логин: ")
	login, _ := reader.ReadString('\n')
	login = strings.TrimSpace(login)

	fmt.Print("Пароль: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Отправка запроса авторизации
	session, err := auth2.Login(client, login, password)
	if err != nil {
		fmt.Printf("❌ Ошибка авторизации: %v\n", err)
		dialog.PressEnterToContinue()
		return dialog.StateMainMenu, dialog.UserSession{}
	}
	fmt.Printf("✅ Авторизация успешна! Добро пожаловать!\n")

	return dialog.StateUserProfile, session
}
