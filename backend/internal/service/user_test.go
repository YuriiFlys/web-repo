package service

import (
	"context"
	"errors"
	"testing"

	"project-management/internal/model"
)

type stubUserRepo struct {
	listFn func(ctx context.Context) ([]model.User, error)
}

func (s stubUserRepo) List(ctx context.Context) ([]model.User, error) { return s.listFn(ctx) }

func TestUserServiceList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &userService{repo: stubUserRepo{listFn: func(ctx context.Context) ([]model.User, error) {
			return []model.User{{ID: 1, Email: "a@example.com"}}, nil
		}}}
		users, err := svc.List(context.Background())
		if err != nil || len(users) != 1 {
			t.Fatalf("users=%v err=%v", users, err)
		}
	})

	t.Run("error", func(t *testing.T) {
		svc := &userService{repo: stubUserRepo{listFn: func(ctx context.Context) ([]model.User, error) {
			return nil, errors.New("db down")
		}}}
		_, err := svc.List(context.Background())
		if err == nil || err.Error() != "db down" {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestNewUserService(t *testing.T) {
	if svc := NewUserService(stubUserRepo{}); svc == nil {
		t.Fatal("NewUserService returned nil")
	}
}
