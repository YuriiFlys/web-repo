package repository

import (
	"context"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"
	"project-management/internal/service"

	"gorm.io/gorm"
)

type TaskRepository struct{ db *gorm.DB }

func NewTaskRepository(db *gorm.DB) service.TaskRepository { return TaskRepository{db: db} }

func (r TaskRepository) List(ctx context.Context, filter service.TaskListFilter) ([]model.Task, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.Task{})
	if filter.ProjectID != "" {
		db = db.Where("project_id = ?", filter.ProjectID)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.AssigneeID != "" {
		db = db.Where("assignee_id = ?", filter.AssigneeID)
	}
	if filter.DueFrom != "" {
		if t, err := time.Parse("2006-01-02", filter.DueFrom); err == nil {
			db = db.Where("due_date >= ?", t)
		}
	}
	if filter.DueTo != "" {
		if t, err := time.Parse("2006-01-02", filter.DueTo); err == nil {
			db = db.Where("due_date <= ?", t)
		}
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if filter.IncludeComments {
		db = db.Preload("Comments")
	}
	allowedSort := map[string]string{"id": "id", "title": "title", "status": "status", "dueDate": "due_date", "createdAt": "created_at"}
	var items []model.Task
	err := httpx.ApplyPagination(httpx.ApplySorting(db, allowedSort, filter.Params, "created_at DESC"), filter.Params).Find(&items).Error
	return items, total, err
}

func (r TaskRepository) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r TaskRepository) Get(ctx context.Context, id string, includeComments bool) (model.Task, error) {
	var task model.Task
	db := r.db.WithContext(ctx)
	if includeComments {
		db = db.Preload("Comments")
	}
	err := db.First(&task, id).Error
	return task, err
}

func (r TaskRepository) Save(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r TaskRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Task{}, id).Error
}

func (r TaskRepository) ListComments(ctx context.Context, taskID string, filter service.TaskCommentListFilter) ([]model.Comment, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.Comment{}).Where("task_id = ?", taskID)
	if filter.Author != "" {
		db = db.Where("author = ?", filter.Author)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowedSort := map[string]string{"id": "id", "createdAt": "created_at"}
	var items []model.Comment
	err := httpx.ApplyPagination(httpx.ApplySorting(db, allowedSort, filter.Params, "created_at DESC"), filter.Params).Find(&items).Error
	return items, total, err
}

func (r TaskRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}
