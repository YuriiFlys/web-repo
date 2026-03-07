//go:build integration

package integration_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"
	"project-management/internal/repository"
	"project-management/internal/service"

	"gorm.io/gorm"
)

func TestAuthRepositoryIntegration(t *testing.T) {
	db := openTestDB(t)
	resetTestDB(t, db)
	repo := repository.NewAuthRepository(db)
	ctx := context.Background()

	user := &model.User{Email: "alice@example.com", Name: "Alice", PasswordHash: "hashed"}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected created user ID")
	}

	byEmail, err := repo.FindByEmail(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if byEmail.ID != user.ID || byEmail.Name != "Alice" {
		t.Fatalf("unexpected user: %+v", byEmail)
	}

	byID, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if byID.Email != "alice@example.com" {
		t.Fatalf("unexpected user: %+v", byID)
	}

	_, err = repo.FindByEmail(ctx, "missing@example.com")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected record not found, got %v", err)
	}
}

func TestProjectRepositoryIntegration(t *testing.T) {
	db := openTestDB(t)
	resetTestDB(t, db)
	repo := repository.NewProjectRepository(db)
	ctx := context.Background()

	projectA := &model.Project{Title: "Backend", Description: "API work", Status: model.ProjectActive}
	projectB := &model.Project{Title: "Archive", Description: "Legacy", Status: model.ProjectArchived}
	if err := repo.Create(ctx, projectA); err != nil {
		t.Fatalf("Create projectA: %v", err)
	}
	if err := repo.Create(ctx, projectB); err != nil {
		t.Fatalf("Create projectB: %v", err)
	}

	assignee := uint(17)
	due := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)
	if err := repo.CreateTask(ctx, &model.Task{ProjectID: projectA.ID, Title: "Implement", Status: model.TaskTodo, AssigneeID: &assignee, DueDate: &due}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := repo.CreateTask(ctx, &model.Task{ProjectID: projectA.ID, Title: "Review", Status: model.TaskDone}); err != nil {
		t.Fatalf("CreateTask second: %v", err)
	}

	items, total, err := repo.List(ctx, service.ProjectListFilter{
		Params:       httpx.ListParams{Page: 1, PageSize: 10, Sort: []httpx.SortField{{Field: "title"}}},
		Query:        "API",
		Status:       string(model.ProjectActive),
		IncludeTasks: true,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || len(items) != 1 || len(items[0].Tasks) != 2 {
		t.Fatalf("unexpected list result: total=%d items=%+v", total, items)
	}

	got, err := repo.Get(ctx, toStringID(projectA.ID), true)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != projectA.ID || len(got.Tasks) != 2 {
		t.Fatalf("unexpected project: %+v", got)
	}

	got.Title = "Backend v2"
	if err := repo.Save(ctx, &got); err != nil {
		t.Fatalf("Save: %v", err)
	}

	updated, err := repo.Get(ctx, toStringID(projectA.ID), false)
	if err != nil {
		t.Fatalf("Get updated: %v", err)
	}
	if updated.Title != "Backend v2" {
		t.Fatalf("expected updated title, got %+v", updated)
	}

	tasks, taskTotal, err := repo.ListTasks(ctx, projectA.ID, service.ProjectTaskListFilter{
		Params:     httpx.ListParams{Page: 1, PageSize: 10, Sort: []httpx.SortField{{Field: "title", Desc: true}}},
		Status:     string(model.TaskTodo),
		AssigneeID: "17",
	})
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if taskTotal != 1 || len(tasks) != 1 || tasks[0].Title != "Implement" {
		t.Fatalf("unexpected tasks result: total=%d tasks=%+v", taskTotal, tasks)
	}

	if err := repo.Delete(ctx, toStringID(projectB.ID)); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = repo.Get(ctx, toStringID(projectB.ID), false)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected deleted project to be missing, got %v", err)
	}
}

func TestProjectRepositoryCountErrors(t *testing.T) {
	db := openTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	repo := repository.NewProjectRepository(db)
	ctx := context.Background()

	if _, _, err := repo.List(ctx, service.ProjectListFilter{}); err == nil {
		t.Fatal("expected list error")
	}
	if _, _, err := repo.ListTasks(ctx, 1, service.ProjectTaskListFilter{}); err == nil {
		t.Fatal("expected list tasks error")
	}
}

func TestTaskRepositoryIntegration(t *testing.T) {
	db := openTestDB(t)
	resetTestDB(t, db)
	repo := repository.NewTaskRepository(db)
	ctx := context.Background()

	project := &model.Project{Title: "Platform", Status: model.ProjectActive}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("seed project: %v", err)
	}
	assignee := uint(8)
	dueEarly := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	dueLate := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

	taskA := &model.Task{ProjectID: project.ID, Title: "Build", Status: model.TaskTodo, AssigneeID: &assignee, DueDate: &dueEarly}
	taskB := &model.Task{ProjectID: project.ID, Title: "Deploy", Status: model.TaskDone, DueDate: &dueLate}
	if err := repo.Create(ctx, taskA); err != nil {
		t.Fatalf("Create taskA: %v", err)
	}
	if err := repo.Create(ctx, taskB); err != nil {
		t.Fatalf("Create taskB: %v", err)
	}

	if err := repo.CreateComment(ctx, &model.Comment{TaskID: taskA.ID, Author: "Ann", Text: "first"}); err != nil {
		t.Fatalf("CreateComment first: %v", err)
	}
	if err := repo.CreateComment(ctx, &model.Comment{TaskID: taskA.ID, Author: "Bob", Text: "second"}); err != nil {
		t.Fatalf("CreateComment second: %v", err)
	}

	items, total, err := repo.List(ctx, service.TaskListFilter{
		Params:          httpx.ListParams{Page: 1, PageSize: 10, Sort: []httpx.SortField{{Field: "dueDate"}}},
		ProjectID:       toStringID(project.ID),
		Status:          string(model.TaskTodo),
		AssigneeID:      "8",
		DueFrom:         "2026-04-01",
		DueTo:           "2026-04-10",
		IncludeComments: true,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 1 || len(items) != 1 || len(items[0].Comments) != 2 {
		t.Fatalf("unexpected list result: total=%d items=%+v", total, items)
	}

	got, err := repo.Get(ctx, toStringID(taskA.ID), true)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != taskA.ID || len(got.Comments) != 2 {
		t.Fatalf("unexpected task: %+v", got)
	}

	got.Title = "Build v2"
	if err := repo.Save(ctx, &got); err != nil {
		t.Fatalf("Save: %v", err)
	}

	comments, commentTotal, err := repo.ListComments(ctx, toStringID(taskA.ID), service.TaskCommentListFilter{
		Params: httpx.ListParams{Page: 1, PageSize: 10},
		Author: "Ann",
	})
	if err != nil {
		t.Fatalf("ListComments: %v", err)
	}
	if commentTotal != 1 || len(comments) != 1 || comments[0].Author != "Ann" {
		t.Fatalf("unexpected comments result: total=%d comments=%+v", commentTotal, comments)
	}

	if err := repo.Delete(ctx, toStringID(taskB.ID)); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err = repo.Get(ctx, toStringID(taskB.ID), false)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected deleted task to be missing, got %v", err)
	}
}

func TestTaskRepositoryCountErrors(t *testing.T) {
	db := openTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	repo := repository.NewTaskRepository(db)
	ctx := context.Background()

	if _, _, err := repo.List(ctx, service.TaskListFilter{}); err == nil {
		t.Fatal("expected list error")
	}
	if _, _, err := repo.ListComments(ctx, "1", service.TaskCommentListFilter{}); err == nil {
		t.Fatal("expected list comments error")
	}
}

func TestCommentAndUserRepositoryIntegration(t *testing.T) {
	db := openTestDB(t)
	resetTestDB(t, db)
	commentRepo := repository.NewCommentRepository(db)
	userRepo := repository.NewUserRepository(db)
	ctx := context.Background()

	userA := &model.User{Email: "ann@example.com", Name: "Ann", PasswordHash: "hash-a"}
	userB := &model.User{Email: "bob@example.com", Name: "Bob", PasswordHash: "hash-b"}
	if err := db.Create(userA).Error; err != nil {
		t.Fatalf("seed userA: %v", err)
	}
	if err := db.Create(userB).Error; err != nil {
		t.Fatalf("seed userB: %v", err)
	}

	project := &model.Project{Title: "Docs", Status: model.ProjectActive}
	if err := db.Create(project).Error; err != nil {
		t.Fatalf("seed project: %v", err)
	}
	task := &model.Task{ProjectID: project.ID, Title: "Write", Status: model.TaskTodo}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("seed task: %v", err)
	}

	comment := &model.Comment{TaskID: task.ID, Author: "Ann", Text: "draft"}
	if err := commentRepo.Create(ctx, comment); err != nil {
		t.Fatalf("Create comment: %v", err)
	}

	items, total, err := commentRepo.List(ctx, service.CommentListFilter{
		Params: httpx.ListParams{Page: 1, PageSize: 10},
		TaskID: toStringID(task.ID),
		Author: "Ann",
	})
	if err != nil {
		t.Fatalf("List comments: %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].Text != "draft" {
		t.Fatalf("unexpected comments result: total=%d items=%+v", total, items)
	}

	got, err := commentRepo.Get(ctx, toStringID(comment.ID))
	if err != nil {
		t.Fatalf("Get comment: %v", err)
	}
	got.Text = "final"
	if err := commentRepo.Save(ctx, &got); err != nil {
		t.Fatalf("Save comment: %v", err)
	}

	updated, err := commentRepo.Get(ctx, toStringID(comment.ID))
	if err != nil {
		t.Fatalf("Get updated comment: %v", err)
	}
	if updated.Text != "final" {
		t.Fatalf("unexpected updated comment: %+v", updated)
	}

	users, err := userRepo.List(ctx)
	if err != nil {
		t.Fatalf("List users: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if err := commentRepo.Delete(ctx, toStringID(comment.ID)); err != nil {
		t.Fatalf("Delete comment: %v", err)
	}
	_, err = commentRepo.Get(ctx, toStringID(comment.ID))
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected deleted comment to be missing, got %v", err)
	}
}

func TestCommentRepositoryCountError(t *testing.T) {
	db := openTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}

	repo := repository.NewCommentRepository(db)
	if _, _, err := repo.List(context.Background(), service.CommentListFilter{}); err == nil {
		t.Fatal("expected list error")
	}
}

func toStringID(id uint) string {
	return fmt.Sprintf("%d", id)
}
