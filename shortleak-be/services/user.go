package services

import (
	"shortleak/models"
	"shortleak/repositories"
)

func GetUsers() ([]models.User, error) {
	return repositories.GetAllUsers()
}

func AddUser(user *models.User) error {
	return repositories.CreateUser(user)
}
