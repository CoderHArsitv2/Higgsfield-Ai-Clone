package models

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// Migrate brings the schema in line with the structs in this package.
//
// The table list lives here so there is one place to add a new table, and each
// struct's `gorm` tags are the single source of truth for its columns and
// indexes -- there is no separate DDL to keep in sync.
//
// Worth knowing: AutoMigrate only ever adds. It creates missing tables,
// columns and indexes, but never drops a column, narrows a type, or removes an
// index that is no longer declared. Renaming a field therefore leaves the old
// column behind holding data, and anything destructive has to be done
// deliberately rather than by editing a struct.
func Migrate(ctx context.Context, db *gorm.DB, log *slog.Logger) error {
	if err := db.WithContext(ctx).AutoMigrate(tables()...); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}
	log.Info("database schema up to date", "tables", len(tables()))
	return nil
}

// AutoMigrate is the same thing without a logger, for tests.
func AutoMigrate(db *gorm.DB) error { return db.AutoMigrate(tables()...) }
