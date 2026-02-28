// @title Project Management API
// @version 1.0
// @description REST API for projects, tasks, and comments.
// @BasePath /api
package main

import (
	"log"
	"os"

	_ "project-management/docs"
	"project-management/internal/db"
	"project-management/internal/handler"

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

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")

	handler.NewProjectHandler(database).Register(api)
	handler.NewTaskHandler(database).Register(api)
	handler.NewCommentHandler(database).Register(api)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running at http://localhost:" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
