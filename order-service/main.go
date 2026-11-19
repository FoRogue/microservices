// Package main is the entrypoint for order-service.
//
// @title Order Service API
// @version 1.0
// @description Creation, payment and status of customer orders.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter Bearer followed by the token from auth-service.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/ForRogue/microservices/eventing"
	_ "github.com/ForRogue/microservices/order-service/docs"
	"github.com/ForRogue/microservices/order-service/internal"
)

func main() {
	if len(os.Args) > 1 && (len(os.Args) != 2 || os.Args[1] != "migrate") {
		log.Fatal("usage: order-service [migrate]")
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
	clients := internal.NewClients(cfg)
	handler := internal.NewHandler(internal.NewService(repo, clients), clients)

	mux := http.NewServeMux()
	mux.Handle("/", handler.Routes())
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Printf("order-service listening on :%s", cfg.Port)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := eventing.Serve(server,
		func(ctx context.Context) { eventing.Publish(ctx, db, eventing.PaymentsRequested) },
		func(ctx context.Context) {
			eventing.Consume(ctx, eventing.OrderGroup, eventing.PaymentsCompleted, repo.ApplyPaymentResult)
		},
	); err != nil {
		log.Fatal(err)
	}
}
