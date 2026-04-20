package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"project-management/internal/auth"
	"project-management/internal/httpx"
	"project-management/internal/model"

	"github.com/gin-gonic/gin"
)

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "test-secret")

	newRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(JWTAuth())
		r.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"userID":    c.GetUint("userID"),
				"userEmail": c.GetString("userEmail"),
			})
		})
		return r
	}

	t.Run("missing header", func(t *testing.T) {
		w := httptest.NewRecorder()
		newRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))
		assertMiddlewareError(t, w, "missing authorization header")
	})

	t.Run("invalid header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Token abc")
		newRouter().ServeHTTP(w, req)
		assertMiddlewareError(t, w, "invalid authorization header")
	})

	t.Run("invalid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer nope")
		newRouter().ServeHTTP(w, req)
		assertMiddlewareError(t, w, "invalid or expired token")
	})

	t.Run("success", func(t *testing.T) {
		token, err := auth.IssueToken(model.User{ID: 7, Email: "u@example.com"})
		if err != nil {
			t.Fatalf("IssueToken error = %v", err)
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		newRouter().ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		var resp struct {
			UserID    uint   `json:"userID"`
			UserEmail string `json:"userEmail"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.UserID != 7 || resp.UserEmail != "u@example.com" {
			t.Fatalf("resp = %+v", resp)
		}
	})
}

func assertMiddlewareError(t *testing.T, w *httptest.ResponseRecorder, wantMsg string) {
	t.Helper()
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	var resp httpx.APIError
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != httpx.CodeUnauthorized || resp.Message != wantMsg {
		t.Fatalf("resp = %+v", resp)
	}
}
