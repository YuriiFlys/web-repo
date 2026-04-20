package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"
	"project-management/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct{ service service.ProjectService }

func NewProjectHandler(service service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

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
	r.GET("/projects/:id/tasks", h.ListProjectTasks)
	r.POST("/projects/:id/tasks", h.CreateProjectTask)
}

func (h *ProjectHandler) List(c *gin.Context) {
	lp := httpx.ParseListParams(c.Query("page"), c.Query("pageSize"), c.Query("sort"))
	q := strings.TrimSpace(c.Query("q"))
	status := strings.TrimSpace(c.Query("status"))
	include := strings.TrimSpace(c.Query("include"))

	items, total, err := h.service.List(c.Request.Context(), service.ProjectListFilter{
		Params:       lp,
		Query:        q,
		Status:       status,
		IncludeTasks: include == "tasks",
	})
	if err != nil {
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

func (h *ProjectHandler) Create(c *gin.Context) {
	var body ProjectCreate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	p, err := h.service.Create(c.Request.Context(), service.ProjectCreateInput{
		Title:       body.Title,
		Description: body.Description,
		Status:      body.Status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (h *ProjectHandler) Get(c *gin.Context) {
	p, err := h.service.Get(c.Request.Context(), c.Param("id"), strings.TrimSpace(c.Query("include")) == "tasks")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "project not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	var body ProjectUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	p, err := h.service.Update(c.Request.Context(), c.Param("id"), service.ProjectUpdateInput{
		Title:       body.Title,
		Description: body.Description,
		Status:      body.Status,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "project not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

type TaskCreateUnderProject struct {
	Title       string           `json:"title" binding:"required"`
	Description string           `json:"description"`
	Status      model.TaskStatus `json:"status" binding:"required,oneof=todo in_progress done"`
	AssigneeID  *uint            `json:"assigneeId"`
	DueDate     *time.Time       `json:"dueDate"`
}

func (h *ProjectHandler) ListProjectTasks(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, "invalid projectId"))
		return
	}

	lp := httpx.ParseListParams(c.Query("page"), c.Query("pageSize"), c.Query("sort"))
	status := strings.TrimSpace(c.Query("status"))
	assigneeID := strings.TrimSpace(c.Query("assigneeId"))

	items, total, err := h.service.ListTasks(c.Request.Context(), uint(projectID), service.ProjectTaskListFilter{
		Params:     lp,
		Status:     status,
		AssigneeID: assigneeID,
	})
	if err != nil {
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

	t, err := h.service.CreateTask(c.Request.Context(), service.ProjectTaskCreateInput{
		ProjectID:   uint(projectID),
		Title:       body.Title,
		Description: body.Description,
		Status:      body.Status,
		AssigneeID:  body.AssigneeID,
		DueDate:     body.DueDate,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, t)
}
