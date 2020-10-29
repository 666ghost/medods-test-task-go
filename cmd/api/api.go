package main

import (
	"github.com/666ghost/medods-test-task-go/auth/api"
	"github.com/666ghost/medods-test-task-go/config"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"log"
	"os"
	"time"
)

func init() {
	config.LoadFromFile()
}

func main() {
	now := time.Now()
	cfg := config.New()
	file, err := os.OpenFile("logs/api_logfile_"+now.Format("20060102")+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	log.SetOutput(file)

	var isLoggedIn = middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:    []byte(cfg.TokenSecret),
		SigningMethod: "HS512",
	})

	log.Print("Api server launched!")

	h := new(api.Handler)
	e := echo.New()

	e.POST("/api/register", h.Create)
	e.POST("/api/login", h.Login)
	e.POST("/api/security/token/refresh", h.Refresh, isLoggedIn)
	e.POST("/api/security/remove_refresh", h.RemoveToken, isLoggedIn)
	e.POST("/api/users/security/truncate_refresh", h.TruncateUserTokens, isLoggedIn)

	e.Logger.Fatal(e.Start(":" + cfg.APIPort))
}
