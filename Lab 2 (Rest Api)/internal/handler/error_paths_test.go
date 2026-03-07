package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"project-management/internal/httpx"
	"project-management/internal/model"
	"project-management/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestAuthHandlerAdditionalErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("register internal error", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthService{registerFn: func(ctx context.Context, input service.RegisterInput) (model.User, string, error) {
			return model.User{}, "", errors.New("db down")
		}})
		r := gin.New()
		r.POST("/auth/register", h.RegisterUser)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{"email":"user@example.com","password":"secret1","name":"Alice"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "db down")
	})

	t.Run("login user not found", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthService{loginFn: func(ctx context.Context, input service.LoginInput) (model.User, string, error) {
			return model.User{}, "", gorm.ErrRecordNotFound
		}})
		r := gin.New()
		r.POST("/auth/login", h.Login)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{"email":"user@example.com","password":"secret1"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assertAPIError(t, w, http.StatusUnauthorized, httpx.CodeUnauthorized, "invalid credentials")
	})

	t.Run("me invalid context type", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthService{})
		r := gin.New()
		r.GET("/auth/me", func(c *gin.Context) { c.Set("userID", "bad"); h.Me(c) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/me", nil))
		assertAPIError(t, w, http.StatusUnauthorized, httpx.CodeUnauthorized, "unauthorized")
	})

	t.Run("me service error", func(t *testing.T) {
		h := NewAuthHandler(&mockAuthService{getByIDFn: func(ctx context.Context, id uint) (model.User, error) {
			return model.User{}, errors.New("boom")
		}})
		r := gin.New()
		r.GET("/auth/me", func(c *gin.Context) { c.Set("userID", uint(2)); h.Me(c) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/me", nil))
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})
}

func TestProjectHandlerAdditionalErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list internal error", func(t *testing.T) {
		h := NewProjectHandler(&mockProjectService{listFn: func(ctx context.Context, filter service.ProjectListFilter) ([]model.Project, int64, error) {
			return nil, 0, errors.New("boom")
		}})
		r := gin.New()
		r.GET("/projects", h.List)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/projects", nil))
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})

	t.Run("create internal error", func(t *testing.T) {
		h := NewProjectHandler(&mockProjectService{createFn: func(ctx context.Context, input service.ProjectCreateInput) (model.Project, error) {
			return model.Project{}, errors.New("insert failed")
		}})
		r := gin.New()
		r.POST("/projects", h.Create)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"title":"A","status":"active"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "insert failed")
	})

	t.Run("get internal error", func(t *testing.T) {
		h := NewProjectHandler(&mockProjectService{getFn: func(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
			return model.Project{}, errors.New("boom")
		}})
		r := gin.New()
		r.GET("/projects/:id", h.Get)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/projects/1", nil))
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})

	t.Run("create project task invalid body", func(t *testing.T) {
		h := NewProjectHandler(&mockProjectService{})
		r := gin.New()
		r.POST("/projects/:id/tasks", h.CreateProjectTask)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects/1/tasks", bytes.NewBufferString(`{"title":"","status":"bad"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", w.Code)
		}
	})

	t.Run("create project task internal error", func(t *testing.T) {
		h := NewProjectHandler(&mockProjectService{createTaskFn: func(ctx context.Context, input service.ProjectTaskCreateInput) (model.Task, error) {
			return model.Task{}, errors.New("boom")
		}})
		r := gin.New()
		r.POST("/projects/:id/tasks", h.CreateProjectTask)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects/1/tasks", bytes.NewBufferString(`{"title":"Task","status":"todo"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})
}

func TestTaskHandlerAdditionalErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list internal error", func(t *testing.T) {
		h := NewTaskHandler(&mockTaskService{listFn: func(ctx context.Context, filter service.TaskListFilter) ([]model.Task, int64, error) {
			return nil, 0, errors.New("boom")
		}})
		r := gin.New()
		r.GET("/tasks", h.List)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tasks", nil))
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})

	t.Run("get internal error", func(t *testing.T) {
		h := NewTaskHandler(&mockTaskService{getFn: func(ctx context.Context, id string, includeComments bool) (model.Task, error) {
			return model.Task{}, errors.New("boom")
		}})
		r := gin.New()
		r.GET("/tasks/:id", h.Get)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tasks/1", nil))
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})

	t.Run("update invalid body", func(t *testing.T) {
		h := NewTaskHandler(&mockTaskService{})
		r := gin.New()
		r.PUT("/tasks/:id", h.Update)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/tasks/1", bytes.NewBufferString(`{"status":"bad"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", w.Code)
		}
	})

	t.Run("list comments internal error", func(t *testing.T) {
		h := NewTaskHandler(&mockTaskService{listCommentsFn: func(ctx context.Context, taskID string, filter service.TaskCommentListFilter) ([]model.Comment, int64, error) {
			return nil, 0, errors.New("boom")
		}})
		r := gin.New()
		r.GET("/tasks/:id/comments", h.ListTaskComments)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/tasks/3/comments", nil))
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})

	t.Run("create task comment bad body", func(t *testing.T) {
		h := NewTaskHandler(&mockTaskService{})
		r := gin.New()
		r.POST("/tasks/:id/comments", h.CreateTaskComment)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/tasks/3/comments", bytes.NewBufferString(`{"author":""}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d", w.Code)
		}
	})

	t.Run("create task comment internal error", func(t *testing.T) {
		h := NewTaskHandler(&mockTaskService{createCommentFn: func(ctx context.Context, input service.TaskCommentCreateInput) (model.Comment, error) {
			return model.Comment{}, errors.New("boom")
		}})
		r := gin.New()
		r.POST("/tasks/:id/comments", h.CreateTaskComment)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/tasks/3/comments", bytes.NewBufferString(`{"author":"Ann","text":"hi"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})

	t.Run("must uint invalid", func(t *testing.T) {
		if got := mustUint("abc"); got != 0 {
			t.Fatalf("mustUint = %d", got)
		}
	})
}

func TestCommentHandlerAdditionalErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list internal error", func(t *testing.T) {
		h := NewCommentHandler(&mockCommentService{listFn: func(ctx context.Context, filter service.CommentListFilter) ([]model.Comment, int64, error) {
			return nil, 0, errors.New("boom")
		}})
		r := gin.New()
		r.GET("/comments", h.List)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/comments", nil))
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})

	t.Run("create internal error", func(t *testing.T) {
		h := NewCommentHandler(&mockCommentService{createFn: func(ctx context.Context, input service.CommentCreateInput) (model.Comment, error) {
			return model.Comment{}, errors.New("boom")
		}})
		r := gin.New()
		r.POST("/comments", h.Create)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBufferString(`{"taskId":1,"author":"Ann","text":"hi"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})

	t.Run("get internal error", func(t *testing.T) {
		h := NewCommentHandler(&mockCommentService{getFn: func(ctx context.Context, id string) (model.Comment, error) {
			return model.Comment{}, errors.New("boom")
		}})
		r := gin.New()
		r.GET("/comments/:id", h.Get)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/comments/1", nil))
		assertAPIError(t, w, http.StatusInternalServerError, "INTERNAL", "boom")
	})
}

func assertAPIError(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, wantCode, wantMsg string) {
	t.Helper()
	if w.Code != wantStatus {
		t.Fatalf("status = %d, want %d", w.Code, wantStatus)
	}
	var resp httpx.APIError
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Code != wantCode || resp.Message != wantMsg {
		t.Fatalf("resp = %+v, want code=%q msg=%q", resp, wantCode, wantMsg)
	}
}
