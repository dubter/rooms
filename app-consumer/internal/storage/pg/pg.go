package pg

import (
	"app-consumer/internal/domain"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func New(pgURL string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(pgURL)
	if err != nil {
		return nil, fmt.Errorf("storage.pg.New: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("storage.pg.New: %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("storage.pg.New: %w", err)
	}

	return &Postgres{
		pool: pool,
	}, nil
}

func (pg *Postgres) CloseConnection() {
	pg.pool.Close()
}

func (pg *Postgres) PushMessage(ctx context.Context, msg *domain.Message) error {
	_, err := pg.pool.Exec(ctx, "INSERT INTO messages(user_id, content, room_id, time_created) VALUES ($1, $2, $3, $4)", msg.UserID, msg.Content, msg.RoomID, msg.TimeCreated)
	if err != nil {
		return fmt.Errorf("storage.pg.PushMessage: %w", err)
	}

	return nil
}
