package components

import (
	"app-websocket/internal/broker/kafka"
	"app-websocket/internal/config"
	"app-websocket/internal/ports"
	"app-websocket/internal/ports/ws"
	"app-websocket/internal/services/auth"
	"app-websocket/internal/services/message_cache"
	"app-websocket/internal/services/message_online"
	"app-websocket/internal/services/rooms"
	"app-websocket/internal/storage/pg"
	"app-websocket/internal/storage/redis"
	"app-websocket/pkg/jwt"
	"app-websocket/pkg/logger/slogpretty"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Components struct {
	HttpServer         *ports.Server
	Postgres           *pg.Postgres
	Redis              *redis.Redis
	KafkaProducer      *kafka.KafkaProducer
	KafkaConsumerGroup *kafka.ConsumerGroup
}

func InitComponents(cfg *config.Config, logger *slog.Logger) (*Components, error) {
	postgres, err := pg.New(cfg.Postgres.PostgresURL)
	if err != nil {
		return nil, err
	}

	rds, err := redis.New(&cfg.Redis, logger)
	if err != nil {
		return nil, err
	}

	kafkaProducer, err := kafka.NewProducer(&cfg.Kafka, logger)
	if err != nil {
		return nil, err
	}

	kafkaConsumerGroup, err := kafka.NewConsumerGroup(&cfg.Kafka, logger)
	if err != nil {
		return nil, err
	}

	hub := ws.NewHub(kafkaConsumerGroup, logger)

	serviceAuth, err := auth.New(&cfg.Auth, postgres)
	if err != nil {
		return nil, err
	}

	roomService := rooms.New(postgres)

	chatCache := message_cache.New(&cfg.Chat, rds, postgres)

	chatOnline := message_online.New(kafkaProducer, kafkaConsumerGroup, rds, hub)

	tokenManager, err := jwt.NewManager(cfg.Auth.JWTSigningKey)
	if err != nil {
		return nil, err
	}

	httpServer, err := ports.NewServer(&cfg.Http, serviceAuth, chatCache, chatOnline, roomService, logger, tokenManager, hub)
	if err != nil {
		return nil, err
	}

	return &Components{
		HttpServer:         httpServer,
		Postgres:           postgres,
		Redis:              rds,
		KafkaProducer:      kafkaProducer,
		KafkaConsumerGroup: kafkaConsumerGroup,
	}, nil
}

func (c *Components) Shutdown() {
	c.Postgres.CloseConnection()
	c.Redis.Close()
	c.KafkaProducer.Close()
	c.KafkaConsumerGroup.Close()
	c.HttpServer.Stop()
}

func SetupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slogpretty.SetupPrettySlog()
	case envDev:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return logger
}
