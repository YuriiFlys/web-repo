package service

import (
	"context"
	"errors"
	"testing"

	"project-management/internal/httpx"
	"project-management/internal/model"
)

type stubCommentRepo struct {
	listFn   func(ctx context.Context, filter CommentListFilter) ([]model.Comment, int64, error)
	createFn func(ctx context.Context, comment *model.Comment) error
	getFn    func(ctx context.Context, id string) (model.Comment, error)
	saveFn   func(ctx context.Context, comment *model.Comment) error
	deleteFn func(ctx context.Context, id string) error
}

func (s stubCommentRepo) List(ctx context.Context, filter CommentListFilter) ([]model.Comment, int64, error) {
	return s.listFn(ctx, filter)
}
func (s stubCommentRepo) Create(ctx context.Context, comment *model.Comment) error {
	return s.createFn(ctx, comment)
}
func (s stubCommentRepo) Get(ctx context.Context, id string) (model.Comment, error) {
	return s.getFn(ctx, id)
}
func (s stubCommentRepo) Save(ctx context.Context, comment *model.Comment) error {
	return s.saveFn(ctx, comment)
}
func (s stubCommentRepo) Delete(ctx context.Context, id string) error { return s.deleteFn(ctx, id) }

func TestCommentService(t *testing.T) {
	ctx := context.Background()

	t.Run("list delegates", func(t *testing.T) {
		svc := &commentService{repo: stubCommentRepo{listFn: func(ctx context.Context, filter CommentListFilter) ([]model.Comment, int64, error) {
			if filter.Params.Page != 3 || filter.Author != "Bob" {
				t.Fatalf("filter = %+v", filter)
			}
			return []model.Comment{{ID: 1}}, 1, nil
		}}}
		items, total, err := svc.List(ctx, CommentListFilter{Params: httpx.ListParams{Page: 3}, Author: "Bob"})
		if err != nil || total != 1 || len(items) != 1 {
			t.Fatalf("items=%v total=%d err=%v", items, total, err)
		}
	})

	t.Run("create maps input", func(t *testing.T) {
		svc := &commentService{repo: stubCommentRepo{createFn: func(ctx context.Context, comment *model.Comment) error {
			if comment.TaskID != 5 || comment.Author != "Ann" {
				t.Fatalf("comment = %+v", comment)
			}
			return nil
		}}}
		_, err := svc.Create(ctx, CommentCreateInput{TaskID: 5, Author: "Ann", Text: "hello"})
		if err != nil {
			t.Fatalf("Create error = %v", err)
		}
	})

	t.Run("create error", func(t *testing.T) {
		svc := &commentService{repo: stubCommentRepo{createFn: func(ctx context.Context, comment *model.Comment) error {
			return errors.New("insert failed")
		}}}
		_, err := svc.Create(ctx, CommentCreateInput{TaskID: 5, Author: "Ann", Text: "hello"})
		if err == nil || err.Error() != "insert failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("get delegates success", func(t *testing.T) {
		svc := &commentService{repo: stubCommentRepo{getFn: func(ctx context.Context, id string) (model.Comment, error) {
			if id != "5" {
				t.Fatalf("id=%s", id)
			}
			return model.Comment{ID: 5}, nil
		}}}
		comment, err := svc.Get(ctx, "5")
		if err != nil || comment.ID != 5 {
			t.Fatalf("comment=%+v err=%v", comment, err)
		}
	})

	t.Run("update patches entity", func(t *testing.T) {
		svc := &commentService{repo: stubCommentRepo{
			getFn: func(ctx context.Context, id string) (model.Comment, error) {
				return model.Comment{ID: 1, Author: "Old", Text: "Old"}, nil
			},
			saveFn: func(ctx context.Context, comment *model.Comment) error {
				if comment.Author != "New" || comment.Text != "Updated" {
					t.Fatalf("comment = %+v", comment)
				}
				return nil
			},
		}}
		comment, err := svc.Update(ctx, "1", CommentUpdateInput{Author: ptr("New"), Text: ptr("Updated")})
		if err != nil || comment.Author != "New" {
			t.Fatalf("comment=%+v err=%v", comment, err)
		}
	})

	t.Run("update get error", func(t *testing.T) {
		svc := &commentService{repo: stubCommentRepo{getFn: func(ctx context.Context, id string) (model.Comment, error) {
			return model.Comment{}, errors.New("boom")
		}}}
		_, err := svc.Update(ctx, "1", CommentUpdateInput{})
		if err == nil || err.Error() != "boom" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("update save error", func(t *testing.T) {
		svc := &commentService{repo: stubCommentRepo{
			getFn:  func(ctx context.Context, id string) (model.Comment, error) { return model.Comment{ID: 1}, nil },
			saveFn: func(ctx context.Context, comment *model.Comment) error { return errors.New("save failed") },
		}}
		_, err := svc.Update(ctx, "1", CommentUpdateInput{})
		if err == nil || err.Error() != "save failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("delete error propagates", func(t *testing.T) {
		svc := &commentService{repo: stubCommentRepo{deleteFn: func(ctx context.Context, id string) error { return errors.New("boom") }}}
		if err := svc.Delete(ctx, "2"); err == nil || err.Error() != "boom" {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestNewCommentService(t *testing.T) {
	if svc := NewCommentService(stubCommentRepo{}); svc == nil {
		t.Fatal("NewCommentService returned nil")
	}
}
