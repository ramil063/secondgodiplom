package registration

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/services/registration"
)

// UserData данные необходимые пользователю для регистрации
type UserData struct {
	Login     string
	Password  string
	FirstName string
	LastName  string
}

func collectRegistrationData() (UserData, error) {
	reader := bufio.NewReader(os.Stdin)
	userData := UserData{}

	// Сбор данных с валидацией
	for {
		fmt.Print("Введите логин: ")
		login, err := reader.ReadString('\n')
		if err != nil {
			return userData, fmt.Errorf("❌ Ошибка считывания: %s\n", err)
		}
		login = strings.TrimSpace(login)

		if len(login) < 3 {
			fmt.Println("❌ Логин должен содержать минимум 3 символа")
			continue
		}
		userData.Login = login
		break
	}

	for {
		fmt.Print("Введите пароль: ")
		password, err := reader.ReadString('\n')
		if err != nil {
			return userData, fmt.Errorf("❌ Ошибка считывания: %s\n", err)
		}
		password = strings.TrimSpace(password)

		if len(password) < 6 {
			fmt.Println("❌ Пароль должен содержать минимум 6 символов")
			continue
		}
		userData.Password = password
		break
	}

	for {
		fmt.Print("Введите подтверждение пароля: ")
		passwordConfirm, err := reader.ReadString('\n')
		if err != nil {
			return userData, fmt.Errorf("❌ Ошибка считывания: %s\n", err)
		}
		passwordConfirm = strings.TrimSpace(passwordConfirm)

		if passwordConfirm != userData.Password {
			fmt.Println("❌ Неправильно задано подтверждение пароля")
			continue
		}
		break
	}

	for {
		fmt.Print("Введите имя: ")
		firstName, err := reader.ReadString('\n')
		if err != nil {
			return userData, fmt.Errorf("❌ Ошибка считывания: %s\n", err)
		}
		firstName = strings.TrimSpace(firstName)

		if len(firstName) < 3 {
			fmt.Println("❌ Имя должно содержать минимум 3 символа")
			continue
		}
		userData.FirstName = firstName
		break
	}

	for {
		fmt.Print("Введите фамилию: ")
		lastName, err := reader.ReadString('\n')
		if err != nil {
			return userData, fmt.Errorf("❌ Ошибка считывания: %s\n", err)
		}
		lastName = strings.TrimSpace(lastName)

		if len(lastName) < 3 {
			fmt.Println("❌ Фамилия должна содержать минимум 3 символа")
			continue
		}
		userData.LastName = lastName
		break
	}

	return userData, nil
}

// Registration функция для отображения интерфейса регистрации пользователя
func Registration(service registration.Servicer) dialog.AppState {
	fmt.Println("\n=== РЕГИСТРАЦИЯ ===")

	userData, err := collectRegistrationData()
	if err != nil {
		fmt.Println("❌ Возникла ошибка при регистрации пользователя:", err)
		return dialog.StateMainMenu
	}

	// Подтверждение
	fmt.Println("\n--- Подтверждение ---")
	fmt.Printf("Логин: %s\n", userData.Login)
	fmt.Printf("Имя: %s\n", userData.FirstName)
	fmt.Printf("Фамилия: %s\n", userData.LastName)

	if confirmData() {
		resp, err := service.RegisterUser(
			userData.Login,
			userData.Password,
			userData.FirstName,
			userData.LastName)
		if err != nil {
			fmt.Println("❌ Возникла ошибка при регистрации пользователя:", err)
			return dialog.StateMainMenu
		}
		fmt.Printf("Пользователь %s зарегистрирован успешно!\n", resp.UserId)
	} else {
		fmt.Println("❌ Регистрация отменена")
	}
	return dialog.StateMainMenu
}

func confirmData() bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Подтвердить регистрацию? (y/n): ")
		answer, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ Ошибка считывания: %s\n", err)
			return false
		}
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer == "y" || answer == "yes" {
			return true
		} else if answer == "n" || answer == "no" {
			return false
		}
		fmt.Println("❌ Пожалуйста, введите 'y' или 'n'")
	}
}
