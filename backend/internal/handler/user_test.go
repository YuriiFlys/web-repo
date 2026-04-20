package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"project-management/internal/httpx"
	"project-management/internal/model"

	"github.com/gin-gonic/gin"
)

type mockUserService struct {
	listFn func(ctx context.Context) ([]model.User, error)
}

func (m *mockUserService) List(ctx context.Context) ([]model.User, error) {
	return m.listFn(ctx)
}

func TestUserHandlerList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		h := NewUserHandler(&mockUserService{listFn: func(ctx context.Context) ([]model.User, error) {
			return []model.User{{ID: 1, Email: "a@example.com", Name: "Alice", PasswordHash: "secret"}}, nil
		}})
		r := gin.New()
		r.GET("/users", h.List)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		var resp UsersListResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if len(resp.Items) != 1 || resp.Items[0].Email != "a@example.com" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		h := NewUserHandler(&mockUserService{listFn: func(ctx context.Context) ([]model.User, error) {
			return nil, errors.New("db down")
		}})
		r := gin.New()
		r.GET("/users", h.List)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users", nil))

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
		var apiErr httpx.APIError
		if err := json.Unmarshal(w.Body.Bytes(), &apiErr); err != nil {
			t.Fatalf("unmarshal error response: %v", err)
		}
		if apiErr.Message != "db down" {
			t.Fatalf("message = %q", apiErr.Message)
		}
	})
}
