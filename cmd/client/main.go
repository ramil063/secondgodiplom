package main

import (
	"fmt"
	"log"

	_ "github.com/ramil063/secondgodiplom/cmd/client/handlers/auth"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog/auth"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog/menu"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog/profile"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/dialog/registration"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/grpc"
	"github.com/ramil063/secondgodiplom/cmd/client/handlers/queue"
	cookieContants "github.com/ramil063/secondgodiplom/internal/constants/cookie"
	"github.com/ramil063/secondgodiplom/internal/security/cookie"
)

const serverAddr = "localhost:3202"

func main() {
	clients, err := grpc.NewGRPCClients(serverAddr)
	if err != nil {
		fmt.Println("---- ОШИБКА ЗАПУСКА СЕРВИСОВ ----")
		log.Fatal(err)
	}
	accessToken, refreshToken, _, err := cookie.LoadTokens(cookieContants.FileToSaveCookie)
	if err != nil {
		fmt.Println("---- ОШИБКА ЗАГРУЗКИ ТОКЕНОВ ----\n", err.Error())
		fmt.Println("---- Попробуйте войти еще раз ----")
	}

	var session dialog.UserSession
	currentState := dialog.StateMainMenu
	session.AccessToken = accessToken
	session.RefreshToken = refreshToken

	// Создаем канал для обработки состояний
	stateChan := make(chan dialog.AppState, 1)
	stateChan <- currentState

	// Запускаем сервис отправки очереди
	queueSender := queue.NewSender(*clients)
	go queueSender.Start()
	defer queueSender.Stop()

	for {
		currentState = <-stateChan
		if currentState == dialog.StateExit {
			fmt.Println("До свидания!")
			return
		}

		var nextState dialog.AppState
		var newSession dialog.UserSession

		switch currentState {
		case dialog.StateMainMenu:
			nextState = menu.ShowMainMenu()
			dialog.ClearScreen()
		case dialog.StateRegistration:
			nextState = registration.Registration(clients.RegistrationClient)
		case dialog.StateLogin:
			nextState, newSession = auth.Login(clients.AuthClient)
			session = newSession
		case dialog.StateUserProfile:
			nextState = profile.UserProfile(session, clients)
		default:
			nextState = dialog.StateMainMenu
		}

		stateChan <- nextState
	}
}
