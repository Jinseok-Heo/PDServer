package model

type User struct {
	ID       uint64 `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Birth    string `json:"birth"`
}
