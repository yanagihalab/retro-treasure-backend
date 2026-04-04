package model

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token  string `json:"token"`
	UserID int64  `json:"user_id"`
}

type ExploreRequest struct {
	AreaID int64 `json:"area_id"`
}
