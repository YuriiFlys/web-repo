package auth

import (
	"errors"
	"testing"
	"time"

	"project-management/internal/config"
	"project-management/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueAndParseToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_TTL_MINUTES", "5")

	token, err := IssueToken(model.User{ID: 42, Email: "user@example.com"})
	if err != nil {
		t.Fatalf("IssueToken error = %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken error = %v", err)
	}
	if claims.UserID != 42 || claims.Email != "user@example.com" || claims.Subject != "42" {
		t.Fatalf("claims = %+v", claims)
	}
	if claims.ExpiresAt == nil || time.Until(claims.ExpiresAt.Time) > 5*time.Minute+time.Minute {
		t.Fatalf("unexpected expiry: %+v", claims.ExpiresAt)
	}
}

func TestParseTokenErrors(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	t.Run("empty token", func(t *testing.T) {
		_, err := ParseToken("")
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("wrong signing method", func(t *testing.T) {
		tok := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{UserID: 1})
		signed, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("SignedString error = %v", err)
		}
		_, err = ParseToken(signed)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		expired := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
			UserID: 1,
			Email:  "a@example.com",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			},
		})
		signed, err := expired.SignedString([]byte("test-secret"))
		if err != nil {
			t.Fatalf("SignedString error = %v", err)
		}
		_, err = ParseToken(signed)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestSecretAndGetEnvInt(t *testing.T) {
	t.Run("secret default", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "")
		if got := secret(); got != "dev-secret" {
			t.Fatalf("secret = %q", got)
		}
	})

	t.Run("secret from env", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "abc")
		if got := secret(); got != "abc" {
			t.Fatalf("secret = %q", got)
		}
	})

	t.Run("get env int cases", func(t *testing.T) {
		t.Setenv("JWT_TTL_MINUTES", "15")
		if got := config.GetEnvInt("JWT_TTL_MINUTES", 1); got != 15 {
			t.Fatalf("got = %d", got)
		}
		t.Setenv("JWT_TTL_MINUTES", "bad")
		if got := config.GetEnvInt("JWT_TTL_MINUTES", 7); got != 7 {
			t.Fatalf("got = %d", got)
		}
	})
}
