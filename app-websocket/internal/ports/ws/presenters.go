package ws

import "time"

type Message struct {
	Content     string    `json:"content"`
	RoomID      string    `json:"room_id"`
	Username    string    `json:"nickname"`
	UserID      string    `json:"user_id"`
	TimeCreated time.Time `json:"time_created"`
}
