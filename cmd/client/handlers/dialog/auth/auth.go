package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	authService "github.com/ramil063/secondgodiplom/cmd/client/services/auth"
)

// Login основная функция авторизации пользователя
func Login(client authService.Servicer) (dialog.AppState, dialog.UserSession) {
	err := dialog.ClearScreen()
	if err != nil {
		fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
		return dialog.StateExit, dialog.UserSession{}
	}
	fmt.Println("\n=== АВТОРИЗАЦИЯ ===")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Логин: ")
	login, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ Ошибка считывания логина: %v\n", err)
		err = dialog.PressEnterToContinue()
		if err != nil {
			fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
		}
		return dialog.StateExit, dialog.UserSession{}
	}
	login = strings.TrimSpace(login)

	fmt.Print("Пароль: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ Ошибка считывания пароля: %v\n", err)
		err = dialog.PressEnterToContinue()
		if err != nil {
			fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
		}
		return dialog.StateExit, dialog.UserSession{}
	}
	password = strings.TrimSpace(password)

	// Отправка запроса авторизации
	session, err := client.LoginProcess(login, password)
	if err != nil {
		fmt.Printf("❌ Ошибка авторизации: %v\n", err)
		err = dialog.PressEnterToContinue()
		if err != nil {
			fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
		}
		return dialog.StateMainMenu, dialog.UserSession{}
	}
	fmt.Printf("✅ Авторизация успешна! Добро пожаловать!\n")

	return dialog.StateUserProfile, session
}
