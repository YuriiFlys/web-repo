package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/model"
	"project-management/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskHandler struct{ service service.TaskService }

func NewTaskHandler(service service.TaskService) *TaskHandler { return &TaskHandler{service: service} }

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
	r.GET("/tasks/:id/comments", h.ListTaskComments)
	r.POST("/tasks/:id/comments", h.CreateTaskComment)
}

func (h *TaskHandler) List(c *gin.Context) {
	lp := httpx.ParseListParams(c.Query("page"), c.Query("pageSize"), c.Query("sort"))

	items, total, err := h.service.List(c.Request.Context(), service.TaskListFilter{
		Params:          lp,
		ProjectID:       strings.TrimSpace(c.Query("projectId")),
		Status:          strings.TrimSpace(c.Query("status")),
		AssigneeID:      strings.TrimSpace(c.Query("assigneeId")),
		DueFrom:         strings.TrimSpace(c.Query("dueFrom")),
		DueTo:           strings.TrimSpace(c.Query("dueTo")),
		IncludeComments: strings.TrimSpace(c.Query("include")) == "comments",
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

func (h *TaskHandler) Create(c *gin.Context) {
	var body TaskCreate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	t, err := h.service.Create(c.Request.Context(), service.TaskCreateInput{
		ProjectID:   body.ProjectID,
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

func (h *TaskHandler) Get(c *gin.Context) {
	t, err := h.service.Get(c.Request.Context(), c.Param("id"), strings.TrimSpace(c.Query("include")) == "comments")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "task not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) Update(c *gin.Context) {
	var body TaskUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	t, err := h.service.Update(c.Request.Context(), c.Param("id"), service.TaskUpdateInput{
		Title:       body.Title,
		Description: body.Description,
		Status:      body.Status,
		AssigneeID:  body.AssigneeID,
		DueDate:     body.DueDate,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "task not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, t)
}

func (h *TaskHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

type CommentCreateUnderTask struct {
	Author string `json:"author" binding:"required"`
	Text   string `json:"text" binding:"required"`
}

func (h *TaskHandler) ListTaskComments(c *gin.Context) {
	taskID := c.Param("id")
	lp := httpx.ParseListParams(c.Query("page"), c.Query("pageSize"), c.Query("sort"))
	author := strings.TrimSpace(c.Query("author"))

	items, total, err := h.service.ListComments(c.Request.Context(), taskID, service.TaskCommentListFilter{
		Params: lp,
		Author: author,
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

func (h *TaskHandler) CreateTaskComment(c *gin.Context) {
	taskID := c.Param("id")

	var body CommentCreateUnderTask
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	x, err := h.service.CreateComment(c.Request.Context(), service.TaskCommentCreateInput{
		TaskID: mustUint(taskID),
		Author: body.Author,
		Text:   body.Text,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, x)
}

func mustUint(s string) uint {
	var n uint64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0
		}
		n = n*10 + uint64(s[i]-'0')
	}
	return uint(n)
}
