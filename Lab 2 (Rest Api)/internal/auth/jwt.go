package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"project-management/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultTokenTTLMinutes = 60
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID uint   `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func IssueToken(user model.User) (string, error) {
	ttl := time.Duration(getEnvInt("JWT_TTL_MINUTES", defaultTokenTTLMinutes)) * time.Minute
	now := time.Now().UTC()

	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret()))
}

func ParseToken(tokenStr string) (*Claims, error) {
	if tokenStr == "" {
		return nil, ErrInvalidToken
	}

	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret()), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func secret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "dev-secret"
	}
	return secret
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
