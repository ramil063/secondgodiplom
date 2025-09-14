package profile

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog/profile/items"
	"github.com/ramil063/secondgodiplom/cmd/client/services/items/bankcard"
	"github.com/ramil063/secondgodiplom/cmd/client/services/items/binarydata"
	"github.com/ramil063/secondgodiplom/cmd/client/services/items/password"
	"github.com/ramil063/secondgodiplom/cmd/client/services/items/textdata"
)

// UserProfile функция работы с главным меню профиля пользователя
func UserProfile(
	session dialog.UserSession,
	bcServ bankcard.Servicer,
	bServ binarydata.Servicer,
	passwordServ password.Servicer,
	textdataServ textdata.Servicer,
) dialog.AppState {
	if session.AccessToken == "" {
		err := dialog.ClearScreen()
		if err != nil {
			fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
		}
		fmt.Println("❌ Пожалуйста авторизуйтесь!")
		return dialog.StateMainMenu // Выход в главное меню
	}
	for {
		err := dialog.ClearScreen()
		if err != nil {
			fmt.Printf("❌ Ошибка очистки экрана: %v\n", err)
		}
		showUserProfileMenu()

		reader := bufio.NewReader(os.Stdin)
		choice, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Ошибка считывания выбора: %v\n", err)
			return dialog.StateMainMenu // Выход в главное меню
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			items.WorkWithPassword(passwordServ)
		case "2":
			items.WorkWithTextData(textdataServ)
		case "3":
			items.WorkWithBankCardData(bcServ)
		case "4":
			items.WorkWithFile(bServ)
		case "5":
			return dialog.StateMainMenu // Выход в главное меню
		case "6":
			return dialog.StateExit // Полный выход
		default:
			fmt.Println("❌ Неверный выбор!")
			err = dialog.PressEnterToContinue()
			if err != nil {
				fmt.Printf("❌ Ошибка при нажатии на Enter: %v\n", err)
			}
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
