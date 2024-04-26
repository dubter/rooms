package chat

import "time"

type Message struct {
	Content     string    `json:"content"`
	RoomID      string    `json:"roomId"`
	Username    string    `json:"username"`
	UserID      string    `json:"user_id"`
	TimeCreated time.Time `json:"time_created"`
}

type RoomResp struct {
	ID          string    `json:"id"`
	TimeCreated time.Time `json:"time_created"`
	Name        string    `json:"name"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthReq struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}
