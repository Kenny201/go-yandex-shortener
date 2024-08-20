package entity

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"id,omitempty"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Username string    `json:"username,omitempty"`
}

func NewUser(username, email, password string) *User {
	id := uuid.New()

	return &User{
		ID:       id,
		Username: username,
		Email:    email,
		Password: password,
	}
}
