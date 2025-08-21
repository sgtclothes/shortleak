package database

import (
	"fmt"
	"log"
	"shortleak/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ConnectDBFunc = ConnectDB
	SeedFunc      = Seed
	openDB        = gorm.Open
)

var DB *gorm.DB

func ConnectDB(cfg config.Config) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port,
	)

	db, err := openDB(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db
	fmt.Println("✅ Database connected!")
}
