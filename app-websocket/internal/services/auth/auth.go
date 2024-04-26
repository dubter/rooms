package auth

import (
	"app-websocket/internal/config"
	"app-websocket/internal/domain"
	"app-websocket/pkg/hash"
	"app-websocket/pkg/jwt"
	"context"
	"fmt"
	"time"
)

type UserStorage interface {
	SaveUser(ctx context.Context, user *domain.User) (string, error)
	GetUser(ctx context.Context, nickname string) (*domain.User, error)
	SetSession(ctx context.Context, userID string, session *domain.Session) error
	GetBySession(ctx context.Context, refreshToken string) (*domain.User, error)
}

type Auth struct {
	storage         UserStorage
	tokenManager    jwt.TokenManager
	hasher          hash.PasswordHasher
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func New(config *config.AuthConfig, storage UserStorage) (*Auth, error) {
	tokenManager, err := jwt.NewManager(config.JWTSigningKey)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.New: %w", err)
	}

	hasher, err := hash.NewSHA1Hasher(config.PasswordSalt)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.New: %w", err)
	}

	return &Auth{
		storage:         storage,
		tokenManager:    tokenManager,
		hasher:          hasher,
		accessTokenTTL:  config.AccessTokenTTL,
		refreshTokenTTL: config.RefreshTokenTTL,
	}, nil
}

func (a *Auth) Register(ctx context.Context, nickname, password string) error {
	passHash, err := a.hasher.Hash(password)
	if err != nil {
		return fmt.Errorf("service.Auth.Register: %w", err)
	}

	user := &domain.User{
		Nickname:     nickname,
		PasswordHash: passHash,
	}

	_, err = a.storage.SaveUser(ctx, user)
	if err != nil {
		return fmt.Errorf("service.Auth.Register: %w", err)
	}

	return nil
}

func (a *Auth) Login(ctx context.Context, nickname, password string) (*domain.Tokens, *domain.User, error) {
	passwordHash, err := a.hasher.Hash(password)
	if err != nil {
		return nil, nil, fmt.Errorf("service.Auth.Login: %w", err)
	}

	user, err := a.storage.GetUser(ctx, nickname)
	if err != nil {
		return nil, nil, fmt.Errorf("service.Auth.Login: %w", err)
	}

	if user.PasswordHash != passwordHash {
		return nil, nil, domain.ErrInvalidCredentials
	}

	session, err := a.CreateSession(ctx, user)
	return session, user, err
}

func (a *Auth) Refresh(ctx context.Context, token string) (*domain.Tokens, error) {
	user, err := a.storage.GetBySession(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.Refresh: %w", err)
	}

	return a.CreateSession(ctx, user)
}

func (a *Auth) CreateSession(ctx context.Context, user *domain.User) (*domain.Tokens, error) {
	accessToken, err := a.tokenManager.NewJWT(user.ID, user.Nickname, a.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.CreateSession: %w", err)
	}

	refreshToken, err := a.tokenManager.NewRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("service.Auth.CreateSession: %w", err)
	}

	tokens := &domain.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	session := &domain.Session{
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(a.refreshTokenTTL),
	}

	err = a.storage.SetSession(ctx, user.ID, session)
	if err != nil {
		return nil, fmt.Errorf("service.Auth.CreateSession: %w", err)
	}

	return tokens, nil
}
