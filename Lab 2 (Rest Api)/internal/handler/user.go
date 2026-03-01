package handler

import (
	"net/http"

	"project-management/internal/httpx"
	"project-management/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct{ db *gorm.DB }

func NewUserHandler(db *gorm.DB) *UserHandler { return &UserHandler{db: db} }

type UserSummary struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *UserHandler) Register(r *gin.RouterGroup) {
	r.GET("/users", h.List)
}

// List godoc
// @Summary List users
// @Description List users for assignment.
// @Tags Users
// @Produce json
// @Success 200 {object} UsersListResponse
// @Failure 500 {object} httpx.APIError
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
	var users []model.User
	if err := h.db.Model(&model.User{}).Find(&users).Error; err != nil {
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
