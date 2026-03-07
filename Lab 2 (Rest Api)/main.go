// @title Project Management API
// @version 1.0
// @description REST API for projects, tasks, and comments.
// @BasePath /api
package main

import (
	"log"
	"net/http"
	"os"

	_ "project-management/docs"
	"project-management/internal/db"
	"project-management/internal/handler"
	"project-management/internal/middleware"
	"project-management/internal/repository"
	"project-management/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	database := db.MustOpen()

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")

	authHandler := handler.NewAuthHandler(service.NewAuthService(repository.NewAuthRepository(database)))
	authHandler.Register(api)

	protected := api.Group("/")
	protected.Use(middleware.JWTAuth())
	authHandler.RegisterProtected(protected)
	handler.NewUserHandler(service.NewUserService(repository.NewUserRepository(database))).Register(protected)
	handler.NewProjectHandler(service.NewProjectService(repository.NewProjectRepository(database))).Register(protected)
	handler.NewTaskHandler(service.NewTaskService(repository.NewTaskRepository(database))).Register(protected)
	handler.NewCommentHandler(service.NewCommentService(repository.NewCommentRepository(database))).Register(protected)

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}
	log.Println("Server running on :" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := os.Getenv("CORS_ORIGIN")
		if origin == "" {
			origin = "http://localhost:4200"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
