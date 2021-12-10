package main

import (
	"fmt"
	"net/http"
	"pdserver/pkg/api"
	"pdserver/pkg/app"
	"pdserver/pkg/repository"
)

type Application struct{}

func NewApp() *Application {
	return &Application{}
}

func (application *Application) Run() error {
	fmt.Println("App running start...")
	localDB, err := repository.NewDatabase()
	if err != nil {
		return err
	}
	defer localDB.Close()
	redisDB, err := repository.NewClient()
	if err != nil {
		return err
	}
	defer redisDB.Close()
	userService := api.NewUserDB(localDB)
	tokenService := api.NewTokenDB(redisDB)
	naverService := api.NewNaverService()
	otpService := api.NewOTPService()
	handler := app.NewHandler(userService, tokenService, naverService, otpService)
	handler.SetupRoutes()
	if err := http.ListenAndServe(":8080", handler.Engin); err != nil {
		return err
	}
	return nil
}

func main() {
	application := NewApp()
	if err := application.Run(); err != nil {
		fmt.Println(err.Error())
		return
	}
}
