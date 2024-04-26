package jwt

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// TokenManager provides logic for JWT & Refresh tokens generation and parsing.
type TokenManager interface {
	NewJWT(userId string, nickname string, ttl time.Duration) (string, error)
	Parse(accessToken string) (*UserInfo, error)
	NewRefreshToken() (string, error)
}

type Manager struct {
	signingKey string
}

func NewManager(signingKey string) (*Manager, error) {
	if signingKey == "" {
		return nil, fmt.Errorf("empty signing key")
	}

	return &Manager{signingKey: signingKey}, nil
}

func (m *Manager) NewJWT(userId string, nickname string, ttl time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = userId
	claims["nickname"] = nickname
	claims["exp"] = time.Now().Add(ttl).Unix()

	return token.SignedString([]byte(m.signingKey))
}

type UserInfo struct {
	UserID   string
	Nickname string
}

func (m *Manager) Parse(accessToken string) (*UserInfo, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.signingKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("error get user claims from token")
	}

	return &UserInfo{
		UserID:   claims["id"].(string),
		Nickname: claims["nickname"].(string),
	}, nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}
