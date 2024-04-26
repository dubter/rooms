package pg

import (
	"app-websocket/internal/domain"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
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

func (pg *Postgres) GetLastMessagesFromRoom(ctx context.Context, roomID string, count int) ([]domain.Message, error) {
	rows, err := pg.pool.Query(ctx,
		`SELECT m.content, u.nickname, m.user_id, m.time_created FROM messages AS m 
    		JOIN users AS u ON m.user_id = u.id 
            WHERE m.room_id = $1
            ORDER BY time_created
            LIMIT $2`, roomID, count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domain.Message{}, nil
		}

		return nil, fmt.Errorf("storage.pg.GetLastMessagesFromRoom: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		msg := domain.Message{}
		err = rows.Scan(&msg.Content, &msg.Nickname, &msg.UserID, &msg.TimeCreated)
		if err != nil {
			return nil, fmt.Errorf("storage.pg.GetLastMessagesFromRoom: %w", err)
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

func (pg *Postgres) SaveUser(ctx context.Context, user *domain.User) (string, error) {
	row := pg.pool.QueryRow(ctx, "INSERT INTO users(nickname, password_hash) VALUES ($1, $2) RETURNING id", user.Nickname, user.PasswordHash)

	var id string
	err := row.Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.ConstraintName != "" {
			return "", domain.ErrNicknameAlreadyExist
		}

		return "", fmt.Errorf("storage.pg.SaveUser: %w", err)
	}

	return id, nil
}

func (pg *Postgres) GetUser(ctx context.Context, nickname string) (*domain.User, error) {
	row := pg.pool.QueryRow(ctx, "SELECT id, nickname, password_hash FROM users WHERE nickname = $1", nickname)

	var user domain.User
	err := row.Scan(&user.ID, &user.Nickname, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("storage.pg.GetUser: %w", err)
	}

	return &user, nil
}

func (pg *Postgres) SetSession(ctx context.Context, userID string, session *domain.Session) error {
	_, err := pg.pool.Exec(ctx, "UPDATE users SET refresh_token = $1, expires_at = $2 WHERE id = $3", session.RefreshToken, session.ExpiresAt, userID)
	if err != nil {
		return fmt.Errorf("storage.pg.SetSession: %w", err)
	}

	return nil
}

func (pg *Postgres) GetBySession(ctx context.Context, refreshToken string) (*domain.User, error) {
	row := pg.pool.QueryRow(ctx, "SELECT id, nickname FROM users WHERE refresh_token = $1", refreshToken)

	var user domain.User
	err := row.Scan(&user.ID, &user.Nickname)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("storage.pg.GetBySession: %w", err)
	}

	return &user, nil
}

func (pg *Postgres) GetAllRooms(ctx context.Context) ([]domain.Room, error) {
	rows, err := pg.pool.Query(ctx, "SELECT id, name, time_created FROM rooms")
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domain.Room{}, nil
		}

		return nil, fmt.Errorf("storage.pg.GetRooms: %w", err)
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var room domain.Room
		err = rows.Scan(&room.ID, &room.Name, &room.TimeCreated)
		if err != nil {
			return nil, fmt.Errorf("storage.pg.GetRooms: %w", err)
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (pg *Postgres) CreateRoom(ctx context.Context, name string) (*domain.Room, error) {
	timeCreated := time.Now()
	row := pg.pool.QueryRow(ctx, "INSERT INTO rooms(name, time_created) VALUES ($1, $2) RETURNING id", name, timeCreated)

	var id string
	err := row.Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("storage.pg.CreateRoom: %w", err)
	}

	return &domain.Room{
		ID:          id,
		Name:        name,
		TimeCreated: timeCreated,
	}, nil
}

func (pg *Postgres) GetRoom(ctx context.Context, roomID string) (*domain.Room, error) {
	row := pg.pool.QueryRow(ctx, "SELECT id, name, time_created FROM rooms WHERE id = $1", roomID)

	var room domain.Room
	err := row.Scan(&room.ID, &room.Name, &room.TimeCreated)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, fmt.Errorf("storage.pg.GetRoom: %w", err)
	}

	return &room, nil
}
