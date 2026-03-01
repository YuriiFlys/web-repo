package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskHandler struct{ db *gorm.DB }

func NewTaskHandler(db *gorm.DB) *TaskHandler { return &TaskHandler{db: db} }

type TaskCreate struct {
	ProjectID   uint             `json:"projectId" binding:"required"`
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description"`
	Status      model.TaskStatus `json:"status" binding:"required,oneof=todo in_progress done"`
	AssigneeID  *uint            `json:"assigneeId"`
	DueDate     *time.Time       `json:"dueDate"`
}

type TaskUpdate struct {
	Title       *string           `json:"title"`
	Description *string           `json:"description"`
	Status      *model.TaskStatus `json:"status" binding:"omitempty,oneof=todo in_progress done"`
	AssigneeID  **uint            `json:"assigneeId"`
	DueDate     **time.Time       `json:"dueDate"`
}

func (h *TaskHandler) Register(r *gin.RouterGroup) {
	r.GET("/tasks", h.List)
	r.POST("/tasks", h.Create)
	r.GET("/tasks/:id", h.Get)
	r.PUT("/tasks/:id", h.Update)
	r.DELETE("/tasks/:id", h.Delete)

	// nested comments under task
	r.GET("/tasks/:id/comments", h.ListTaskComments)
	r.POST("/tasks/:id/comments", h.CreateTaskComment)
}

// List godoc
// @Summary List tasks
// @Description List tasks with pagination, sorting, filtering, and optional include of comments.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param page query int false "Page number (1-based)"
// @Param pageSize query int false "Page size"
// @Param sort query string false "Sort by field (prefix with - for desc)"
// @Param projectId query int false "Filter by project ID"
// @Param status query string false "Filter by status" Enums(todo,in_progress,done)
// @Param assigneeId query int false "Filter by assignee ID"
// @Param dueFrom query string false "Filter by due date from (YYYY-MM-DD)"
// @Param dueTo query string false "Filter by due date to (YYYY-MM-DD)"
// @Param include query string false "Include related entities" Enums(comments)
// @Success 200 {object} TasksListResponse
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /tasks [get]
func (h *TaskHandler) List(c *gin.Context) {
	lp := httpx.ParseListParams(c.Query("page"), c.Query("pageSize"), c.Query("sort"))

	projectId := strings.TrimSpace(c.Query("projectId"))
	status := strings.TrimSpace(c.Query("status"))
	assigneeID := strings.TrimSpace(c.Query("assigneeId"))
	dueFrom := strings.TrimSpace(c.Query("dueFrom"))
	dueTo := strings.TrimSpace(c.Query("dueTo"))
	include := strings.TrimSpace(c.Query("include")) // "comments"

	db := h.db.Model(&model.Task{})

	if projectId != "" {
		db = db.Where("project_id = ?", projectId)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if assigneeID != "" {
		db = db.Where("assignee_id = ?", assigneeID)
	}
	if dueFrom != "" {
		if t, err := time.Parse("2006-01-02", dueFrom); err == nil {
			db = db.Where("due_date >= ?", t)
		}
	}
	if dueTo != "" {
		if t, err := time.Parse("2006-01-02", dueTo); err == nil {
			db = db.Where("due_date <= ?", t)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	if include == "comments" {
		db = db.Preload("Comments")
	}

	allowedSort := map[string]string{
		"id":        "id",
		"title":     "title",
		"status":    "status",
		"dueDate":   "due_date",
		"createdAt": "created_at",
	}
	db = httpx.ApplySorting(db, allowedSort, lp, "created_at DESC")
	db = httpx.ApplyPagination(db, lp)

	var items []model.Task
	if err := db.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     lp.Page,
		"pageSize": lp.PageSize,
		"items":    items,
		"isLast":   httpx.IsLast(total, lp),
	})
}

// Create godoc
// @Summary Create task
// @Description Create a new task.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param body body TaskCreate true "Task payload"
// @Success 201 {object} model.Task
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	var body TaskCreate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	t := model.Task{
		ProjectID:   body.ProjectID,
		Title:       body.Title,
		Description: body.Description,
		Status:      body.Status,
		AssigneeID:  body.AssigneeID,
		DueDate:     body.DueDate,
	}

	if err := h.db.Create(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, t)
}

