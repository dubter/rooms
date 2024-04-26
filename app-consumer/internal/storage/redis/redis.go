package redis

import (
	"app-consumer/internal/config"
	"app-consumer/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type Redis struct {
	client redis.UniversalClient
	logger *slog.Logger
}

func New(config *config.RedisConfig, logger *slog.Logger) (*Redis, error) {
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    config.Addrs,
		Password: config.Password,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("storage.redis.New: %w", err)
	}

	return &Redis{
		client: client,
		logger: logger,
	}, nil
}

func (r *Redis) AddToList(ctx context.Context, msg *domain.Message) error {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("storage.redis.AddToList: %w", err)
	}

	return r.client.LPush(ctx, msg.RoomID, jsonMsg).Err()
}

func (r *Redis) Close() {
	err := r.client.Close()
	if err != nil {
		r.logger.Error("storage.redis.Close", slog.String("error", err.Error()))
	}
}
