package model

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenMetaData struct {
	AccessUUID   string
	AccessToken  string
	AtExpire     int64
	RefreshUUID  string
	RefreshToken string
	RtExpire     int64
	UserID       uint64
}
