package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"
)

type stubTaskRepo struct {
	listFn          func(ctx context.Context, filter TaskListFilter) ([]model.Task, int64, error)
	createFn        func(ctx context.Context, task *model.Task) error
	getFn           func(ctx context.Context, id string, includeComments bool) (model.Task, error)
	saveFn          func(ctx context.Context, task *model.Task) error
	deleteFn        func(ctx context.Context, id string) error
	listCommentsFn  func(ctx context.Context, taskID string, filter TaskCommentListFilter) ([]model.Comment, int64, error)
	createCommentFn func(ctx context.Context, comment *model.Comment) error
}

func (s stubTaskRepo) List(ctx context.Context, filter TaskListFilter) ([]model.Task, int64, error) {
	return s.listFn(ctx, filter)
}
func (s stubTaskRepo) Create(ctx context.Context, task *model.Task) error {
	return s.createFn(ctx, task)
}
func (s stubTaskRepo) Get(ctx context.Context, id string, includeComments bool) (model.Task, error) {
	return s.getFn(ctx, id, includeComments)
}
func (s stubTaskRepo) Save(ctx context.Context, task *model.Task) error { return s.saveFn(ctx, task) }
func (s stubTaskRepo) Delete(ctx context.Context, id string) error      { return s.deleteFn(ctx, id) }
func (s stubTaskRepo) ListComments(ctx context.Context, taskID string, filter TaskCommentListFilter) ([]model.Comment, int64, error) {
	return s.listCommentsFn(ctx, taskID, filter)
}
func (s stubTaskRepo) CreateComment(ctx context.Context, comment *model.Comment) error {
	return s.createCommentFn(ctx, comment)
}

