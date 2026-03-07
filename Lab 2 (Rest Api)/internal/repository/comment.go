package repository

import (
	"context"

	"project-management/internal/httpx"
	"project-management/internal/model"
	"project-management/internal/service"

	"gorm.io/gorm"
)

type CommentRepository struct{ db *gorm.DB }

func NewCommentRepository(db *gorm.DB) service.CommentRepository { return CommentRepository{db: db} }

func (r CommentRepository) List(ctx context.Context, filter service.CommentListFilter) ([]model.Comment, int64, error) {
	db := r.db.WithContext(ctx).Model(&model.Comment{})
	if filter.TaskID != "" {
		db = db.Where("task_id = ?", filter.TaskID)
	}
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
func (r CommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}
func (r CommentRepository) Get(ctx context.Context, id string) (model.Comment, error) {
	var comment model.Comment
	err := r.db.WithContext(ctx).First(&comment, id).Error
	return comment, err
}
func (r CommentRepository) Save(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}
func (r CommentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Comment{}, id).Error
}
