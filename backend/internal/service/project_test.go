package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"
)

type stubProjectRepo struct {
	listFn       func(ctx context.Context, filter ProjectListFilter) ([]model.Project, int64, error)
	createFn     func(ctx context.Context, project *model.Project) error
	getFn        func(ctx context.Context, id string, includeTasks bool) (model.Project, error)
	saveFn       func(ctx context.Context, project *model.Project) error
	deleteFn     func(ctx context.Context, id string) error
	listTasksFn  func(ctx context.Context, projectID uint, filter ProjectTaskListFilter) ([]model.Task, int64, error)
	createTaskFn func(ctx context.Context, task *model.Task) error
}

func (s stubProjectRepo) List(ctx context.Context, filter ProjectListFilter) ([]model.Project, int64, error) {
	return s.listFn(ctx, filter)
}
func (s stubProjectRepo) Create(ctx context.Context, project *model.Project) error {
	return s.createFn(ctx, project)
}
func (s stubProjectRepo) Get(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
	return s.getFn(ctx, id, includeTasks)
}
func (s stubProjectRepo) Save(ctx context.Context, project *model.Project) error {
	return s.saveFn(ctx, project)
}
func (s stubProjectRepo) Delete(ctx context.Context, id string) error { return s.deleteFn(ctx, id) }
func (s stubProjectRepo) ListTasks(ctx context.Context, projectID uint, filter ProjectTaskListFilter) ([]model.Task, int64, error) {
	return s.listTasksFn(ctx, projectID, filter)
}
func (s stubProjectRepo) CreateTask(ctx context.Context, task *model.Task) error {
	return s.createTaskFn(ctx, task)
}

