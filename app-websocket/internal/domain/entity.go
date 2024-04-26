package domain

import "time"

type Message struct {
	Content     string
	Nickname    string
	TimeCreated time.Time
	RoomID      string
	UserID      string
}

type User struct {
	ID           string
	Nickname     string
	PasswordHash string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type Session struct {
	RefreshToken string
	ExpiresAt    time.Time
}

type Room struct {
	ID          string
	Name        string
	TimeCreated time.Time
}

type MessageHandler func(msg Message) error
