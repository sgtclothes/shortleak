package database

import (
	"shortleak/models"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var migrations = []*gormigrate.Migration{
	{
		ID: "first",
		Migrate: func(tx *gorm.DB) error {
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	},
	{
		ID: "20250818_user_migration",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.User{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("users")
		},
	},
	{
		ID: "20250818_link_migration",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Link{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("links")
		},
	},
	{
		ID: "20250818_log_migration",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(&models.Log{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("logs")
		},
	},
}

func Migrate(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)
	return m.Migrate()
}

func RollbackLast(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)
	return m.RollbackLast()
}

func RollbackAll(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)
	return m.RollbackTo("first")
}
