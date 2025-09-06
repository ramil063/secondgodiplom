package dialog

type AppState int

const (
	StateMainMenu AppState = iota
	StateRegistration
	StateLogin
	StateUserProfile
	StateExit
)

type UserSession struct {
	AccessToken  string
	RefreshToken string
	IsLoggedIn   bool
}
