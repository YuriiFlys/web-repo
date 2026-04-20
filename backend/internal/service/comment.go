package service

import (
	"context"

	"project-management/internal/httpx"
	"project-management/internal/model"
)

type CommentListFilter struct {
	Params httpx.ListParams
	TaskID string
	Author string
}

type CommentCreateInput struct {
	TaskID uint
	Author string
	Text   string
}

type CommentUpdateInput struct {
	Author *string
	Text   *string
}

type CommentService interface {
	List(ctx context.Context, filter CommentListFilter) ([]model.Comment, int64, error)
	Create(ctx context.Context, input CommentCreateInput) (model.Comment, error)
	Get(ctx context.Context, id string) (model.Comment, error)
	Update(ctx context.Context, id string, input CommentUpdateInput) (model.Comment, error)
	Delete(ctx context.Context, id string) error
}

type CommentRepository interface {
	List(ctx context.Context, filter CommentListFilter) ([]model.Comment, int64, error)
	Create(ctx context.Context, comment *model.Comment) error
	Get(ctx context.Context, id string) (model.Comment, error)
	Save(ctx context.Context, comment *model.Comment) error
	Delete(ctx context.Context, id string) error
}

type commentService struct{ repo CommentRepository }

func NewCommentService(repo CommentRepository) CommentService { return &commentService{repo: repo} }
func (s *commentService) List(ctx context.Context, filter CommentListFilter) ([]model.Comment, int64, error) {
	return s.repo.List(ctx, filter)
}
func (s *commentService) Create(ctx context.Context, input CommentCreateInput) (model.Comment, error) {
	comment := model.Comment{TaskID: input.TaskID, Author: input.Author, Text: input.Text}
	return comment, s.repo.Create(ctx, &comment)
}
func (s *commentService) Get(ctx context.Context, id string) (model.Comment, error) {
	return s.repo.Get(ctx, id)
}
func (s *commentService) Update(ctx context.Context, id string, input CommentUpdateInput) (model.Comment, error) {
	comment, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Comment{}, err
	}
	if input.Author != nil {
		comment.Author = *input.Author
	}
	if input.Text != nil {
		comment.Text = *input.Text
	}
	return comment, s.repo.Save(ctx, &comment)
}
func (s *commentService) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
