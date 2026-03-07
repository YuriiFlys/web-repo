package handler

import (
	"net/http"

	"project-management/internal/httpx"
	"project-management/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{ service service.UserService }

func NewUserHandler(service service.UserService) *UserHandler { return &UserHandler{service: service} }

type UserSummary struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *UserHandler) Register(r *gin.RouterGroup) {
	r.GET("/users", h.List)
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	items := make([]UserSummary, 0, len(users))
	for _, u := range users {
		items = append(items, UserSummary{
			ID:    u.ID,
			Email: u.Email,
			Name:  u.Name,
		})
	}

	c.JSON(http.StatusOK, UsersListResponse{Items: items})
}

type UsersListResponse struct {
	Items []UserSummary `json:"items"`
}
