package main

import (
	"context"
	"fmt"
	"log"

	"order-service/internal/di"

	"github.com/go-redis/redis/v8"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/strider2038/httpserver"
	"github.com/strider2038/ossync"
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

	postgresConnection, err := pgxpool.Connect(context.Background(), config.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to postgres:", err)
	}
	redisClient := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    config.RedisURL,
	})
	ping := redisClient.Ping(context.Background())
	err = ping.Err()
	if err != nil {
		log.Fatal("failed to connect to redis:", err)
	}

	router, err := di.NewAPIRouter(postgresConnection, redisClient, config)
	if err != nil {
		log.Fatal("failed to create router: ", err)
	}

	address := fmt.Sprintf(":%d", config.Port)
	log.Println("starting application server at", address)

	group := ossync.NewGroup(context.Background())
	group.Go(httpserver.New(address, router).ListenAndServe)
	group.Go(di.NewBillingConsumer(postgresConnection, config).Run)
	err = group.Wait()
	if err != nil {
		log.Fatalln(err)
	}
}
