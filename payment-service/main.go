// Package main is the entrypoint for payment-service.
//
// @title Payment Service API
// @version 1.0
// @description Сервис оплаты заказов.
// @BasePath /
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/ForRogue/microservices/eventing"
	_ "github.com/ForRogue/microservices/payment-service/docs"
	"github.com/ForRogue/microservices/payment-service/internal"
)

func main() {
	if len(os.Args) > 1 && (len(os.Args) != 2 || os.Args[1] != "migrate") {
		log.Fatal("usage: payment-service [migrate]")
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
	handler := internal.NewHandler(internal.NewService(repo))

	mux := http.NewServeMux()
	mux.Handle("/", handler.Routes())
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Printf("payment-service listening on :%s", cfg.Port)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := eventing.Serve(server,
		func(ctx context.Context) { eventing.Publish(ctx, db, eventing.PaymentsCompleted) },
		func(ctx context.Context) {
			eventing.Consume(ctx, eventing.PaymentGroup, eventing.PaymentsRequested, repo.ApplyPaymentRequest)
		},
	); err != nil {
		log.Fatal(err)
	}
}
