package dialog

// AppState тип для хранения состояния приложения
type AppState int

const (
	StateMainMenu AppState = iota
	StateRegistration
	StateLogin
	StateUserProfile
	StateExit
)

// UserSession пользовательская сессия, используется для пропуска к работе только авторизованного пользователя
type UserSession struct {
	AccessToken  string
	RefreshToken string
	IsLoggedIn   bool
}
