package user

// User описывает данные о пользователя
type User struct {
	ID           int    `json:"id,omitempty"`
	Login        string `json:"login"`                // Логин
	PasswordHash string `json:"password"`             // Пароль
	FirstName    string `json:"first_name,omitempty"` // Имя
	LastName     string `json:"last_name,omitempty"`  // Фамилия
}
