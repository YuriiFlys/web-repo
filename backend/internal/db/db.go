package db

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"project-management/internal/config"
	"project-management/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host         string
	Port         string
	DBName       string
	User         string
	Password     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

type sqlDB interface {
	SetMaxOpenConns(n int)
	SetMaxIdleConns(n int)
	SetConnMaxLifetime(d time.Duration)
	SetConnMaxIdleTime(d time.Duration)
}

var (
	openGorm = func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
		return gorm.Open(dialector, cfg)
	}
	getSQLDB    = defaultGetSQLDB
	autoMigrate = defaultAutoMigrate
	openFromEnv = OpenFromEnv
	fatalf      = log.Fatalf
)

func defaultGetSQLDB(database *gorm.DB) (sqlDB, error) {
	return database.DB()
}

func defaultAutoMigrate(database *gorm.DB) error {
	return database.AutoMigrate(&model.User{}, &model.Project{}, &model.Task{}, &model.Comment{})
}

func parseDatabaseURL(rawURL string) (Config, error) {
	// DO іноді дає postgres:// замість postgresql://
	normalized := strings.Replace(rawURL, "postgres://", "postgresql://", 1)

	u, err := url.Parse(normalized)
	if err != nil {
		return Config{}, fmt.Errorf("url.Parse: %w", err)
	}

	if u.Host == "" {
		return Config{}, fmt.Errorf("empty host in DATABASE_URL")
	}

	password, _ := u.User.Password()

	port := u.Port()
	if port == "" {
		port = "5432"
	}

	dbName := strings.TrimPrefix(u.Path, "/")

	return Config{
		Host:         u.Hostname(),
		Port:         port,
		DBName:       dbName,
		User:         u.User.Username(),
		Password:     password,
		SSLMode:      "require",
		MaxOpenConns: config.GetEnvInt("DB_MAX_OPEN_CONNS", 10),
		MaxIdleConns: config.GetEnvInt("DB_MAX_IDLE_CONNS", 5),
	}, nil
}

func LoadConfigFromEnv() Config {
	if rawURL := os.Getenv("DATABASE_URL"); rawURL != "" {
		log.Printf("info: DATABASE_URL found, parsing...")
		cfg, err := parseDatabaseURL(rawURL)
		if err != nil {
			log.Printf("warn: failed to parse DATABASE_URL: %v", err)
		} else {
			log.Printf("info: DB config — host=%s port=%s db=%s user=%s",
				cfg.Host, cfg.Port, cfg.DBName, cfg.User)
			return cfg
		}
	} else {
		log.Printf("info: DATABASE_URL not found, using individual env vars")
	}

	return Config{
		Host:         os.Getenv("DB_HOST"),
		Port:         os.Getenv("DB_PORT"),
		DBName:       os.Getenv("DB_NAME"),
		User:         os.Getenv("DB_USER"),
		Password:     os.Getenv("DB_PASSWORD"),
		SSLMode:      os.Getenv("DB_SSLMODE"),
		MaxOpenConns: config.GetEnvInt("DB_MAX_OPEN_CONNS", 10),
		MaxIdleConns: config.GetEnvInt("DB_MAX_IDLE_CONNS", 5),
	}
}

func BuildDSN(cfg Config) string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.Port, cfg.DBName, cfg.User, cfg.Password, cfg.SSLMode,
	)
}

func OpenFromEnv() (*gorm.DB, error) {
	return Open(LoadConfigFromEnv())
}

func Open(cfg Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	database, err := openGorm(postgres.Open(BuildDSN(cfg)), gormCfg)
	if err != nil {
		return nil, err
	}

	sqlConn, err := getSQLDB(database)
	if err != nil {
		return nil, err
	}

	sqlConn.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlConn.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlConn.SetConnMaxLifetime(30 * time.Minute)
	sqlConn.SetConnMaxIdleTime(5 * time.Minute)

	if err := autoMigrate(database); err != nil {
		return nil, err
	}

	return database, nil
}

func MustOpen() *gorm.DB {
	database, err := openFromEnv()
	if err != nil {
		fatalf("db open failed: %v", err)
	}
	return database
}
