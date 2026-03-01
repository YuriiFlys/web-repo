package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct{ db *gorm.DB }

func NewProjectHandler(db *gorm.DB) *ProjectHandler { return &ProjectHandler{db: db} }

type ProjectCreate struct {
	Title       string              `json:"title" binding:"required"`
	Description string              `json:"description"`
	Status      model.ProjectStatus `json:"status" binding:"required,oneof=active archived"`
}

type ProjectUpdate struct {
	Title       *string              `json:"title"`
	Description *string              `json:"description"`
	Status      *model.ProjectStatus `json:"status" binding:"omitempty,oneof=active archived"`
}

func (h *ProjectHandler) Register(r *gin.RouterGroup) {
	r.GET("/projects", h.List)
	r.POST("/projects", h.Create)
	r.GET("/projects/:id", h.Get)
	r.PUT("/projects/:id", h.Update)
	r.DELETE("/projects/:id", h.Delete)

	// nested tasks under project
	r.GET("/projects/:id/tasks", h.ListProjectTasks)
	r.POST("/projects/:id/tasks", h.CreateProjectTask)
}

// List godoc
// @Summary List projects
// @Description List projects with pagination, sorting, filtering, and optional include of tasks.
// @Tags Projects
// @Accept json
// @Produce json
// @Param page query int false "Page number (1-based)"
// @Param pageSize query int false "Page size"
// @Param sort query string false "Sort by field (prefix with - for desc)"
// @Param status query string false "Filter by status" Enums(active,archived)
// @Param q query string false "Search in title or description"
// @Param include query string false "Include related entities" Enums(tasks)
// @Success 200 {object} ProjectsListResponse
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /projects [get]
func (h *ProjectHandler) List(c *gin.Context) {
	lp := httpx.ParseListParams(c.Query("page"), c.Query("pageSize"), c.Query("sort"))
	q := strings.TrimSpace(c.Query("q"))
	status := strings.TrimSpace(c.Query("status"))
	include := strings.TrimSpace(c.Query("include")) // "tasks"

	db := h.db.Model(&model.Project{})

	if q != "" {
		like := "%" + q + "%"
		db = db.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	if include == "tasks" {
		db = db.Preload("Tasks")
	}

	allowedSort := map[string]string{
		"id":        "id",
		"title":     "title",
		"status":    "status",
		"createdAt": "created_at",
	}
	db = httpx.ApplySorting(db, allowedSort, lp, "created_at DESC")
	db = httpx.ApplyPagination(db, lp)

	var items []model.Project
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
// @Summary Create project
// @Description Create a new project.
// @Tags Projects
// @Accept json
// @Produce json
// @Param body body ProjectCreate true "Project payload"
// @Success 201 {object} model.Project
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /projects [post]
func (h *ProjectHandler) Create(c *gin.Context) {
	var body ProjectCreate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	p := model.Project{
		Title:       body.Title,
		Description: body.Description,
		Status:      body.Status,
	}

	if err := h.db.Create(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, p)
}

// Get godoc
// @Summary Get project
// @Description Get a project by ID. Optionally include tasks.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param include query string false "Include related entities" Enums(tasks)
// @Success 200 {object} model.Project
// @Failure 400 {object} httpx.APIError
// @Failure 404 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /projects/{id} [get]
func (h *ProjectHandler) Get(c *gin.Context) {
	var p model.Project

	include := strings.TrimSpace(c.Query("include")) // "tasks"
	db := h.db
	if include == "tasks" {
		db = db.Preload("Tasks")
	}

	if err := db.First(&p, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "project not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.JSON(http.StatusOK, p)
}

// Update godoc
// @Summary Update project
// @Description Partially update a project by ID.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param body body ProjectUpdate true "Project fields to update"
// @Success 200 {object} model.Project
// @Failure 400 {object} httpx.APIError
// @Failure 404 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /projects/{id} [put]
func (h *ProjectHandler) Update(c *gin.Context) {
	var body ProjectUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	var p model.Project
	if err := h.db.First(&p, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "project not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	if body.Title != nil {
		p.Title = *body.Title
	}
	if body.Description != nil {
		p.Description = *body.Description
	}
	if body.Status != nil {
		p.Status = *body.Status
	}

	if err := h.db.Save(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, p)
}

// Delete godoc
// @Summary Delete project
// @Description Delete a project by ID. Returns 204 even if it did not exist.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Success 204
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /projects/{id} [delete]
func (h *ProjectHandler) Delete(c *gin.Context) {
	// We return 204 even if the record didn't exist (idempotent delete is acceptable)
	if err := h.db.Delete(&model.Project{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// -------- Nested tasks under project --------

type TaskCreateUnderProject struct {
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description"`
	Status      model.TaskStatus `json:"status" binding:"required,oneof=todo in_progress done"`
	AssigneeID  *uint            `json:"assigneeId"`
	DueDate     *time.Time       `json:"dueDate"`
}

// ListProjectTasks godoc
// @Summary List tasks in project
// @Description List tasks under a specific project with pagination, sorting, and filtering.
// @Tags Projects,Tasks
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param page query int false "Page number (1-based)"
// @Param pageSize query int false "Page size"
// @Param sort query string false "Sort by field (prefix with - for desc)"
// @Param status query string false "Filter by status" Enums(todo,in_progress,done)
// @Param assigneeId query int false "Filter by assignee ID"
// @Success 200 {object} ProjectTasksListResponse
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /projects/{id}/tasks [get]
func (h *ProjectHandler) ListProjectTasks(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, "invalid projectId"))
		return
	}

	lp := httpx.ParseListParams(c.Query("page"), c.Query("pageSize"), c.Query("sort"))
	status := strings.TrimSpace(c.Query("status"))
	assigneeID := strings.TrimSpace(c.Query("assigneeId"))

	db := h.db.Model(&model.Task{}).Where("project_id = ?", projectID)

	if status != "" {
		db = db.Where("status = ?", status)
	}
	if assigneeID != "" {
		db = db.Where("assignee_id = ?", assigneeID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
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

// CreateProjectTask godoc
// @Summary Create task in project
// @Description Create a new task under a specific project.
// @Tags Projects,Tasks
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param body body TaskCreateUnderProject true "Task payload"
// @Success 201 {object} model.Task
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /projects/{id}/tasks [post]
func (h *ProjectHandler) CreateProjectTask(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, "invalid projectId"))
		return
	}

	var body TaskCreateUnderProject
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	t := model.Task{
		ProjectID:   uint(projectID),
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
