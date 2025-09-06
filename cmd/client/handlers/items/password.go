package items

import (
	"fmt"
	"log"

	"github.com/ramil063/secondgodiplom/internal/proto/gen/items/password"
)

func CreatePassword(client password.ServiceClient) {
	ctx := CreateAuthContext()

	resp, err := client.CreatePassword(ctx, &password.CreatePasswordRequest{
		Login:         "ramil5",
		Password:      "ramil5",
		Target:        "yandex5.ru",
		Description:   "password for Ramil5",
		MetaDataName:  "metadata name for password 1",
		MetaDataValue: "metadata value for password 1",
	})

	if err != nil {
		log.Fatal("CreatePassword failed:", err)
	}

	if resp.Id == 0 {
		log.Fatal("CreatePassword failed: empty user id")
	}

	fmt.Printf("Password %s created successfully!\n", resp.Id)
}

func ListPasswords(client password.ServiceClient, page, perPage int32, filter string) {
	ctx := CreateAuthContext()

	resp, err := client.ListPasswords(ctx, &password.ListPasswordsRequest{
		Page:    page,
		PerPage: perPage,
		Filter:  filter,
	})

	if err != nil {
		log.Fatal("Failed to list passwords:", err)
	}

	fmt.Printf("Total passwords: %d\n", resp.TotalCount)
	fmt.Printf("Page %d\n", resp.CurrentPage)
	fmt.Println("Passwords:")

	for i, p := range resp.Passwords {
		fmt.Printf("%d. %s\n", i+1, p.Id)
		fmt.Printf("   Login: %s\n", p.Login)
		fmt.Printf("   Password: %s\n", p.Password)
		fmt.Printf("   Target: %s\n", p.Target)
		fmt.Printf("   Created: %s\n", p.CreatedAt)
		fmt.Println("---")
	}
}

func GetPassword(client password.ServiceClient, id int64) {
	ctx := CreateAuthContext()

	resp, err := client.GetPassword(ctx, &password.GetPasswordRequest{
		Id: id,
	})

	if err != nil {
		log.Fatal("Failed to get password:", err)
	}

	fmt.Printf("Id: %d\n", resp.Id)
	fmt.Printf("   Login: %s\n", resp.Login)
	fmt.Printf("   Password: %s\n", resp.Password)
	fmt.Printf("   Target: %s\n", resp.Target)
	fmt.Printf("   Created: %s\n", resp.CreatedAt)
	fmt.Println("---")
}

func DeletePassword(client password.ServiceClient, id int64) {
	ctx := CreateAuthContext()

	_, err := client.DeletePassword(ctx, &password.DeletePasswordRequest{
		Id: id,
	})

	if err != nil {
		log.Fatal("Failed to delete password:", err)
	}

	fmt.Printf("Password with %d deleted successfully\n", id)
}

func UpdatePassword(client password.ServiceClient) {
	ctx := CreateAuthContext()

	resp, err := client.UpdatePassword(ctx, &password.UpdatePasswordRequest{
		Id:          int64(5),
		Login:       "ramil5",
		Password:    "ramil5",
		Target:      "yandex5.ru",
		Description: "password for Ramil5",
	})

	if err != nil {
		log.Fatal("UpdateItem failed:", err)
	}

	if resp.Id == 0 {
		log.Fatal("UpdateItem failed: empty user id")
	}

	fmt.Printf("Password %d updated successfully!\n", resp.Id)
}
