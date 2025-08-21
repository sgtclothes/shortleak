package repositories

import (
	"shortleak/database"
	"shortleak/models"
)

func GetAllUsers() ([]models.User, error) {
	var users []models.User
	result := database.DB.Find(&users)
	return users, result.Error
}

func CreateUser(user *models.User) error {
	result := database.DB.Create(user)
	return result.Error
}
