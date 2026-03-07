package handler

import (
	"errors"
	"net/http"
	"strings"

	"project-management/internal/httpx"
	"project-management/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentHandler struct{ service service.CommentService }

func NewCommentHandler(service service.CommentService) *CommentHandler {
	return &CommentHandler{service: service}
}

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

func (h *CommentHandler) List(c *gin.Context) {
	lp := httpx.ParseListParams(c.Query("page"), c.Query("pageSize"), c.Query("sort"))

	items, total, err := h.service.List(c.Request.Context(), service.CommentListFilter{
		Params: lp,
		TaskID: strings.TrimSpace(c.Query("taskId")),
		Author: strings.TrimSpace(c.Query("author")),
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

func (h *CommentHandler) Create(c *gin.Context) {
	var body CommentCreate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	x, err := h.service.Create(c.Request.Context(), service.CommentCreateInput{
		TaskID: body.TaskID,
		Author: body.Author,
		Text:   body.Text,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, x)
}

func (h *CommentHandler) Get(c *gin.Context) {
	x, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "comment not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.JSON(http.StatusOK, x)
}

func (h *CommentHandler) Update(c *gin.Context) {
	var body CommentUpdate
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	x, err := h.service.Update(c.Request.Context(), c.Param("id"), service.CommentUpdateInput{
		Author: body.Author,
		Text:   body.Text,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(httpx.StatusFor(httpx.CodeNotFound), httpx.Err(httpx.CodeNotFound, "comment not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, x)
}

func (h *CommentHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}
