package ws

import (
	"app-websocket/internal/domain"
	"context"
	"log/slog"
	"math"
	"math/rand"
	"sync"
	"time"
)

type MessageConsumer interface {
	Consume(ctx context.Context, handler domain.MessageHandler) error
}

type ClientsService interface {
	AddRoomClient(ctx context.Context, roomID string, user *domain.User) error
	DeleteClient(ctx context.Context, roomID, userID string) error
}

type Hub struct {
	logger   *slog.Logger
	consumer MessageConsumer
	clients  map[string]map[string]*Client // pull of connections in current server
	mu       sync.Mutex
}

func NewHub(consumer MessageConsumer, logger *slog.Logger) *Hub {
	return &Hub{
		consumer: consumer,
		logger:   logger,
		clients:  make(map[string]map[string]*Client),
	}
}

func (h *Hub) AddConnection(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.RoomID] == nil {
		h.clients[client.RoomID] = make(map[string]*Client)
	}

	h.clients[client.RoomID][client.User.ID] = client
}

func (h *Hub) DeleteConnection(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients[client.RoomID], client.User.ID)
}

func (h *Hub) Run(ctx context.Context) {
	attempt := 0

	go func() {
		for {
			err := h.consumer.Consume(ctx, func(msg domain.Message) error {
				h.mu.Lock()
				connections := h.clients[msg.RoomID]
				h.mu.Unlock()

				for _, conn := range connections {
					conn.Message <- &Message{
						Content:     msg.Content,
						TimeCreated: msg.TimeCreated,
						RoomID:      msg.RoomID,
						Username:    msg.Nickname,
						UserID:      msg.UserID,
					}
				}

				return nil
			})

			if err != nil {
				h.logger.Error("failed to consume message:", slog.String("error", err.Error()))

				time.Sleep(expBackoff(attempt))
				attempt++
			}
		}
	}()

	<-ctx.Done()
}

func expBackoff(attempt int) time.Duration {
	maxDelay := 30 * time.Second
	backoff := math.Pow(2, float64(attempt))
	delay := time.Duration(backoff) * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}

	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	return delay + jitter
}
