package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"project-management/internal/auth"
	"project-management/internal/httpx"
	"project-management/internal/model"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct{ db *gorm.DB }

func NewAuthHandler(db *gorm.DB) *AuthHandler { return &AuthHandler{db: db} }

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

// RegisterUser godoc
// @Summary Register new user
// @Description Create a user account and return JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "Register payload"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /auth/register [post]
func (h *AuthHandler) RegisterUser(c *gin.Context) {
	var body RegisterRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Name = strings.TrimSpace(body.Name)

	var existing model.User
	if err := h.db.Where("email = ?", body.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, "email already in use"))
		return
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", "failed to hash password"))
		return
	}

	user := model.User{
		Email:        body.Email,
		Name:         body.Name,
		PasswordHash: string(hash),
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	token, err := auth.IssueToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", "failed to issue token"))
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

// Login godoc
// @Summary Login
// @Description Authenticate user and return JWT token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login payload"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} httpx.APIError
// @Failure 401 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var body LoginRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httpx.Err(httpx.CodeBadRequest, err.Error()))
		return
	}

	email := strings.TrimSpace(strings.ToLower(body.Email))

	var user model.User
	if err := h.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, httpx.Err(httpx.CodeUnauthorized, "invalid credentials"))
			return
		}
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", err.Error()))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, httpx.Err(httpx.CodeUnauthorized, "invalid credentials"))
		return
	}

	token, err := auth.IssueToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, httpx.Err("INTERNAL", "failed to issue token"))
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

// Me godoc
// @Summary Current user
// @Description Return the current authenticated user.
// @Tags Auth
// @Produce json
// @Success 200 {object} AuthUser
// @Failure 401 {object} httpx.APIError
// @Failure 500 {object} httpx.APIError
// @Router /auth/me [get]
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

	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
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
