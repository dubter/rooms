package redis

import (
	"app-websocket/internal/domain"
)

func mapToUsers(m map[string]string) []domain.User {
	var users []domain.User
	for key, value := range m {
		user := domain.User{
			ID:       key,
			Nickname: value,
		}

		users = append(users, user)
	}

	return users
}
