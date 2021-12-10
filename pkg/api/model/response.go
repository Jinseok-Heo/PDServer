package model

type LoginResponse struct {
	Token Token `json:"token"`
	User  User  `json:"user"`
}

type OAuthResponse struct {
	Resultcode string      `json:"resultcode"`
	Message    string      `json:"message"`
	Response   UserReponse `json:"response"`
}

type UserReponse struct {
	ID           string `json:"id"`
	Nickname     string `json:"nickname"`
	ProfileImage string `json:"profile_image"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Birthday     string `json:"birthday"`
	Birthyear    string `json:"birthyear"`
}
