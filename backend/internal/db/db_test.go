package db

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

type stubSQLDB struct {
	maxOpen     int
	maxIdle     int
	maxLifetime time.Duration
	maxIdleTime time.Duration
}

func (s *stubSQLDB) SetMaxOpenConns(n int)              { s.maxOpen = n }
func (s *stubSQLDB) SetMaxIdleConns(n int)              { s.maxIdle = n }
func (s *stubSQLDB) SetConnMaxLifetime(d time.Duration) { s.maxLifetime = d }
func (s *stubSQLDB) SetConnMaxIdleTime(d time.Duration) { s.maxIdleTime = d }

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "pm")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "postgres")
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("DB_MAX_OPEN_CONNS", "20")
	t.Setenv("DB_MAX_IDLE_CONNS", "8")

	cfg := LoadConfigFromEnv()
	if cfg.Host != "localhost" || cfg.Port != "5432" || cfg.DBName != "pm" || cfg.User != "postgres" || cfg.Password != "postgres" || cfg.SSLMode != "disable" || cfg.MaxOpenConns != 20 || cfg.MaxIdleConns != 8 {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestBuildDSN(t *testing.T) {
	dsn := BuildDSN(Config{
		Host:     "localhost",
		Port:     "5432",
		DBName:   "pm",
		User:     "postgres",
		Password: "postgres",
		SSLMode:  "disable",
	})
	if !strings.Contains(dsn, "host=localhost") || !strings.Contains(dsn, "dbname=pm") || !strings.Contains(dsn, "sslmode=disable") {
		t.Fatalf("dsn = %q", dsn)
	}
}

func TestOpen(t *testing.T) {
	origOpenGorm := openGorm
	origGetSQLDB := getSQLDB
	origAutoMigrate := autoMigrate
	t.Cleanup(func() {
		openGorm = origOpenGorm
		getSQLDB = origGetSQLDB
		autoMigrate = origAutoMigrate
	})

	cfg := Config{Host: "localhost", Port: "5432", DBName: "pm", User: "postgres", Password: "postgres", SSLMode: "disable", MaxOpenConns: 11, MaxIdleConns: 7}

	t.Run("open error", func(t *testing.T) {
		openGorm = func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
			return nil, errors.New("open failed")
		}
		_, err := Open(cfg)
		if err == nil || err.Error() != "open failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("sql db error", func(t *testing.T) {
		openGorm = func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
			return &gorm.DB{}, nil
		}
		getSQLDB = func(database *gorm.DB) (sqlDB, error) {
			return nil, errors.New("db failed")
		}
		_, err := Open(cfg)
		if err == nil || err.Error() != "db failed" {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("migrate error", func(t *testing.T) {
		stub := &stubSQLDB{}
		openGorm = func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
			return &gorm.DB{}, nil
		}
		getSQLDB = func(database *gorm.DB) (sqlDB, error) { return stub, nil }
		autoMigrate = func(database *gorm.DB) error { return errors.New("migrate failed") }
		_, err := Open(cfg)
		if err == nil || err.Error() != "migrate failed" {
			t.Fatalf("err = %v", err)
		}
		if stub.maxOpen != 11 || stub.maxIdle != 7 || stub.maxLifetime != 30*time.Minute || stub.maxIdleTime != 5*time.Minute {
			t.Fatalf("stub = %+v", stub)
		}
	})

	t.Run("success", func(t *testing.T) {
		stub := &stubSQLDB{}
		db := &gorm.DB{}
		openGorm = func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
			return db, nil
		}
		getSQLDB = func(database *gorm.DB) (sqlDB, error) { return stub, nil }
		autoMigrate = func(database *gorm.DB) error { return nil }
		got, err := Open(cfg)
		if err != nil || got != db {
			t.Fatalf("got=%v err=%v", got, err)
		}
	})
}

func TestOpenFromEnv(t *testing.T) {
	origOpenGorm := openGorm
	origGetSQLDB := getSQLDB
	origAutoMigrate := autoMigrate
	t.Cleanup(func() {
		openGorm = origOpenGorm
		getSQLDB = origGetSQLDB
		autoMigrate = origAutoMigrate
	})

	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "pm")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "postgres")
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("DB_MAX_OPEN_CONNS", "12")
	t.Setenv("DB_MAX_IDLE_CONNS", "6")

	openGorm = func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	}
	getSQLDB = func(database *gorm.DB) (sqlDB, error) { return &stubSQLDB{}, nil }
	autoMigrate = func(database *gorm.DB) error { return nil }

	if _, err := OpenFromEnv(); err != nil {
		t.Fatalf("OpenFromEnv err = %v", err)
	}
}

func TestMustOpen(t *testing.T) {
	origOpenFromEnv := openFromEnv
	origFatalf := fatalf
	t.Cleanup(func() {
		openFromEnv = origOpenFromEnv
		fatalf = origFatalf
	})

	t.Run("success", func(t *testing.T) {
		want := &gorm.DB{}
		openFromEnv = func() (*gorm.DB, error) { return want, nil }
		got := MustOpen()
		if got != want {
			t.Fatalf("got = %v, want %v", got, want)
		}
	})

	t.Run("fatal on error", func(t *testing.T) {
		openFromEnv = func() (*gorm.DB, error) { return nil, errors.New("boom") }
		called := false
		fatalf = func(format string, args ...any) {
			called = true
			if !strings.Contains(format, "db open failed") {
				t.Fatalf("format = %q", format)
			}
		}
		if got := MustOpen(); got != nil {
			t.Fatalf("got = %v, want nil", got)
		}
		if !called {
			t.Fatal("expected fatalf to be called")
		}
	})
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
