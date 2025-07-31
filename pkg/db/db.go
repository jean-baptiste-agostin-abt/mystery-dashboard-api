package db

import (
	"context"
	"fmt"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jibe0123/mysteryfactory/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
)

// DB wraps gorm.DB with additional functionality
type DB struct {
	*gorm.DB
}

// New creates a new GORM database connection
func New(dsn string) (*DB, error) {
	config := &gorm.Config{
		Logger:  logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time { return time.Now().UTC() },
	}

	var gormDB *gorm.DB
	operation := func() error {
		var err error
		gormDB, err = gorm.Open(mysql.Open(dsn), config)
		return err
	}
	b := backoff.NewExponentialBackOff()
	if err := backoff.Retry(operation, b); err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := gormDB.Use(tracing.NewPlugin()); err != nil {
		return nil, fmt.Errorf("failed to enable tracing: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: gormDB}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// AutoMigrate runs GORM auto-migrations for all models
func (db *DB) AutoMigrate() error {
	err := db.DB.AutoMigrate(
		&models.User{},
		&models.Video{},
		&models.VideoStats{},
		&models.VideoStatsSnapshot{},
		&models.PublicationJob{},
		&models.Tenant{},
		&models.Workspace{},
	)
	if err != nil {
		return fmt.Errorf("failed to run auto-migrations: %w", err)
	}
	return nil
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(fn func(*gorm.DB) error) error {
	return db.DB.Transaction(fn)
}

// Health checks the database connection health
func (db *DB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.PingContext(ctx)
}

// GetDB returns the underlying GORM DB instance
func (db *DB) GetDB() *gorm.DB {
	return db.DB
}

// Repository base struct for GORM repositories
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new base repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// WithTenant adds tenant filtering to queries
func (r *Repository) WithTenant(tenantID string) *gorm.DB {
	return r.db.Where("tenant_id = ?", tenantID)
}

// WithSoftDelete includes soft-deleted records in queries
func (r *Repository) WithSoftDelete() *gorm.DB {
	return r.db.Unscoped()
}

// Paginate adds pagination to queries
func (r *Repository) Paginate(limit, offset int) *gorm.DB {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return r.db.Limit(limit).Offset(offset)
}

// RunMigrations applies database migrations from the specified directory.
func RunMigrations(dsn string) error {
	m, err := migrate.New(
		"file://db/migrations",
		"mysql://"+dsn,
	)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
