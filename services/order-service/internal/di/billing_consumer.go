package di

import (
	"order-service/internal/kafka"
	"order-service/internal/messaging"
	"order-service/internal/postgres"
	"order-service/internal/postgres/database"

	"github.com/jackc/pgx/v4/pgxpool"
	segmentio "github.com/segmentio/kafka-go"
)

func NewBillingConsumer(connection *pgxpool.Pool, config Config) *kafka.Consumer {
	db := database.New(connection)
	orders := postgres.NewOrderRepository(db)

	writer := &segmentio.Writer{
		Addr:     segmentio.TCP(config.KafkaProducerURL),
		Topic:    "order-events",
		Balancer: &segmentio.LeastBytes{},
	}
	dispatcher := kafka.NewDispatcher(writer)

	reader := segmentio.NewReader(segmentio.ReaderConfig{
		Brokers: []string{config.KafkaConsumerURL},
		GroupID: "order",
		Topic:   "billing-events",
	})

	consumer := kafka.NewConsumer(reader, map[string]kafka.Processor{
		"Billing/PaymentCreated": messaging.NewPaymentCreatedProcessor(orders, dispatcher),
		"Billing/PaymentFailed":  messaging.NewPaymentFailedProcessor(orders, dispatcher),
	})

	return consumer
}
