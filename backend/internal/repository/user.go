package repository

import (
	"context"

	"project-management/internal/model"
	"project-management/internal/service"

	"gorm.io/gorm"
)

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) service.UserRepository { return UserRepository{db: db} }

func (r UserRepository) List(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).Model(&model.User{}).Find(&users).Error
	return users, err
}
