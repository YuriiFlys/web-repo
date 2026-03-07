package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"project-management/internal/httpx"
	"project-management/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct{ service service.AuthService }

func NewAuthHandler(service service.AuthService) *AuthHandler { return &AuthHandler{service: service} }

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthUser struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuthResponse struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

func (h *AuthHandler) Register(r *gin.RouterGroup) {
	r.POST("/auth/register", h.RegisterUser)
	r.POST("/auth/login", h.Login)
}

func (h *AuthHandler) RegisterProtected(r *gin.RouterGroup) {
	r.GET("/auth/me", h.Me)
}

func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var body RegisterRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	user, token, err := h.service.Register(c.Request.Context(), service.RegisterInput{
		Email:    body.Email,
		Password: body.Password,
		Name:     body.Name,
	})
	if errors.Is(err, service.ErrEmailInUse) {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, "email already in use"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User: AuthUser{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var body LoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	user, token, err := h.service.Login(c.Request.Context(), service.LoginInput{
		Email:    strings.TrimSpace(strings.ToLower(body.Email)),
		Password: body.Password,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			c.JSON(http.StatusUnauthorized, httpx.Err(httpx.CodeUnauthorized, "invalid credentials"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User: AuthUser{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userIDRaw, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, httpx.Err(httpx.CodeUnauthorized, "unauthorized"))
		return
	}

	userID, ok := userIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, httpx.Err(httpx.CodeUnauthorized, "unauthorized"))
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	c.JSON(http.StatusOK, AuthUser{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	})
}
