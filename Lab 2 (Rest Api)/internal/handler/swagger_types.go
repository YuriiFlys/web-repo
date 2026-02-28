package handler

import "project-management/internal/model"

// ProjectsListResponse is a paginated list response for projects.
type ProjectsListResponse struct {
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	Items    []model.Project `json:"items"`
}

// ProjectTasksListResponse is a paginated list response for tasks under a project.
type ProjectTasksListResponse struct {
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
	Items    []model.Task `json:"items"`
}

// TasksListResponse is a paginated list response for tasks.
type TasksListResponse struct {
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
	Items    []model.Task `json:"items"`
}

// TaskCommentsListResponse is a paginated list response for comments under a task.
type TaskCommentsListResponse struct {
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	Items    []model.Comment `json:"items"`
}

// CommentsListResponse is a paginated list response for comments.
type CommentsListResponse struct {
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	Items    []model.Comment `json:"items"`
}
