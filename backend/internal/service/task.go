package service

import (
	"context"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"
)

type TaskListFilter struct {
	Params          httpx.ListParams
	ProjectID       string
	Status          string
	AssigneeID      string
	DueFrom         string
	DueTo           string
	IncludeComments bool
}

type TaskCreateInput struct {
	ProjectID   uint
	Title       string
	Description string
	Status      model.TaskStatus
	AssigneeID  *uint
	DueDate     *time.Time
}

type TaskUpdateInput struct {
	Title       *string
	Description *string
	Status      *model.TaskStatus
	AssigneeID  **uint
	DueDate     **time.Time
}

type TaskCommentListFilter struct {
	Params httpx.ListParams
	Author string
}

type TaskCommentCreateInput struct {
	TaskID uint
	Author string
	Text   string
}

type TaskService interface {
	List(ctx context.Context, filter TaskListFilter) ([]model.Task, int64, error)
	Create(ctx context.Context, input TaskCreateInput) (model.Task, error)
	Get(ctx context.Context, id string, includeComments bool) (model.Task, error)
	Update(ctx context.Context, id string, input TaskUpdateInput) (model.Task, error)
	Delete(ctx context.Context, id string) error
	ListComments(ctx context.Context, taskID string, filter TaskCommentListFilter) ([]model.Comment, int64, error)
	CreateComment(ctx context.Context, input TaskCommentCreateInput) (model.Comment, error)
}

type TaskRepository interface {
	List(ctx context.Context, filter TaskListFilter) ([]model.Task, int64, error)
	Create(ctx context.Context, task *model.Task) error
	Get(ctx context.Context, id string, includeComments bool) (model.Task, error)
	Save(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id string) error
	ListComments(ctx context.Context, taskID string, filter TaskCommentListFilter) ([]model.Comment, int64, error)
	CreateComment(ctx context.Context, comment *model.Comment) error
}

type taskService struct{ repo TaskRepository }

func NewTaskService(repo TaskRepository) TaskService { return &taskService{repo: repo} }
func (s *taskService) List(ctx context.Context, filter TaskListFilter) ([]model.Task, int64, error) {
	return s.repo.List(ctx, filter)
}
func (s *taskService) Create(ctx context.Context, input TaskCreateInput) (model.Task, error) {
	task := model.Task{ProjectID: input.ProjectID, Title: input.Title, Description: input.Description, Status: input.Status, AssigneeID: input.AssigneeID, DueDate: input.DueDate}
	return task, s.repo.Create(ctx, &task)
}
func (s *taskService) Get(ctx context.Context, id string, includeComments bool) (model.Task, error) {
	return s.repo.Get(ctx, id, includeComments)
}
func (s *taskService) Update(ctx context.Context, id string, input TaskUpdateInput) (model.Task, error) {
	task, err := s.repo.Get(ctx, id, false)
	if err != nil {
		return model.Task{}, err
	}
	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Status != nil {
		task.Status = *input.Status
	}
	if input.AssigneeID != nil {
		task.AssigneeID = *input.AssigneeID
	}
	if input.DueDate != nil {
		task.DueDate = *input.DueDate
	}
	return task, s.repo.Save(ctx, &task)
}
func (s *taskService) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
func (s *taskService) ListComments(ctx context.Context, taskID string, filter TaskCommentListFilter) ([]model.Comment, int64, error) {
	return s.repo.ListComments(ctx, taskID, filter)
}
func (s *taskService) CreateComment(ctx context.Context, input TaskCommentCreateInput) (model.Comment, error) {
	comment := model.Comment{TaskID: input.TaskID, Author: input.Author, Text: input.Text}
	return comment, s.repo.CreateComment(ctx, &comment)
}
