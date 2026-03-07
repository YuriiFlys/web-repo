package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"
	"project-management/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type mockAuthService struct {
	registerFn func(ctx context.Context, input service.RegisterInput) (model.User, string, error)
	loginFn    func(ctx context.Context, input service.LoginInput) (model.User, string, error)
	getByIDFn  func(ctx context.Context, id uint) (model.User, error)
}

func (m *mockAuthService) Register(ctx context.Context, input service.RegisterInput) (model.User, string, error) {
	return m.registerFn(ctx, input)
}

func (m *mockAuthService) Login(ctx context.Context, input service.LoginInput) (model.User, string, error) {
	return m.loginFn(ctx, input)
}

func (m *mockAuthService) GetByID(ctx context.Context, id uint) (model.User, error) {
	return m.getByIDFn(ctx, id)
}

func TestAuthHandlerRegisterUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createdAt := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		body       string
		service    *mockAuthService
		wantStatus int
		wantErr    *httpx.APIError
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			body: `{"email":"USER@example.com","password":"secret1","name":" Alice "}`,
			service: &mockAuthService{registerFn: func(ctx context.Context, input service.RegisterInput) (model.User, string, error) {
				if input.Email != "USER@example.com" || input.Name != " Alice " {
					t.Fatalf("unexpected input: %+v", input)
				}
				return model.User{ID: 7, Email: "user@example.com", Name: "Alice", CreatedAt: createdAt}, "token-123", nil
			}},
			wantStatus: http.StatusCreated,
			assertBody: func(t *testing.T, body []byte) {
				var resp AuthResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if resp.Token != "token-123" || resp.User.ID != 7 || resp.User.Email != "user@example.com" {
					t.Fatalf("unexpected response: %+v", resp)
				}
			},
		},
		{
			name:       "bad request",
			body:       `{"email":"bad","password":"123","name":""}`,
			service:    &mockAuthService{},
			wantStatus: http.StatusBadRequest,
			wantErr:    &httpx.APIError{Code: httpx.CodeBadRequest},
		},
		{
			name: "duplicate email",
			body: `{"email":"user@example.com","password":"secret1","name":"Alice"}`,
			service: &mockAuthService{registerFn: func(ctx context.Context, input service.RegisterInput) (model.User, string, error) {
				return model.User{}, "", service.ErrEmailInUse
			}},
			wantStatus: http.StatusBadRequest,
			wantErr:    &httpx.APIError{Code: httpx.CodeBadRequest, Message: "email already in use"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAuthHandler(tt.service)
			r := gin.New()
			r.POST("/auth/register", h.RegisterUser)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantErr != nil {
				var got httpx.APIError
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal error response: %v", err)
				}
				if got.Code != tt.wantErr.Code {
					t.Fatalf("error code = %q, want %q", got.Code, tt.wantErr.Code)
				}
				if tt.wantErr.Message != "" && got.Message != tt.wantErr.Message {
					t.Fatalf("error message = %q, want %q", got.Message, tt.wantErr.Message)
				}
			}
			if tt.assertBody != nil {
				tt.assertBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestAuthHandlerLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createdAt := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		body       string
		service    *mockAuthService
		wantStatus int
		wantErr    string
	}{
		{
			name:       "bad request",
			body:       `{"email":"bad","password":"123","name":""}`,
			service:    &mockAuthService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "success",
			body: `{"email":"User@example.com","password":"secret1"}`,
			service: &mockAuthService{loginFn: func(ctx context.Context, input service.LoginInput) (model.User, string, error) {
				if input.Email != "user@example.com" {
					t.Fatalf("email = %q", input.Email)
				}
				return model.User{ID: 3, Email: input.Email, Name: "Bob", CreatedAt: createdAt}, "jwt", nil
			}},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			body: `{"email":"user@example.com","password":"wrong"}`,
			service: &mockAuthService{loginFn: func(ctx context.Context, input service.LoginInput) (model.User, string, error) {
				return model.User{}, "", bcrypt.ErrMismatchedHashAndPassword
			}},
			wantStatus: http.StatusUnauthorized,
			wantErr:    "invalid credentials",
		},
		{
			name: "internal error",
			body: `{"email":"user@example.com","password":"secret1"}`,
			service: &mockAuthService{loginFn: func(ctx context.Context, input service.LoginInput) (model.User, string, error) {
				return model.User{}, "", errors.New("boom")
			}},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAuthHandler(tt.service)
			r := gin.New()
			r.POST("/auth/login", h.Login)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantErr != "" {
				var got httpx.APIError
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal error response: %v", err)
				}
				if got.Message != tt.wantErr {
					t.Fatalf("message = %q, want %q", got.Message, tt.wantErr)
				}
			}
		})
	}
}

func TestAuthHandlerMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createdAt := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)

	t.Run("missing user id", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthService{})
		r := gin.New()
		r.GET("/auth/me", h.Me)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/me", nil))

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("success", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthService{getByIDFn: func(ctx context.Context, id uint) (model.User, error) {
			if id != 9 {
				t.Fatalf("id = %d", id)
			}
			return model.User{ID: 9, Email: "me@example.com", Name: "Me", CreatedAt: createdAt}, nil
		}})
		r := gin.New()
		r.GET("/auth/me", func(c *gin.Context) {
			c.Set("userID", uint(9))
			h.Me(c)
		})

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/me", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		var resp AuthUser
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Email != "me@example.com" || resp.ID != 9 {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})
}
