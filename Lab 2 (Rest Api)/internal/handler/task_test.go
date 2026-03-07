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

type mockTaskService struct {
	listFn          func(ctx context.Context, filter service.TaskListFilter) ([]model.Task, int64, error)
	createFn        func(ctx context.Context, input service.TaskCreateInput) (model.Task, error)
	getFn           func(ctx context.Context, id string, includeComments bool) (model.Task, error)
	updateFn        func(ctx context.Context, id string, input service.TaskUpdateInput) (model.Task, error)
	deleteFn        func(ctx context.Context, id string) error
	listCommentsFn  func(ctx context.Context, taskID string, filter service.TaskCommentListFilter) ([]model.Comment, int64, error)
	createCommentFn func(ctx context.Context, input service.TaskCommentCreateInput) (model.Comment, error)
}

func (m *mockTaskService) List(ctx context.Context, filter service.TaskListFilter) ([]model.Task, int64, error) {
	return m.listFn(ctx, filter)
}
func (m *mockTaskService) Create(ctx context.Context, input service.TaskCreateInput) (model.Task, error) {
	return m.createFn(ctx, input)
}
func (m *mockTaskService) Get(ctx context.Context, id string, includeComments bool) (model.Task, error) {
	return m.getFn(ctx, id, includeComments)
}
func (m *mockTaskService) Update(ctx context.Context, id string, input service.TaskUpdateInput) (model.Task, error) {
	return m.updateFn(ctx, id, input)
}
func (m *mockTaskService) Delete(ctx context.Context, id string) error { return m.deleteFn(ctx, id) }
func (m *mockTaskService) ListComments(ctx context.Context, taskID string, filter service.TaskCommentListFilter) ([]model.Comment, int64, error) {
	return m.listCommentsFn(ctx, taskID, filter)
}
func (m *mockTaskService) CreateComment(ctx context.Context, input service.TaskCommentCreateInput) (model.Comment, error) {
	return m.createCommentFn(ctx, input)
}

func TestTaskHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("bad request", func(t *testing.T) {
		h := NewTaskHandler(&mockTaskService{})
		r := gin.New()
		r.POST("/tasks", h.Create)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{"projectId":0,"title":"","status":"broken"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("success", func(t *testing.T) {
		h := NewTaskHandler(&mockTaskService{createFn: func(ctx context.Context, input service.TaskCreateInput) (model.Task, error) {
			if input.ProjectID != 2 || input.Title != "Implement" {
				t.Fatalf("unexpected input: %+v", input)
			}
			return model.Task{ID: 9, ProjectID: input.ProjectID, Title: input.Title, Status: input.Status}, nil
		}})
		r := gin.New()
		r.POST("/tasks", h.Create)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{"projectId":2,"title":"Implement","description":"desc","status":"todo"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		h := NewTaskHandler(&mockTaskService{createFn: func(ctx context.Context, input service.TaskCreateInput) (model.Task, error) {
			return model.Task{}, errors.New("boom")
		}})
		r := gin.New()
		r.POST("/tasks", h.Create)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{"projectId":2,"title":"Implement","description":"desc","status":"todo"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestTaskHandlerGetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&mockTaskService{getFn: func(ctx context.Context, id string, includeComments bool) (model.Task, error) {
		return model.Task{}, gorm.ErrRecordNotFound
	}})
	r := gin.New()
	r.GET("/tasks/:id", h.Get)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tasks/123", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestTaskHandlerGetSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&mockTaskService{getFn: func(ctx context.Context, id string, includeComments bool) (model.Task, error) {
		if id != "123" || !includeComments {
			t.Fatalf("unexpected get: id=%s includeComments=%v", id, includeComments)
		}
		return model.Task{ID: 123, Title: "Task", Status: model.TaskTodo}, nil
	}})
	r := gin.New()
	r.GET("/tasks/:id", h.Get)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tasks/123?include=comments", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestTaskHandlerUpdateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	status := model.TaskDone
	h := NewTaskHandler(&mockTaskService{updateFn: func(ctx context.Context, id string, input service.TaskUpdateInput) (model.Task, error) {
		if id != "6" || input.Status == nil || *input.Status != model.TaskDone {
			t.Fatalf("unexpected input: id=%s input=%+v", id, input)
		}
		return model.Task{ID: 6, Title: "Done", Status: *input.Status}, nil
	}})
	r := gin.New()
	r.PUT("/tasks/:id", h.Update)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/tasks/6", bytes.NewBufferString(`{"status":"done"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Status != status {
		t.Fatalf("status = %q, want %q", resp.Status, status)
	}
}

func TestTaskHandlerUpdateInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&mockTaskService{updateFn: func(ctx context.Context, id string, input service.TaskUpdateInput) (model.Task, error) {
		return model.Task{}, errors.New("boom")
	}})
	r := gin.New()
	r.PUT("/tasks/:id", h.Update)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/tasks/6", bytes.NewBufferString(`{"status":"done"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestTaskHandlerUpdateNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&mockTaskService{updateFn: func(ctx context.Context, id string, input service.TaskUpdateInput) (model.Task, error) {
		if id != "6" {
			t.Fatalf("id = %s", id)
		}
		return model.Task{}, gorm.ErrRecordNotFound
	}})
	r := gin.New()
	r.PUT("/tasks/:id", h.Update)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/tasks/6", bytes.NewBufferString(`{"status":"done"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestTaskHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&mockTaskService{listFn: func(ctx context.Context, filter service.TaskListFilter) ([]model.Task, int64, error) {
		if filter.ProjectID != "2" || !filter.IncludeComments {
			t.Fatalf("unexpected filter: %+v", filter)
		}
		return []model.Task{{ID: 1, ProjectID: 2, Title: "Implement", Status: model.TaskTodo}}, 1, nil
	}})
	r := gin.New()
	r.GET("/tasks", h.List)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tasks?projectId=2&include=comments", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestTaskHandlerListTaskComments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createdAt := time.Date(2026, 3, 6, 10, 0, 0, 0, time.UTC)
	h := NewTaskHandler(&mockTaskService{listCommentsFn: func(ctx context.Context, taskID string, filter service.TaskCommentListFilter) ([]model.Comment, int64, error) {
		if taskID != "8" || filter.Author != "Alice" {
			t.Fatalf("unexpected filter: taskID=%s filter=%+v", taskID, filter)
		}
		return []model.Comment{{ID: 1, TaskID: 8, Author: "Alice", Text: "Looks good", CreatedAt: createdAt}}, 1, nil
	}})
	r := gin.New()
	r.GET("/tasks/:id/comments", h.ListTaskComments)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tasks/8/comments?author=Alice", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestTaskHandlerCreateTaskComment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&mockTaskService{createCommentFn: func(ctx context.Context, input service.TaskCommentCreateInput) (model.Comment, error) {
		if input.TaskID != 11 || input.Author != "Ann" {
			t.Fatalf("unexpected input: %+v", input)
		}
		return model.Comment{ID: 4, TaskID: input.TaskID, Author: input.Author, Text: input.Text}, nil
	}})
	r := gin.New()
	r.POST("/tasks/:id/comments", h.CreateTaskComment)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks/11/comments", bytes.NewBufferString(`{"author":"Ann","text":"Ship it"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestTaskHandlerDeleteInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&mockTaskService{deleteFn: func(ctx context.Context, id string) error {
		return errors.New("delete failed")
	}})
	r := gin.New()
	r.DELETE("/tasks/:id", h.Delete)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/tasks/5", nil))

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

func TestTaskHandlerDeleteNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTaskHandler(&mockTaskService{deleteFn: func(ctx context.Context, id string) error {
		return nil
	}})
	r := gin.New()
	r.DELETE("/tasks/:id", h.Delete)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/tasks/5", nil))

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}
