package registration

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/registration"
	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
)

type UserData struct {
	Login     string
	Password  string
	FirstName string
	LastName  string
}

func collectRegistrationData() UserData {
	reader := bufio.NewReader(os.Stdin)
	userData := UserData{}

	// Сбор данных с валидацией
	for {
		fmt.Print("Введите логин: ")
		login, _ := reader.ReadString('\n')
		login = strings.TrimSpace(login)

		if len(login) <= 1 {
			fmt.Println("❌ Логин должен содержать минимум 2 символа")
			continue
		}
		userData.Login = login
		break
	}

	for {
		fmt.Print("Введите пароль: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		if len(password) <= 5 {
			fmt.Println("❌ Пароль должен содержать минимум 6 символов")
			continue
		}
		userData.Password = password
		break
	}

	for {
		fmt.Print("Введите подтверждение пароля: ")
		passwordConfirm, _ := reader.ReadString('\n')
		passwordConfirm = strings.TrimSpace(passwordConfirm)

		if passwordConfirm != userData.Password {
			fmt.Println("❌ Неправильно задано подтверждение пароля")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите имя: ")
		firstName, _ := reader.ReadString('\n')
		firstName = strings.TrimSpace(firstName)

		if len(firstName) <= 1 {
			fmt.Println("❌ Имя должно содержать минимум 2 символа")
			continue
		}
		userData.FirstName = firstName
		break
	}

	for {
		fmt.Print("Введите фамилию: ")
		lastName, _ := reader.ReadString('\n')
		lastName = strings.TrimSpace(lastName)

		if len(lastName) <= 2 {
			fmt.Println("❌ Фамилия должна содержать минимум 2 символа")
			continue
		}
		userData.LastName = lastName
		break
	}

	return userData
}

func Registration(client auth.RegistrationServiceClient) dialog.AppState {
	fmt.Println("\n=== РЕГИСТРАЦИЯ ===")

	userData := collectRegistrationData()

	// Подтверждение
	fmt.Println("\n--- Подтверждение ---")
	fmt.Printf("Логин: %s\n", userData.Login)
	fmt.Printf("Имя: %s\n", userData.FirstName)
	fmt.Printf("Фамилия: %s\n", userData.LastName)

	if confirmData() {
		registration.RegisterUser(
			client,
			userData.Login,
			userData.Password,
			userData.FirstName,
			userData.LastName)
	} else {
		fmt.Println("❌ Регистрация отменена")
	}
	return dialog.StateMainMenu
}

func confirmData() bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Подтвердить регистрацию? (y/n): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer == "y" || answer == "yes" {
			return true
		} else if answer == "n" || answer == "no" {
			return false
		}
		fmt.Println("❌ Пожалуйста, введите 'y' или 'n'")
	}
}
