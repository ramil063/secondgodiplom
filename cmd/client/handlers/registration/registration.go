package registration

import (
	"context"
	"errors"
	"fmt"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/auth"
)

// RegisterUser регистрации пользователя
func RegisterUser(client auth.RegistrationServiceClient, login, password, firstName, lastName string) error {
	resp, err := client.Register(context.Background(), &auth.RegisterRequest{
		Login:     login,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	})

	if err != nil {
		fmt.Println("Registration failed:", err)
		return err
	}

	if resp.UserId == "" {
		fmt.Println("Registration failed: empty user id")
		return errors.New("Registration failed: empty user id")
	}

	fmt.Printf("User %s created successfully!\n", resp.UserId)
	return nil
}
