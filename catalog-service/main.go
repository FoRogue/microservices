// Package main is the entrypoint for catalog-service.
//
// @title Catalog Service API
// @version 1.0
// @description Сервис каталога товаров и корзины покупок.
// @BasePath /
package main

import (
	"log"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/ForRogue/microservices/catalog-service/docs"
	"github.com/ForRogue/microservices/catalog-service/internal"
)

func main() {
	cfg := internal.LoadConfig()

	db, err := internal.OpenDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	handler := internal.NewHandler(db)

	mux := http.NewServeMux()
	mux.Handle("/", handler.Routes())
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Printf("catalog-service listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
