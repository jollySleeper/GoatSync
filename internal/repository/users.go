package repository

import (
	"goatsync/internal/models"
)

var users map[string]models.User

func GetUsers() map[string]models.User {
	if users == nil {
		users = make(map[string]models.User)
	}

	return users
}
