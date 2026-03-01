package handler

import (
	"errors"
	"net/http"
	"strings"

	"project-management/internal/httpx"
	"project-management/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentHandler struct{ db *gorm.DB }

func NewCommentHandler(db *gorm.DB) *CommentHandler { return &CommentHandler{db: db} }

type CommentCreate struct {
	TaskID uint   `json:"taskId" binding:"required"`
	Author string `json:"author" binding:"required"`
	Text   string `json:"text" binding:"required"`
}

type CommentUpdate struct {
	Author *string `json:"author"`
	Text   *string `json:"text"`
}

func (h *CommentHandler) Register(r *gin.RouterGroup) {
	r.GET("/comments", h.List)
	r.POST("/comments", h.Create)
	r.GET("/comments/:id", h.Get)
	r.PUT("/comments/:id", h.Update)
	r.DELETE("/comments/:id", h.Delete)
}

// List godoc
// @Summary List comments
// @Description List comments with pagination, sorting, and filtering.
// @Tags Comments
// @Accept json
// @Produce json
// @Param page query int false "Page number (1-based)"
// @Param pageSize query int false "Page size"
// @Param sort query string false "Sort by field (prefix with - for desc)"
// @Param taskId query int false "Filter by task ID"
// @Param author query string false "Filter by author"
// @Success 200 {object} CommentsListResponse
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /comments [get]
func (h *CommentHandler) List(c *gin.Context) {
	lp := httpx.ParseListParams(c.Query("page"), c.Query("pageSize"), c.Query("sort"))

	taskId := strings.TrimSpace(c.Query("taskId"))
	author := strings.TrimSpace(c.Query("author"))

	db := h.db.Model(&model.Comment{})
	if taskId != "" {
		db = db.Where("task_id = ?", taskId)
	}
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

// Create godoc
// @Summary Create comment
// @Description Create a new comment.
// @Tags Comments
// @Accept json
// @Produce json
// @Param body body CommentCreate true "Comment payload"
// @Success 201 {object} model.Comment
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	var body CommentCreate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	x := model.Comment{
		TaskID: body.TaskID,
		Author: body.Author,
		Text:   body.Text,
	}
	if err := h.db.Create(&x).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, x)
}

// Get godoc
// @Summary Get comment
// @Description Get a comment by ID.
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Comment ID"
// @Success 200 {object} model.Comment
// @Failure 400 {object} httpx.APIError
// @Failure 404 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /comments/{id} [get]
func (h *CommentHandler) Get(c *gin.Context) {
	var x model.Comment
	if err := h.db.First(&x, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "comment not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.JSON(http.StatusOK, x)
}

// Update godoc
// @Summary Update comment
// @Description Partially update a comment by ID.
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Comment ID"
// @Param body body CommentUpdate true "Comment fields to update"
// @Success 200 {object} model.Comment
// @Failure 400 {object} httpx.APIError
// @Failure 404 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /comments/{id} [put]
func (h *CommentHandler) Update(c *gin.Context) {
	var body CommentUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	var x model.Comment
	if err := h.db.First(&x, c.Param("id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "comment not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	if body.Author != nil {
		x.Author = *body.Author
	}
	if body.Text != nil {
		x.Text = *body.Text
	}

	if err := h.db.Save(&x).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, x)
}

// Delete godoc
// @Summary Delete comment
// @Description Delete a comment by ID.
// @Tags Comments
// @Accept json
// @Produce json
// @Param id path int true "Comment ID"
// @Success 204
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /comments/{id} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	if err := h.db.Delete(&model.Comment{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}
