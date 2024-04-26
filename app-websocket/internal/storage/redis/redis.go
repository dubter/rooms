package redis

import (
	"app-websocket/internal/config"
	"app-websocket/internal/domain"
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

func (r *Redis) GetLastMessagesFromRoom(ctx context.Context, roomID string, count int) ([]domain.Message, error) {
	jsonMsgs, err := r.client.LRange(ctx, roomID, 0, int64(count)).Result()
	if err != nil {
		return nil, fmt.Errorf("storage.redis.GetList: %w", err)
	}

	var messages []domain.Message
	for _, jsonMsg := range jsonMsgs {
		var msg domain.Message
		err = json.Unmarshal([]byte(jsonMsg), &msg)
		if err != nil {
			return nil, fmt.Errorf("storage.redis.GetList: %w", err)
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *Redis) GetRoomClients(ctx context.Context, roomID string) ([]domain.User, error) {
	key := "room:" + roomID
	hashTable, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("storage.redis.GetRoomClients: %w", err)
	}

	users := mapToUsers(hashTable)

	return users, nil
}

func (r *Redis) AddRoomClient(ctx context.Context, roomID string, user *domain.User) error {
	key := "room:" + roomID
	err := r.client.HSet(ctx, key, user.ID, user.Nickname).Err()
	if err != nil {
		return fmt.Errorf("storage.redis.AddRoomClient: %w", err)
	}

	return nil
}

func (r *Redis) DeleteClient(ctx context.Context, roomID string, user *domain.User) error {
	key := "room:" + roomID
	err := r.client.HDel(ctx, key, user.ID).Err()
	if err != nil {
		return fmt.Errorf("storage.redis.DeleteClient: %w", err)
	}

	return nil
}

func (r *Redis) Close() {
	err := r.client.Close()
	if err != nil {
		r.logger.Error("storage.redis.Close", slog.String("error", err.Error()))
	}
}
