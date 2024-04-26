package auth

type registerRequest struct {
	Nickname string `json:"nickname" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8,max=50"`
}

type loginRequest struct {
	Nickname string `json:"nickname" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8,max=50"`
}

type tokenResponse struct {
	Nickname     string `json:"nickname"`
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
