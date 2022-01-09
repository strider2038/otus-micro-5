package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"identity-service/internal/di"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	version = ""
)

func main() {
	config := di.Config{Version: version}
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		log.Fatal("invalid config:", err)
	}

	connection, err := pgxpool.Connect(context.Background(), config.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to postgres:", err)
	}
	router, err := di.NewRouter(connection, config)
	if err != nil {
		log.Fatal("failed to create router: ", err)
	}

	address := fmt.Sprintf(":%d", config.Port)
	log.Println("starting HTTP server at", address)
	log.Fatal(http.ListenAndServe(address, router))
}
