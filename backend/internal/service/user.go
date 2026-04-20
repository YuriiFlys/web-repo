package service

import (
	"context"

	"project-management/internal/model"
)

type UserService interface {
	List(ctx context.Context) ([]model.User, error)
}

type UserRepository interface {
	List(ctx context.Context) ([]model.User, error)
}

type userService struct{ repo UserRepository }

func NewUserService(repo UserRepository) UserService                  { return &userService{repo: repo} }
func (s *userService) List(ctx context.Context) ([]model.User, error) { return s.repo.List(ctx) }
