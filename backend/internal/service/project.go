package service

import (
	"context"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"
)

type ProjectListFilter struct {
	Params       httpx.ListParams
	Query        string
	Status       string
	IncludeTasks bool
}

type ProjectCreateInput struct {
	Title       string
	Description string
	Status      model.ProjectStatus
}

type ProjectUpdateInput struct {
	Title       *string
	Description *string
	Status      *model.ProjectStatus
}

type ProjectTaskListFilter struct {
	Params     httpx.ListParams
	Status     string
	AssigneeID string
}

type ProjectTaskCreateInput struct {
	ProjectID   uint
	Title       string
	Description string
	Status      model.TaskStatus
	AssigneeID  *uint
	DueDate     *time.Time
}

type ProjectService interface {
	List(ctx context.Context, filter ProjectListFilter) ([]model.Project, int64, error)
	Create(ctx context.Context, input ProjectCreateInput) (model.Project, error)
	Get(ctx context.Context, id string, includeTasks bool) (model.Project, error)
	Update(ctx context.Context, id string, input ProjectUpdateInput) (model.Project, error)
	Delete(ctx context.Context, id string) error
	ListTasks(ctx context.Context, projectID uint, filter ProjectTaskListFilter) ([]model.Task, int64, error)
	CreateTask(ctx context.Context, input ProjectTaskCreateInput) (model.Task, error)
}

type ProjectRepository interface {
	List(ctx context.Context, filter ProjectListFilter) ([]model.Project, int64, error)
	Create(ctx context.Context, project *model.Project) error
	Get(ctx context.Context, id string, includeTasks bool) (model.Project, error)
	Save(ctx context.Context, project *model.Project) error
	Delete(ctx context.Context, id string) error
	ListTasks(ctx context.Context, projectID uint, filter ProjectTaskListFilter) ([]model.Task, int64, error)
	CreateTask(ctx context.Context, task *model.Task) error
}

type projectService struct{ repo ProjectRepository }

func NewProjectService(repo ProjectRepository) ProjectService { return &projectService{repo: repo} }

func (s *projectService) List(ctx context.Context, filter ProjectListFilter) ([]model.Project, int64, error) {
	return s.repo.List(ctx, filter)
}

func (s *projectService) Create(ctx context.Context, input ProjectCreateInput) (model.Project, error) {
	project := model.Project{Title: input.Title, Description: input.Description, Status: input.Status}
	return project, s.repo.Create(ctx, &project)
}

func (s *projectService) Get(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
	return s.repo.Get(ctx, id, includeTasks)
}

func (s *projectService) Update(ctx context.Context, id string, input ProjectUpdateInput) (model.Project, error) {
	project, err := s.repo.Get(ctx, id, false)
	if err != nil {
		return model.Project{}, err
	}
	if input.Title != nil {
		project.Title = *input.Title
	}
	if input.Description != nil {
		project.Description = *input.Description
	}
	if input.Status != nil {
		project.Status = *input.Status
	}
	return project, s.repo.Save(ctx, &project)
}

func (s *projectService) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }

func (s *projectService) ListTasks(ctx context.Context, projectID uint, filter ProjectTaskListFilter) ([]model.Task, int64, error) {
	return s.repo.ListTasks(ctx, projectID, filter)
}

func (s *projectService) CreateTask(ctx context.Context, input ProjectTaskCreateInput) (model.Task, error) {
	task := model.Task{ProjectID: input.ProjectID, Title: input.Title, Description: input.Description, Status: input.Status, AssigneeID: input.AssigneeID, DueDate: input.DueDate}
	return task, s.repo.CreateTask(ctx, &task)
}
