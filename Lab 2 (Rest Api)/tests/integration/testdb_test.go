//go:build integration

package integration_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"project-management/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := testDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("integration database unavailable: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db.DB(): %v", err)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)
	sqlDB.SetConnMaxLifetime(time.Minute)

	if err := sqlDB.Ping(); err != nil {
		t.Skipf("integration database ping failed: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Project{}, &model.Task{}, &model.Comment{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func resetTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec("TRUNCATE TABLE comments, tasks, projects, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

func testDSN() string {
	if dsn := os.Getenv("TEST_DB_DSN"); dsn != "" {
		return dsn
	}

	host := getenvDefault("TEST_DB_HOST", getenvDefault("DB_HOST", "localhost"))
	port := getenvDefault("TEST_DB_PORT", getenvDefault("DB_PORT", "5432"))
	name := getenvDefault("TEST_DB_NAME", getenvDefault("DB_NAME", "project-management-test"))
	user := getenvDefault("TEST_DB_USER", getenvDefault("DB_USER", "postgres"))
	pass := getenvDefault("TEST_DB_PASSWORD", getenvDefault("DB_PASSWORD", "postgres"))
	ssl := getenvDefault("TEST_DB_SSLMODE", getenvDefault("DB_SSLMODE", "disable"))

	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s TimeZone=UTC", host, port, name, user, pass, ssl)
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
