package model

import "os"

var (
	ACCESS_SECRET  = os.Getenv("ACCESS_SECRET")
	REFRESH_SECRET = os.Getenv("REFRESH_SECRET")
)
