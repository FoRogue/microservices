// Package main is the entrypoint for auth-service.
//
// @title Auth Service API
// @version 1.0
// @description Сервис авторизации/регистрации пользователей.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"log"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/ForRogue/microservices/auth-service/docs"
	"github.com/ForRogue/microservices/auth-service/internal"
)

func main() {
	cfg := internal.LoadConfig()

	db, err := internal.OpenDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	handler := internal.NewHandler(db, cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.Handle("/", handler.Routes())
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Printf("auth-service listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
