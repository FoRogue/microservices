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
	"os"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/ForRogue/microservices/auth-service/docs"
	"github.com/ForRogue/microservices/auth-service/internal"
)

func main() {
	if len(os.Args) > 1 && (len(os.Args) != 2 || os.Args[1] != "migrate") {
		log.Fatal("usage: auth-service [migrate]")
	}
	cfg := internal.LoadConfig()

	db, err := internal.OpenDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if len(os.Args) == 2 {
		if err := internal.Migrate(db); err != nil {
			log.Fatalf("failed to apply migrations: %v", err)
		}
		return
	}

	repo := internal.NewRepository(db)
	handler := internal.NewHandler(internal.NewService(repo, cfg.JWTSecret))

	mux := http.NewServeMux()
	mux.Handle("/", handler.Routes())
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Printf("auth-service listening on :%s", cfg.Port)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
