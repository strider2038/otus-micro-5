package di

import (
	"billing-service/internal/kafka"
	"billing-service/internal/messaging"
	"billing-service/internal/postgres"
	"billing-service/internal/postgres/database"

	"github.com/jackc/pgx/v4/pgxpool"
	segmentio "github.com/segmentio/kafka-go"
)

func NewIdentityConsumer(connection *pgxpool.Pool, config Config) *kafka.Consumer {
	db := database.New(connection)
	accounts := postgres.NewAccountRepository(db)

	reader := segmentio.NewReader(segmentio.ReaderConfig{
		Brokers: []string{config.KafkaConsumerURL},
		GroupID: "billing",
		Topic:   "identity-events",
	})

	consumer := kafka.NewConsumer(reader, map[string]kafka.Processor{
		"Identity/UserCreated": messaging.NewUserCreatedProcessor(accounts),
	})

	return consumer
}
