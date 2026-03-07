package repository

import (
	"context"

	"project-management/internal/model"
	"project-management/internal/service"

	"gorm.io/gorm"
)

type AuthRepository struct{ db *gorm.DB }

func NewAuthRepository(db *gorm.DB) service.AuthRepository { return AuthRepository{db: db} }

func (r AuthRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (r AuthRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r AuthRepository) GetByID(ctx context.Context, id uint) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return user, err
}
