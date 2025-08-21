package database

import (
	"gorm.io/gorm"
)

type Migrator interface {
	Migrate(db *gorm.DB) error
	RollbackAll(db *gorm.DB) error
	Seed() error
	HasMigrationsTable() bool
}

type DefaultMigrator struct{}

func (m *DefaultMigrator) Migrate(db *gorm.DB) error {
	return Migrate(db)
}

func (m *DefaultMigrator) RollbackAll(db *gorm.DB) error {
	return RollbackAll(db)
}

func (m *DefaultMigrator) Seed() error {
	return Seed()
}

func (m *DefaultMigrator) HasMigrationsTable() bool {
	return DB.Migrator().HasTable("migrations")
}
