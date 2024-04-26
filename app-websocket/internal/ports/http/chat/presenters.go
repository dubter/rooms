package chat

import "time"

type CreateRoomReq struct {
	Name string `json:"name"`
}

type RoomRes struct {
	ID          string    `json:"id"`
	TimeCreated time.Time `json:"time_created"`
	Name        string    `json:"name"`
}

type ClientRes struct {
	Username string `json:"nickname"`
}
