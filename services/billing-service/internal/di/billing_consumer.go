package di

import (
	"billing-service/internal/kafka"
	"billing-service/internal/messaging"
	"billing-service/internal/postgres"
	"billing-service/internal/postgres/database"

	"github.com/jackc/pgx/v4/pgxpool"
	segmentio "github.com/segmentio/kafka-go"
)

func NewBillingConsumer(connection *pgxpool.Pool, config Config) *kafka.Consumer {
	db := database.New(connection)
	accounts := postgres.NewAccountRepository(db)
	payments := postgres.NewPaymentRepository(db)

	writer := &segmentio.Writer{
		Addr:     segmentio.TCP(config.KafkaProducerURL),
		Topic:    "billing-events",
		Balancer: &segmentio.LeastBytes{},
	}
	dispatcher := kafka.NewDispatcher(writer)

	reader := segmentio.NewReader(segmentio.ReaderConfig{
		Brokers: []string{config.KafkaConsumerURL},
		GroupID: "billing",
		Topic:   "billing-commands",
	})

	consumer := kafka.NewConsumer(reader, map[string]kafka.Processor{
		"Billing/CreatePayment": messaging.NewCreatePaymentProcessor(accounts, payments, dispatcher),
	})

	return consumer
}
