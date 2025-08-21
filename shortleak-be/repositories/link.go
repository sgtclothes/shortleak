package repositories

import (
	"shortleak/database"
	"shortleak/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetAllLinksByUserID(userID uuid.UUID) ([]models.Link, error) {
	var links []models.Link
	result := database.DB.Where("user_id = ?", userID).Find(&links)
	return links, result.Error
}

func CreateLink(link *models.Link) error {
	result := database.DB.Create(link)
	return result.Error
}

func GetLinkByShortToken(shortToken string) (*models.Link, error) {
	var link models.Link
	result := database.DB.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "fullname", "email")
	}).First(&link, "short_token = ?", shortToken)
	return &link, result.Error
}

func DeleteLink(shortToken string) error {
	result := database.DB.Where("short_token = ?", shortToken).Delete(&models.Link{})
	return result.Error
}
