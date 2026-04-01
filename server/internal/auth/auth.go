package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"remotecontrol/server/internal/model"
)

type User struct {
	Username string
	Password string
	Role     model.Role
}

type Claims struct {
	Role model.Role `json:"role"`
	jwt.RegisteredClaims
}

type AuthenticatedUser struct {
	Username string
	Role     model.Role
}

type Service struct {
	secret []byte
	users  map[string]User
	ttl    time.Duration
}

func NewServiceFromEnv() *Service {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change-me-please"
	}

	return &Service{
		secret: []byte(secret),
		ttl:    12 * time.Hour,
		users: map[string]User{
			defaultIfEmpty(os.Getenv("ADMIN_USER"), "admin"): {
				Username: defaultIfEmpty(os.Getenv("ADMIN_USER"), "admin"),
				Password: defaultIfEmpty(os.Getenv("ADMIN_PASS"), "admin123"),
				Role:     model.RoleAdmin,
			},
			defaultIfEmpty(os.Getenv("VIEWER_USER"), "viewer"): {
				Username: defaultIfEmpty(os.Getenv("VIEWER_USER"), "viewer"),
				Password: defaultIfEmpty(os.Getenv("VIEWER_PASS"), "viewer123"),
				Role:     model.RoleViewer,
			},
			defaultIfEmpty(os.Getenv("AGENT_USER"), "agent"): {
				Username: defaultIfEmpty(os.Getenv("AGENT_USER"), "agent"),
				Password: defaultIfEmpty(os.Getenv("AGENT_PASS"), "agent123"),
				Role:     model.RoleAgent,
			},
		},
	}
}

func (s *Service) Login(username, password string) (string, model.Role, error) {
	user, ok := s.users[username]
	if !ok || user.Password != password {
		return "", "", errors.New("invalid username or password")
	}

	now := time.Now()
	claims := Claims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			Issuer:    "remotecontrol-server",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", "", err
	}
	return signed, user.Role, nil
}

func (s *Service) Verify(tokenString string) (*AuthenticatedUser, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &AuthenticatedUser{
		Username: claims.Subject,
		Role:     claims.Role,
	}, nil
}

func defaultIfEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
