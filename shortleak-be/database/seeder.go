package database

import (
	"fmt"
	"log"
	"shortleak/models"

	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
)

/** SeedUsersFromExcel seeds the database with users from an Excel file */
func SeedUsersFromExcel(filePath string) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatalf("❌ Failed to open file: %v", err)
	}

	/** Get all rows from the "Users" sheet */
	rows, err := f.GetRows("data")
	if err != nil {
		log.Fatalf("❌ Failed to get rows: %v", err)
	}

	/** Check if there are any rows, skip header */
	for i, row := range rows {
		if i == 0 {
			continue
		}

		if len(row) < 3 {
			fmt.Printf("⚠️ Row %d is incomplete, skipping...\n", i+1)
			continue
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(row[2]), 10)
		if err != nil {
			fmt.Printf("❌ Failed to hash password on row %d: %v\n", i+1, err)
			continue
		}

		user := models.User{
			FullName: row[0],
			Email:    row[1],
			Password: string(hashed),
		}

		if err := DB.Create(&user).Error; err != nil {
			fmt.Printf("❌ Failed to add user on row %d: %v\n", i+1, err)
		} else {
			fmt.Printf("✅ User added: %s\n", user.Email)
		}
	}
	return err
}

func Seed() error {
	/** Seed users from Excel file */
	if err := SeedUsersFromExcel("seeders/users.xlsx"); err != nil {
		log.Fatalf("❌ Failed to seed users: %v", err)
		return err
	}

	log.Println("✅ Database seeded successfully")
	return nil
}