func TestProjectService(t *testing.T) {
	ctx := context.Background()

	t.Run("list delegates", func(t *testing.T) {
		svc := &projectService{repo: stubProjectRepo{listFn: func(ctx context.Context, filter ProjectListFilter) ([]model.Project, int64, error) {
			if filter.Params.Page != 2 || !filter.IncludeTasks {
				t.Fatalf("filter = %+v", filter)
			}
			return []model.Project{{ID: 1}}, 1, nil
		}}}
		items, total, err := svc.List(ctx, ProjectListFilter{Params: httpx.ListParams{Page: 2}, IncludeTasks: true})
		if err != nil || total != 1 || len(items) != 1 {
			t.Fatalf("items=%v total=%d err=%v", items, total, err)
		}
	})

	t.Run("create maps input", func(t *testing.T) {
		svc := &projectService{repo: stubProjectRepo{createFn: func(ctx context.Context, project *model.Project) error {
			if project.Title != "API" || project.Status != model.ProjectActive {
				t.Fatalf("project = %+v", project)
			}
			project.ID = 7
			return nil
		}}}
		project, err := svc.Create(ctx, ProjectCreateInput{Title: "API", Description: "desc", Status: model.ProjectActive})
		if err != nil || project.ID != 7 {
			t.Fatalf("project=%+v err=%v", project, err)
		}
	})

	t.Run("create error", func(t *testing.T) {
		svc := &projectService{repo: stubProjectRepo{createFn: func(ctx context.Context, project *model.Project) error {
			return errors.New("insert failed")
		}}}
		_, err := svc.Create(ctx, ProjectCreateInput{Title: "API"})
		if err == nil || err.Error() != "insert failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("get delegates success", func(t *testing.T) {
		svc := &projectService{repo: stubProjectRepo{getFn: func(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
			if id != "11" || !includeTasks {
				t.Fatalf("id=%s includeTasks=%v", id, includeTasks)
			}
			return model.Project{ID: 11}, nil
		}}}
		project, err := svc.Get(ctx, "11", true)
		if err != nil || project.ID != 11 {
			t.Fatalf("project=%+v err=%v", project, err)
		}
	})

	t.Run("update patches existing entity", func(t *testing.T) {
		status := model.ProjectArchived
		svc := &projectService{repo: stubProjectRepo{
			getFn: func(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
				return model.Project{ID: 3, Title: "Old", Description: "Old desc", Status: model.ProjectActive}, nil
			},
			saveFn: func(ctx context.Context, project *model.Project) error {
				if project.Title != "New" || project.Description != "New desc" || project.Status != status {
					t.Fatalf("project = %+v", project)
				}
				return nil
			},
		}}
		project, err := svc.Update(ctx, "3", ProjectUpdateInput{Title: ptr("New"), Description: ptr("New desc"), Status: &status})
		if err != nil || project.Title != "New" {
			t.Fatalf("project=%+v err=%v", project, err)
		}
	})

	t.Run("update get error", func(t *testing.T) {
		svc := &projectService{repo: stubProjectRepo{
			getFn: func(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
				return model.Project{}, errors.New("boom")
			},
		}}
		_, err := svc.Update(ctx, "3", ProjectUpdateInput{})
		if err == nil || err.Error() != "boom" {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("update save error", func(t *testing.T) {
		svc := &projectService{repo: stubProjectRepo{
			getFn: func(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
				return model.Project{ID: 1}, nil
			},
			saveFn: func(ctx context.Context, project *model.Project) error { return errors.New("save failed") },
		}}
		_, err := svc.Update(ctx, "1", ProjectUpdateInput{})
		if err == nil || err.Error() != "save failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("delete delegates", func(t *testing.T) {
		svc := &projectService{repo: stubProjectRepo{deleteFn: func(ctx context.Context, id string) error {
			if id != "9" {
				t.Fatalf("id = %s", id)
			}
			return nil
		}}}
		if err := svc.Delete(ctx, "9"); err != nil {
			t.Fatalf("Delete error = %v", err)
		}
	})

	t.Run("list tasks delegates", func(t *testing.T) {
		svc := &projectService{repo: stubProjectRepo{listTasksFn: func(ctx context.Context, projectID uint, filter ProjectTaskListFilter) ([]model.Task, int64, error) {
			if projectID != 8 || filter.AssigneeID != "4" {
				t.Fatalf("projectID=%d filter=%+v", projectID, filter)
			}
			return []model.Task{{ID: 1}}, 1, nil
		}}}
		items, total, err := svc.ListTasks(ctx, 8, ProjectTaskListFilter{AssigneeID: "4"})
		if err != nil || total != 1 || len(items) != 1 {
			t.Fatalf("items=%v total=%d err=%v", items, total, err)
		}
	})

	t.Run("create task maps input", func(t *testing.T) {
		due := time.Now()
		svc := &projectService{repo: stubProjectRepo{createTaskFn: func(ctx context.Context, task *model.Task) error {
			if task.ProjectID != 5 || task.Title != "Ship" || task.DueDate != &due {
				t.Fatalf("task = %+v", task)
			}
			return nil
		}}}
		_, err := svc.CreateTask(ctx, ProjectTaskCreateInput{ProjectID: 5, Title: "Ship", Status: model.TaskTodo, DueDate: &due})
		if err != nil {
			t.Fatalf("CreateTask error = %v", err)
		}
	})

	t.Run("create task error", func(t *testing.T) {
		svc := &projectService{repo: stubProjectRepo{createTaskFn: func(ctx context.Context, task *model.Task) error {
			return errors.New("insert failed")
		}}}
		_, err := svc.CreateTask(ctx, ProjectTaskCreateInput{ProjectID: 5, Title: "Ship"})
		if err == nil || err.Error() != "insert failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("get error propagates", func(t *testing.T) {
		svc := &projectService{repo: stubProjectRepo{getFn: func(ctx context.Context, id string, includeTasks bool) (model.Project, error) {
			return model.Project{}, errors.New("boom")
		}}}
		_, err := svc.Get(ctx, "1", false)
		if err == nil || err.Error() != "boom" {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestNewProjectService(t *testing.T) {
	if svc := NewProjectService(stubProjectRepo{}); svc == nil {
		t.Fatal("NewProjectService returned nil")
	}
}

func ptr[T any](v T) *T { return &v }
