package services

import (
	"shortleak/models"
	"shortleak/repositories"

	"github.com/google/uuid"
)

func GetLinksByUserID(userID uuid.UUID) ([]models.Link, error) {
	return repositories.GetAllLinksByUserID(userID)
}

func CreateLink(link *models.Link) error {
	return repositories.CreateLink(link)
}

func GetLinkByShortToken(shortToken string) (*models.Link, error) {
	return repositories.GetLinkByShortToken(shortToken)
}

func DeleteLink(shortToken string) error {
	return repositories.DeleteLink(shortToken)
}
