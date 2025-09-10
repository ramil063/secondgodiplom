package profile

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog/profile/items"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/grpc"
)

// UserProfile функция работы с главным меню профиля пользователя
func UserProfile(session dialog.UserSession, clients *grpc.Clients) dialog.AppState {
	if session.AccessToken == "" {
		dialog.ClearScreen()
		fmt.Println("❌ Пожалуйста авторизуйтесь!")
		return dialog.StateMainMenu // Выход в главное меню
	}
	for {
		dialog.ClearScreen()
		showUserProfileMenu()

		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			items.WorkWithPassword(clients.PasswordsClient)
		case "2":
			items.WorkWithTextData(clients.TextDataClient)
		case "3":
			items.WorkWithBankCardData(clients.BankCardDataClient)
		case "4":
			items.WorkWithFile(clients.BinaryDataClient)
		case "5":
			return dialog.StateMainMenu // Выход в главное меню
		case "6":
			return dialog.StateExit // Полный выход
		default:
			fmt.Println("❌ Неверный выбор!")
			dialog.PressEnterToContinue()
		}
	}
}

func showUserProfileMenu() {
	fmt.Printf("=== ЛИЧНЫЙ КАБИНЕТ ===\n")
	fmt.Println("========================")
	fmt.Println("1. Работа с паролями")
	fmt.Println("2. Работа с текстом")
	fmt.Println("3. Работа с банковскими картами")
	fmt.Println("4. Работа с файлами")
	fmt.Println("5. Выйти в главное меню")
	fmt.Println("6. Выйти из приложения")
	fmt.Println("========================")
	fmt.Print("Выберите действие: ")
}
