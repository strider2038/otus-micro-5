package di

import (
	"context"
	"encoding/json"
	"net/http"

	"order-service/internal/api"
	"order-service/internal/kafka"
	"order-service/internal/postgres"
	"order-service/internal/postgres/database"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	segmentio "github.com/segmentio/kafka-go"
)

func NewAPIRouter(
	postgresConnection *pgxpool.Pool,
	redisClient *redis.Client,
	config Config,
) (http.Handler, error) {
	db := database.New(postgresConnection)

	writer := &segmentio.Writer{
		Addr:     segmentio.TCP(config.KafkaProducerURL),
		Topic:    "billing-commands",
		Balancer: &segmentio.LeastBytes{},
	}
	dispatcher := kafka.NewDispatcher(writer)

	locker := redislock.New(redisClient)

	orders := postgres.NewOrderRepository(db)
	billingApiService := api.NewOrderingApiService(orders, dispatcher, locker)
	billingApiController := api.NewOrderingApiController(billingApiService)

	apiRouter := api.NewRouter(billingApiController)
	metrics := api.NewMetrics("order_service")
	apiRouter.Use(func(handler http.Handler) http.Handler {
		return api.MetricsMiddleware(handler, metrics)
	})

	router := NewRouter(postgresConnection, config)
	router.PathPrefix("/api").Handler(apiRouter)
	router.Handle("/metrics", promhttp.Handler())

	return router, nil
}

func NewRouter(connection *pgxpool.Pool, config Config) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(`{"status":"ok"}`))
	})

	router.HandleFunc("/ready", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		err := connection.Ping(context.Background())
		if err == nil {
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte(`{"status":"ok"}`))
		} else {
			writer.WriteHeader(http.StatusServiceUnavailable)
			writer.Write([]byte(`{"status":"not available"}`))
		}
	})

	router.HandleFunc("/version", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(http.StatusOK)
		response, _ := json.Marshal(struct {
			ApplicationName    string `json:"application_name"`
			Environment        string `json:"environment"`
			ApplicationVersion string `json:"application_version"`
		}{
			ApplicationName:    "BillingService",
			Environment:        config.Environment,
			ApplicationVersion: config.Version,
		})
		writer.Write(response)
	})

	return router
}
