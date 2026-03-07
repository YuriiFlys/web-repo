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
	"gorm.io/gorm"
)

type mockCommentService struct {
	listFn   func(ctx context.Context, filter service.CommentListFilter) ([]model.Comment, int64, error)
	createFn func(ctx context.Context, input service.CommentCreateInput) (model.Comment, error)
	getFn    func(ctx context.Context, id string) (model.Comment, error)
	updateFn func(ctx context.Context, id string, input service.CommentUpdateInput) (model.Comment, error)
	deleteFn func(ctx context.Context, id string) error
}

func (m *mockCommentService) List(ctx context.Context, filter service.CommentListFilter) ([]model.Comment, int64, error) {
	return m.listFn(ctx, filter)
}
func (m *mockCommentService) Create(ctx context.Context, input service.CommentCreateInput) (model.Comment, error) {
	return m.createFn(ctx, input)
}
func (m *mockCommentService) Get(ctx context.Context, id string) (model.Comment, error) {
	return m.getFn(ctx, id)
}
func (m *mockCommentService) Update(ctx context.Context, id string, input service.CommentUpdateInput) (model.Comment, error) {
	return m.updateFn(ctx, id, input)
}
func (m *mockCommentService) Delete(ctx context.Context, id string) error { return m.deleteFn(ctx, id) }

func TestCommentHandlerList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createdAt := time.Date(2026, 3, 6, 11, 0, 0, 0, time.UTC)
	h := NewCommentHandler(&mockCommentService{listFn: func(ctx context.Context, filter service.CommentListFilter) ([]model.Comment, int64, error) {
		if filter.TaskID != "3" || filter.Author != "Bob" {
			t.Fatalf("unexpected filter: %+v", filter)
		}
		return []model.Comment{{ID: 1, TaskID: 3, Author: "Bob", Text: "Hi", CreatedAt: createdAt}}, 1, nil
	}})
	r := gin.New()
	r.GET("/comments", h.List)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/comments?taskId=3&author=Bob", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCommentHandlerCreateBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCommentHandler(&mockCommentService{})
	r := gin.New()
	r.POST("/comments", h.Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBufferString(`{"taskId":0,"author":"","text":""}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCommentHandlerCreateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCommentHandler(&mockCommentService{createFn: func(ctx context.Context, input service.CommentCreateInput) (model.Comment, error) {
		if input.TaskID != 2 || input.Author != "Ann" || input.Text != "hello" {
			t.Fatalf("unexpected input: %+v", input)
		}
		return model.Comment{ID: 5, TaskID: input.TaskID, Author: input.Author, Text: input.Text}, nil
	}})
	r := gin.New()
	r.POST("/comments", h.Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBufferString(`{"taskId":2,"author":"Ann","text":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestCommentHandlerGetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCommentHandler(&mockCommentService{getFn: func(ctx context.Context, id string) (model.Comment, error) {
		return model.Comment{}, gorm.ErrRecordNotFound
	}})
	r := gin.New()
	r.GET("/comments/:id", h.Get)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/comments/9", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCommentHandlerGetSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCommentHandler(&mockCommentService{getFn: func(ctx context.Context, id string) (model.Comment, error) {
		if id != "9" {
			t.Fatalf("id = %s", id)
		}
		return model.Comment{ID: 9, TaskID: 1, Author: "Bob", Text: "hello"}, nil
	}})
	r := gin.New()
	r.GET("/comments/:id", h.Get)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/comments/9", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCommentHandlerUpdateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCommentHandler(&mockCommentService{updateFn: func(ctx context.Context, id string, input service.CommentUpdateInput) (model.Comment, error) {
		if id != "4" || input.Text == nil || *input.Text != "Updated" {
			t.Fatalf("unexpected input: id=%s input=%+v", id, input)
		}
		return model.Comment{ID: 4, TaskID: 1, Author: "Joe", Text: *input.Text}, nil
	}})
	r := gin.New()
	r.PUT("/comments/:id", h.Update)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/comments/4", bytes.NewBufferString(`{"text":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestCommentHandlerUpdateBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCommentHandler(&mockCommentService{})
	r := gin.New()

	r.PUT("/comments/:id", h.Update)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/comments/4", bytes.NewBufferString(`{"text":123}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCommentHandlerUpdateNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentHandler(&mockCommentService{
		updateFn: func(ctx context.Context, id string, input service.CommentUpdateInput) (model.Comment, error) {
			if id != "4" {
				t.Fatalf("unexpected id: %s", id)
			}
			return model.Comment{}, gorm.ErrRecordNotFound
		},
	})

	r := gin.New()
	r.PUT("/comments/:id", h.Update)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPut,
		"/comments/4",
		bytes.NewBufferString(`{"text":"Updated"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestCommentHandlerUpdateInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCommentHandler(&mockCommentService{updateFn: func(ctx context.Context, id string, input service.CommentUpdateInput) (model.Comment, error) {
		return model.Comment{}, errors.New("boom")
	}})
	r := gin.New()
	r.PUT("/comments/:id", h.Update)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/comments/4", bytes.NewBufferString(`{"text":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestCommentHandlerDeleteInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCommentHandler(&mockCommentService{deleteFn: func(ctx context.Context, id string) error {
		return errors.New("delete failed")
	}})
	r := gin.New()
	r.DELETE("/comments/:id", h.Delete)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/comments/4", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	var apiErr httpx.APIError
	if err := json.Unmarshal(w.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if apiErr.Message != "delete failed" {
		t.Fatalf("message = %q", apiErr.Message)
	}
}

func TestCommentHandlerDeleteNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCommentHandler(&mockCommentService{deleteFn: func(ctx context.Context, id string) error {
		return nil
	}})
	r := gin.New()
	r.DELETE("/comments/:id", h.Delete)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/comments/4", nil))

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}
