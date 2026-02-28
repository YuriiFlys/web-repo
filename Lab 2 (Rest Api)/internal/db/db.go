package db

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"project-management/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func MustOpen() *gorm.DB {

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	ssl := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s TimeZone=UTC",
		host, port, name, user, pass, ssl,
	)

	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	database, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		log.Fatalf("db open failed: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("db() failed: %v", err)
	}

	sqlDB.SetMaxOpenConns(getEnvInt("DB_MAX_OPEN_CONNS", 10))
	sqlDB.SetMaxIdleConns(getEnvInt("DB_MAX_IDLE_CONNS", 5))
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err := database.AutoMigrate(&model.Project{}, &model.Task{}, &model.Comment{}); err != nil {
		log.Fatalf("db migrate failed: %v", err)
	}

	return database
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
