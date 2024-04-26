package message_online

import (
	"app-websocket/internal/domain"
	"app-websocket/internal/ports/ws"
	"context"
	"fmt"
	"time"
)

type MessagePusher interface {
	Produce(msg *domain.Message) error
}

type MessageConsumer interface {
	Consume(ctx context.Context, handler domain.MessageHandler) error
}

type RoomClientsStorage interface {
	AddRoomClient(ctx context.Context, roomID string, user *domain.User) error
	DeleteClient(ctx context.Context, roomID string, user *domain.User) error
}

type MessageOnlineService struct {
	pusher      MessagePusher
	consumer    MessageConsumer
	roomClients RoomClientsStorage
	hub         *ws.Hub
}

func New(pusher MessagePusher, consumer MessageConsumer, roomClients RoomClientsStorage, hub *ws.Hub) *MessageOnlineService {
	return &MessageOnlineService{
		pusher:      pusher,
		consumer:    consumer,
		roomClients: roomClients,
		hub:         hub,
	}
}

func (m *MessageOnlineService) PushMessage(_ context.Context, msg *domain.Message) error {
	return m.pusher.Produce(msg)
}

func (m *MessageOnlineService) Consume(ctx context.Context, handler func(message domain.Message) error) error {
	return m.consumer.Consume(ctx, handler)
}

func (m *MessageOnlineService) Subscribe(ctx context.Context, client *ws.Client) error {
	m.hub.AddConnection(client)

	err := m.roomClients.AddRoomClient(ctx, client.RoomID, client.User)
	if err != nil {
		return fmt.Errorf("service.MessageOnlineService.Subscribe: %w", err)
	}

	return m.pusher.Produce(&domain.Message{
		Content:     "joined the room",
		RoomID:      client.RoomID,
		UserID:      client.User.ID,
		TimeCreated: time.Now(),
		Nickname:    client.User.Nickname,
	})
}

func (m *MessageOnlineService) Unsubscribe(ctx context.Context, client *ws.Client) error {
	m.hub.DeleteConnection(client)

	err := m.roomClients.DeleteClient(ctx, client.RoomID, client.User)
	if err != nil {
		return fmt.Errorf("service.MessageOnlineService.Unsubscribe: %w", err)
	}

	return m.pusher.Produce(&domain.Message{
		Content:     "left the room",
		RoomID:      client.RoomID,
		UserID:      client.User.ID,
		TimeCreated: time.Now(),
		Nickname:    client.User.Nickname,
	})
}