// Get godoc
// @Summary Get task
// @Description Get a task by ID. Optionally include comments.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param include query string false "Include related entities" Enums(comments)
// @Success 200 {object} model.Task
// @Failure 400 {object} httpx.APIError
// @Failure 404 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /tasks/{id} [get]
func (h *TaskHandler) Get(c *gin.Context) {
	var t model.Task

	include := strings.TrimSpace(c.Query("include")) // "comments"
	db := h.db
	if include == "comments" {
		db = db.Preload("Comments")
	}

	if err := db.First(&t, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "task not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.JSON(http.StatusOK, t)
}

// Update godoc
// @Summary Update task
// @Description Partially update a task by ID. To clear dueDate, send {"dueDate": null}.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param body body TaskUpdate true "Task fields to update"
// @Success 200 {object} model.Task
// @Failure 400 {object} httpx.APIError
// @Failure 404 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /tasks/{id} [put]
func (h *TaskHandler) Update(c *gin.Context) {
	var body TaskUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	var t model.Task
	if err := h.db.First(&t, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "task not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	if body.Title != nil {
		t.Title = *body.Title
	}
	if body.Description != nil {
		t.Description = *body.Description
	}
	if body.Status != nil {
		t.Status = *body.Status
	}
	if body.AssigneeID != nil {
		t.AssigneeID = *body.AssigneeID
	}
	if body.DueDate != nil {
		t.DueDate = *body.DueDate // if *DueDate == nil => set NULL
	}

	if err := h.db.Save(&t).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, t)
}

// Delete godoc
// @Summary Delete task
// @Description Delete a task by ID.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 204
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /tasks/{id} [delete]
func (h *TaskHandler) Delete(c *gin.Context) {
	if err := h.db.Delete(&model.Task{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// -------- Nested comments under task --------

type CommentCreateUnderTask struct {
	Author string `json:"author" binding:"required"`
	Text   string `json:"text" binding:"required"`
}

// ListTaskComments godoc
// @Summary List comments for task
// @Description List comments under a specific task with pagination, sorting, and filtering.
// @Tags Tasks,Comments
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param page query int false "Page number (1-based)"
// @Param pageSize query int false "Page size"
// @Param sort query string false "Sort by field (prefix with - for desc)"
// @Param author query string false "Filter by author"
// @Success 200 {object} TaskCommentsListResponse
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /tasks/{id}/comments [get]
func (h *TaskHandler) ListTaskComments(c *gin.Context) {
	taskID := c.Param("id")

	lp := httpx.ParseListParams(c.Query("page"), c.Query("pageSize"), c.Query("sort"))
	author := strings.TrimSpace(c.Query("author"))

	db := h.db.Model(&model.Comment{}).Where("task_id = ?", taskID)
	if author != "" {
		db = db.Where("author = ?", author)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	allowedSort := map[string]string{
		"id":        "id",
		"createdAt": "created_at",
	}

	db = httpx.ApplySorting(db, allowedSort, lp, "created_at DESC")
	db = httpx.ApplyPagination(db, lp)

	var items []model.Comment
	if err := db.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"page":     lp.Page,
		"pageSize": lp.PageSize,
		"items":    items,
		"isLast":   httpx.IsLast(total, lp),
	})
}

// CreateTaskComment godoc
// @Summary Create comment for task
// @Description Create a new comment under a specific task.
// @Tags Tasks,Comments
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param body body CommentCreateUnderTask true "Comment payload"
// @Success 201 {object} model.Comment
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /tasks/{id}/comments [post]
func (h *TaskHandler) CreateTaskComment(c *gin.Context) {
	taskID := c.Param("id")

	var body CommentCreateUnderTask
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	x := model.Comment{
		TaskID: mustUint(taskID),
		Author: body.Author,
		Text:   body.Text,
	}
	if err := h.db.Create(&x).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, x)
}

func mustUint(s string) uint {
	// safe enough for coursework; you can make it strict if you want
	var n uint64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0
		}
		n = n*10 + uint64(s[i]-'0')
	}
	return uint(n)
}
