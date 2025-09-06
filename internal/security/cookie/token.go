package cookie

import (
	"encoding/json"
	"os"
)

type TokenStorage struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func SaveTokens(accessToken, refreshToken string, expiresIn int64) error {
	tokens := TokenStorage{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}
	data, _ := json.MarshalIndent(tokens, "", "  ")
	return os.WriteFile(".tokens.json", data, 0644)
}

func LoadTokens() (string, string, int64, error) {
	data, err := os.ReadFile(".tokens.json")
	if err != nil {
		return "", "", 0, err
	}

	var tokens TokenStorage
	if err := json.Unmarshal(data, &tokens); err != nil {
		return "", "", 0, err
	}

	return tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresIn, nil
}