func TestTaskService(t *testing.T) {
	ctx := context.Background()

	t.Run("list delegates", func(t *testing.T) {
		svc := &taskService{repo: stubTaskRepo{listFn: func(ctx context.Context, filter TaskListFilter) ([]model.Task, int64, error) {
			if filter.Params.PageSize != 5 || !filter.IncludeComments {
				t.Fatalf("filter = %+v", filter)
			}
			return []model.Task{{ID: 1}}, 1, nil
		}}}
		items, total, err := svc.List(ctx, TaskListFilter{Params: httpx.ListParams{PageSize: 5}, IncludeComments: true})
		if err != nil || total != 1 || len(items) != 1 {
			t.Fatalf("items=%v total=%d err=%v", items, total, err)
		}
	})

	t.Run("create maps input", func(t *testing.T) {
		svc := &taskService{repo: stubTaskRepo{createFn: func(ctx context.Context, task *model.Task) error {
			if task.ProjectID != 2 || task.Title != "Build" {
				t.Fatalf("task = %+v", task)
			}
			return nil
		}}}
		_, err := svc.Create(ctx, TaskCreateInput{ProjectID: 2, Title: "Build", Status: model.TaskTodo})
		if err != nil {
			t.Fatalf("Create error = %v", err)
		}
	})

	t.Run("create error", func(t *testing.T) {
		svc := &taskService{repo: stubTaskRepo{createFn: func(ctx context.Context, task *model.Task) error {
			return errors.New("insert failed")
		}}}
		_, err := svc.Create(ctx, TaskCreateInput{ProjectID: 2, Title: "Build"})
		if err == nil || err.Error() != "insert failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("get delegates success", func(t *testing.T) {
		svc := &taskService{repo: stubTaskRepo{getFn: func(ctx context.Context, id string, includeComments bool) (model.Task, error) {
			if id != "7" || !includeComments {
				t.Fatalf("id=%s includeComments=%v", id, includeComments)
			}
			return model.Task{ID: 7}, nil
		}}}
		task, err := svc.Get(ctx, "7", true)
		if err != nil || task.ID != 7 {
			t.Fatalf("task=%+v err=%v", task, err)
		}
	})

	t.Run("update patches nullables", func(t *testing.T) {
		status := model.TaskDone
		assignee := uint(7)
		due := time.Now()
		svc := &taskService{repo: stubTaskRepo{
			getFn: func(ctx context.Context, id string, includeComments bool) (model.Task, error) {
				return model.Task{ID: 1, Title: "Old", Status: model.TaskTodo}, nil
			},
			saveFn: func(ctx context.Context, task *model.Task) error {
				if task.Title != "New" || task.Description != "Updated desc" || task.Status != status || task.AssigneeID == nil || *task.AssigneeID != assignee || task.DueDate == nil || !task.DueDate.Equal(due) {
					t.Fatalf("task = %+v", task)
				}
				return nil
			},
		}}
		_, err := svc.Update(ctx, "1", TaskUpdateInput{Title: ptr("New"), Description: ptr("Updated desc"), Status: &status, AssigneeID: ptr(&assignee), DueDate: ptr(&due)})
		if err != nil {
			t.Fatalf("Update error = %v", err)
		}
	})

	t.Run("update get error", func(t *testing.T) {
		svc := &taskService{repo: stubTaskRepo{getFn: func(ctx context.Context, id string, includeComments bool) (model.Task, error) {
			return model.Task{}, errors.New("boom")
		}}}
		_, err := svc.Update(ctx, "1", TaskUpdateInput{})
		if err == nil || err.Error() != "boom" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("update save error", func(t *testing.T) {
		svc := &taskService{repo: stubTaskRepo{
			getFn: func(ctx context.Context, id string, includeComments bool) (model.Task, error) {
				return model.Task{ID: 1}, nil
			},
			saveFn: func(ctx context.Context, task *model.Task) error { return errors.New("save failed") },
		}}
		_, err := svc.Update(ctx, "1", TaskUpdateInput{})
		if err == nil || err.Error() != "save failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("list comments delegates", func(t *testing.T) {
		svc := &taskService{repo: stubTaskRepo{listCommentsFn: func(ctx context.Context, taskID string, filter TaskCommentListFilter) ([]model.Comment, int64, error) {
			if taskID != "3" || filter.Author != "Ann" {
				t.Fatalf("taskID=%s filter=%+v", taskID, filter)
			}
			return []model.Comment{{ID: 1}}, 1, nil
		}}}
		items, total, err := svc.ListComments(ctx, "3", TaskCommentListFilter{Author: "Ann"})
		if err != nil || total != 1 || len(items) != 1 {
			t.Fatalf("items=%v total=%d err=%v", items, total, err)
		}
	})

	t.Run("create comment maps input", func(t *testing.T) {
		svc := &taskService{repo: stubTaskRepo{createCommentFn: func(ctx context.Context, comment *model.Comment) error {
			if comment.TaskID != 4 || comment.Text != "hello" {
				t.Fatalf("comment = %+v", comment)
			}
			return nil
		}}}
		_, err := svc.CreateComment(ctx, TaskCommentCreateInput{TaskID: 4, Author: "Ann", Text: "hello"})
		if err != nil {
			t.Fatalf("CreateComment error = %v", err)
		}
	})

	t.Run("create comment error", func(t *testing.T) {
		svc := &taskService{repo: stubTaskRepo{createCommentFn: func(ctx context.Context, comment *model.Comment) error {
			return errors.New("insert failed")
		}}}
		_, err := svc.CreateComment(ctx, TaskCommentCreateInput{TaskID: 4, Author: "Ann", Text: "hello"})
		if err == nil || err.Error() != "insert failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("delete error propagates", func(t *testing.T) {
		svc := &taskService{repo: stubTaskRepo{deleteFn: func(ctx context.Context, id string) error { return errors.New("boom") }}}
		if err := svc.Delete(ctx, "9"); err == nil || err.Error() != "boom" {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestNewTaskService(t *testing.T) {
	if svc := NewTaskService(stubTaskRepo{}); svc == nil {
		t.Fatal("NewTaskService returned nil")
	}
}
