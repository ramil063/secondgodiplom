package menu

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
)

func ShowMainMenu() dialog.AppState {
	dialog.ClearScreen()
	fmt.Println("Добро пожаловать!")
	fmt.Println("====================")
	fmt.Println("1. Регистрация")
	fmt.Println("2. Авторизация")
	fmt.Println("3. Профиль")
	fmt.Println("4. Выход")
	fmt.Println("====================")
	fmt.Print("Выберите действие: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return dialog.StateRegistration
	case "2":
		return dialog.StateLogin
	case "3":
		return dialog.StateUserProfile
	case "4":
		return dialog.StateExit
	default:
		fmt.Println("❌ Неверный выбор! Попробуйте снова.")
		time.Sleep(1 * time.Second)
		return dialog.StateMainMenu
	}
}
