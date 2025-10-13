package repository

import "goatsync/internal/models"

// token -> username
var tokens map[string]models.Token

func GetTokens() map[string]models.Token {
	if tokens == nil {
		tokens = make(map[string]models.Token)
	}

	return tokens
}
