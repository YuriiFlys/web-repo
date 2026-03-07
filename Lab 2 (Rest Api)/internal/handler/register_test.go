package handler

import (
	"context"
	"sort"
	"testing"

	"project-management/internal/model"
	"project-management/internal/service"

	"github.com/gin-gonic/gin"
)

type routeAuthService struct{}

func (routeAuthService) Register(ctx context.Context, input service.RegisterInput) (model.User, string, error) {
	panic("not used")
}
func (routeAuthService) Login(ctx context.Context, input service.LoginInput) (model.User, string, error) {
	panic("not used")
}
func (routeAuthService) GetByID(ctx context.Context, id uint) (model.User, error) { panic("not used") }

type routeProjectService struct{}

func (routeProjectService) List(ctx context.Context, filter service.ProjectListFilter) ([]model.Project, int64, error) {
	panic("not used")
}
func (routeProjectService) Create(ctx context.Context, input service.ProjectCreateInput) (model.Project, error) {
	panic("not used")
}
func (routeProjectService) Get(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
	panic("not used")
}
func (routeProjectService) Update(ctx context.Context, id string, input service.ProjectUpdateInput) (model.Project, error) {
	panic("not used")
}
func (routeProjectService) Delete(ctx context.Context, id string) error { panic("not used") }
func (routeProjectService) ListTasks(ctx context.Context, projectID uint, filter service.ProjectTaskListFilter) ([]model.Task, int64, error) {
	panic("not used")
}
func (routeProjectService) CreateTask(ctx context.Context, input service.ProjectTaskCreateInput) (model.Task, error) {
	panic("not used")
}

type routeTaskService struct{}

func (routeTaskService) List(ctx context.Context, filter service.TaskListFilter) ([]model.Task, int64, error) {
	panic("not used")
}
func (routeTaskService) Create(ctx context.Context, input service.TaskCreateInput) (model.Task, error) {
	panic("not used")
}
func (routeTaskService) Get(ctx context.Context, id string, includeComments bool) (model.Task, error) {
	panic("not used")
}
func (routeTaskService) Update(ctx context.Context, id string, input service.TaskUpdateInput) (model.Task, error) {
	panic("not used")
}
func (routeTaskService) Delete(ctx context.Context, id string) error { panic("not used") }
func (routeTaskService) ListComments(ctx context.Context, taskID string, filter service.TaskCommentListFilter) ([]model.Comment, int64, error) {
	panic("not used")
}
func (routeTaskService) CreateComment(ctx context.Context, input service.TaskCommentCreateInput) (model.Comment, error) {
	panic("not used")
}

type routeCommentService struct{}

func (routeCommentService) List(ctx context.Context, filter service.CommentListFilter) ([]model.Comment, int64, error) {
	panic("not used")
}
func (routeCommentService) Create(ctx context.Context, input service.CommentCreateInput) (model.Comment, error) {
	panic("not used")
}
func (routeCommentService) Get(ctx context.Context, id string) (model.Comment, error) {
	panic("not used")
}
func (routeCommentService) Update(ctx context.Context, id string, input service.CommentUpdateInput) (model.Comment, error) {
	panic("not used")
}
func (routeCommentService) Delete(ctx context.Context, id string) error { panic("not used") }

type routeUserService struct{}

func (routeUserService) List(ctx context.Context) ([]model.User, error) { panic("not used") }

func TestHandlerRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api")

	NewAuthHandler(routeAuthService{}).Register(api)
	NewAuthHandler(routeAuthService{}).RegisterProtected(api)
	NewUserHandler(routeUserService{}).Register(api)
	NewProjectHandler(routeProjectService{}).Register(api)
	NewTaskHandler(routeTaskService{}).Register(api)
	NewCommentHandler(routeCommentService{}).Register(api)

	got := make([]string, 0, len(r.Routes()))
	for _, route := range r.Routes() {
		got = append(got, route.Method+" "+route.Path)
	}
	sort.Strings(got)

	want := []string{
		"DELETE /api/comments/:id",
		"DELETE /api/projects/:id",
		"DELETE /api/tasks/:id",
		"GET /api/auth/me",
		"GET /api/comments",
		"GET /api/comments/:id",
		"GET /api/projects",
		"GET /api/projects/:id",
		"GET /api/projects/:id/tasks",
		"GET /api/tasks",
		"GET /api/tasks/:id",
		"GET /api/tasks/:id/comments",
		"GET /api/users",
		"POST /api/auth/login",
		"POST /api/auth/register",
		"POST /api/comments",
		"POST /api/projects",
		"POST /api/projects/:id/tasks",
		"POST /api/tasks",
		"POST /api/tasks/:id/comments",
		"PUT /api/comments/:id",
		"PUT /api/projects/:id",
		"PUT /api/tasks/:id",
	}
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("route count = %d, want %d\nroutes: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("route[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
