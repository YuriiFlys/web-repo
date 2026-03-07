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

type mockProjectService struct {
	listFn       func(ctx context.Context, filter service.ProjectListFilter) ([]model.Project, int64, error)
	createFn     func(ctx context.Context, input service.ProjectCreateInput) (model.Project, error)
	getFn        func(ctx context.Context, id string, includeTasks bool) (model.Project, error)
	updateFn     func(ctx context.Context, id string, input service.ProjectUpdateInput) (model.Project, error)
	deleteFn     func(ctx context.Context, id string) error
	listTasksFn  func(ctx context.Context, projectID uint, filter service.ProjectTaskListFilter) ([]model.Task, int64, error)
	createTaskFn func(ctx context.Context, input service.ProjectTaskCreateInput) (model.Task, error)
}

func (m *mockProjectService) List(ctx context.Context, filter service.ProjectListFilter) ([]model.Project, int64, error) {
	return m.listFn(ctx, filter)
}
func (m *mockProjectService) Create(ctx context.Context, input service.ProjectCreateInput) (model.Project, error) {
	return m.createFn(ctx, input)
}
func (m *mockProjectService) Get(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
	return m.getFn(ctx, id, includeTasks)
}
func (m *mockProjectService) Update(ctx context.Context, id string, input service.ProjectUpdateInput) (model.Project, error) {
	return m.updateFn(ctx, id, input)
}
func (m *mockProjectService) Delete(ctx context.Context, id string) error { return m.deleteFn(ctx, id) }
func (m *mockProjectService) ListTasks(ctx context.Context, projectID uint, filter service.ProjectTaskListFilter) ([]model.Task, int64, error) {
	return m.listTasksFn(ctx, projectID, filter)
}
func (m *mockProjectService) CreateTask(ctx context.Context, input service.ProjectTaskCreateInput) (model.Task, error) {
	return m.createTaskFn(ctx, input)
}

func TestProjectHandlerList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{listFn: func(ctx context.Context, filter service.ProjectListFilter) ([]model.Project, int64, error) {
		if filter.Params.Page != 2 || filter.Params.PageSize != 5 || len(filter.Params.Sort) != 1 || filter.Params.Sort[0].Field != "title" || !filter.IncludeTasks {
			t.Fatalf("unexpected filter: %+v", filter)
		}
		return []model.Project{{ID: 1, Title: "API", Status: model.ProjectActive}}, 11, nil
	}})
	r := gin.New()
	r.GET("/projects", h.List)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects?page=2&pageSize=5&sort=-title&status=active&q=api&include=tasks", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp struct {
		Page     int             `json:"page"`
		PageSize int             `json:"pageSize"`
		Items    []model.Project `json:"items"`
		IsLast   bool            `json:"isLast"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Page != 2 || resp.PageSize != 5 || resp.IsLast {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestProjectHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("bad request", func(t *testing.T) {
		h := NewProjectHandler(&mockProjectService{})
		r := gin.New()
		r.POST("/projects", h.Create)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"title":"","status":"broken"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("success", func(t *testing.T) {
		h := NewProjectHandler(&mockProjectService{createFn: func(ctx context.Context, input service.ProjectCreateInput) (model.Project, error) {
			if input.Title != "Project X" || input.Status != model.ProjectActive {
				t.Fatalf("unexpected input: %+v", input)
			}
			return model.Project{ID: 10, Title: input.Title, Status: input.Status}, nil
		}})
		r := gin.New()
		r.POST("/projects", h.Create)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"title":"Project X","description":"desc","status":"active"}`))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
		}
	})
}

func TestProjectHandlerGetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{getFn: func(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
		return model.Project{}, gorm.ErrRecordNotFound
	}})
	r := gin.New()
	r.GET("/projects/:id", h.Get)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/projects/12", nil))

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestProjectHandlerGetSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{getFn: func(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
		if id != "12" || !includeTasks {
			t.Fatalf("unexpected get: id=%s includeTasks=%v", id, includeTasks)
		}
		return model.Project{ID: 12, Title: "Roadmap", Status: model.ProjectActive}, nil
	}})
	r := gin.New()
	r.GET("/projects/:id", h.Get)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/projects/12?include=tasks", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestProjectHandlerUpdateBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{getFn: func(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
		return model.Project{}, gorm.ErrRecordNotFound
	}})
	r := gin.New()
	r.PUT("/projects/:id", h.Update)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/projects/12", bytes.NewBufferString(`{"title":"","status":"broken"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestProjectHandlerUpdateNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewProjectHandler(&mockProjectService{
		updateFn: func(ctx context.Context, id string, input service.ProjectUpdateInput) (model.Project, error) {
			if id != "12" {
				t.Fatalf("unexpected id: %s", id)
			}
			return model.Project{}, gorm.ErrRecordNotFound
		},
	})

	r := gin.New()
	r.PUT("/projects/:id", h.Update)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/projects/12",
		bytes.NewBufferString(`{"title":"Updated","status":"archived"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestProjectHandlerUpdateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{updateFn: func(ctx context.Context, id string, input service.ProjectUpdateInput) (model.Project, error) {
		if id != "4" || input.Title == nil || *input.Title != "Updated" {
			t.Fatalf("unexpected update: id=%s input=%+v", id, input)
		}
		return model.Project{ID: 4, Title: *input.Title, Status: model.ProjectArchived}, nil
	}})
	r := gin.New()
	r.PUT("/projects/:id", h.Update)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/projects/4",
		bytes.NewBufferString(`{"title":"Updated","status":"archived"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestProjectHandlerUpdateInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{updateFn: func(ctx context.Context, id string, input service.ProjectUpdateInput) (model.Project, error) {
		return model.Project{}, errors.New("boom")
	}})
	r := gin.New()
	r.PUT("/projects/:id", h.Update)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/projects/4", bytes.NewBufferString(`{"title":"Updated","status":"archived"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestProjectHandlerListProjectTasksInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{listTasksFn: func(ctx context.Context, projectID uint, filter service.ProjectTaskListFilter) ([]model.Task, int64, error) {
		return nil, 0, errors.New("boom")
	}})
	r := gin.New()
	r.GET("/projects/:id/tasks", h.ListProjectTasks)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/projects/3/tasks", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestProjectHandlerListProjectTasks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dueDate := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)

	t.Run("invalid project id", func(t *testing.T) {
		h := NewProjectHandler(&mockProjectService{})
		r := gin.New()
		r.GET("/projects/:id/tasks", h.ListProjectTasks)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet,
			"/projects/nope/tasks", nil))

		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("success", func(t *testing.T) {
		serviceMock := &mockProjectService{listTasksFn: func(ctx context.Context, projectID uint, filter service.ProjectTaskListFilter) ([]model.Task, int64, error) {
			if projectID != 3 || filter.AssigneeID != "8" {
				t.Fatalf("unexpected filter: projectID=%d filter=%+v", projectID, filter)
			}
			return []model.Task{{ID: 1, ProjectID: 3, Title: "Task", Status: model.TaskTodo, DueDate: &dueDate}}, 1, nil
		}}
		h := NewProjectHandler(serviceMock)
		r := gin.New()
		r.GET("/projects/:id/tasks", h.ListProjectTasks)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/projects/3/tasks?assigneeId=8", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestProjectHandlerCreateProjectTask(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{createTaskFn: func(ctx context.Context, input service.ProjectTaskCreateInput) (model.Task, error) {
		if input.ProjectID != 5 || input.Title != "Task A" {
			t.Fatalf("unexpected input: %+v", input)
		}
		return model.Task{ID: 2, ProjectID: input.ProjectID, Title: input.Title, Status: input.Status}, nil
	}})
	r := gin.New()
	r.POST("/projects/:id/tasks", h.CreateProjectTask)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/projects/5/tasks",
		bytes.NewBufferString(`{"title":"Task A","description":"desc","status":"todo"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestProjectHandlerCreateProjectTaskInvalidProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{})
	r := gin.New()
	r.POST("/projects/:id/tasks", h.CreateProjectTask)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/projects/nope/tasks", bytes.NewBufferString(`{"title":"Task A","status":"todo"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestProjectHandlerDeleteInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{deleteFn: func(ctx context.Context, id string) error {
		return errors.New("delete failed")
	}})
	r := gin.New()
	r.DELETE("/projects/:id", h.Delete)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/projects/1", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	var apiErr httpx.APIError
	if err := json.Unmarshal(w.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if apiErr.Message != "delete failed" {
		t.Fatalf("message = %q", apiErr.Message)
	}
}

func TestProjectHandlerDeleteNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewProjectHandler(&mockProjectService{deleteFn: func(ctx context.Context, id string) error {
		return nil
	}})
	r := gin.New()
	r.DELETE("/projects/:id", h.Delete)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/projects/1", nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}
