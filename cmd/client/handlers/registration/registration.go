package registration

import (
	"context"
	"fmt"
	"log"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
)

// RegisterUser регистрации пользователя
func RegisterUser(client auth.RegistrationServiceClient, login, password, firstName, lastName string) {
	resp, err := client.Register(context.Background(), &auth.RegisterRequest{
		Login:     login,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	})

	if err != nil {
		log.Fatal("Registration failed:", err)
	}

	if resp.UserId == "" {
		log.Fatal("Registration failed: empty user id")
	}

	fmt.Printf("User %s created successfully!\n", resp.UserId)
}
