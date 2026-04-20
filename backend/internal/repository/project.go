package repository

import (
	"context"

	"project-management/internal/httpx"
	"project-management/internal/model"
	"project-management/internal/service"

	"gorm.io/gorm"
)

type ProjectRepository struct{ db *gorm.DB }

func NewProjectRepository(db *gorm.DB) service.ProjectRepository { return ProjectRepository{db: db} }

func (r ProjectRepository) List(ctx context.Context, filter service.ProjectListFilter) ([]model.Project, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.Project{})
	if filter.Query != "" {
		like := "%" + filter.Query + "%"
		db = db.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if filter.IncludeTasks {
		db = db.Preload("Tasks")
	}
	allowedSort := map[string]string{"id": "id", "title": "title", "status": "status", "createdAt": "created_at"}
	var items []model.Project
	err := httpx.ApplyPagination(httpx.ApplySorting(db, allowedSort, filter.Params, "created_at DESC"), filter.Params).Find(&items).Error
	return items, total, err
}

func (r ProjectRepository) Create(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r ProjectRepository) Get(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
	var project model.Project
	db := r.db.WithContext(ctx)
	if includeTasks {
		db = db.Preload("Tasks")
	}
	err := db.First(&project, id).Error
	return project, err
}

func (r ProjectRepository) Save(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r ProjectRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Project{}, id).Error
}

func (r ProjectRepository) ListTasks(ctx context.Context, projectID uint, filter service.ProjectTaskListFilter) ([]model.Task, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.Task{}).Where("project_id = ?", projectID)
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.AssigneeID != "" {
		db = db.Where("assignee_id = ?", filter.AssigneeID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	allowedSort := map[string]string{"id": "id", "title": "title", "status": "status", "dueDate": "due_date", "createdAt": "created_at"}
	var items []model.Task
	err := httpx.ApplyPagination(httpx.ApplySorting(db, allowedSort, filter.Params, "created_at DESC"), filter.Params).Find(&items).Error
	return items, total, err
}

func (r ProjectRepository) CreateTask(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}
