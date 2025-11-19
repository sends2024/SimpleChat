package dao

import (
	"server/common/pkg/db"
	"server/models"
)

func CreateUser(user *models.User) {
	db.DB.Create(user)
}

func GetUserByUsername(username string) *models.User {
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil
	}
	return &user
}

func GetUserByID(id string) *models.User {
	var user models.User
	db.DB.First(&user, id)
	return &user
}

func UpdatePassword(userID, newHashedPassword string) {
	db.DB.Model(&models.User{}).Where("id = ?", userID).Update("password", newHashedPassword)
}

func UpdateAvatar(url, userID string) {
	db.DB.Model(&models.User{}).Where("id = ?", userID).Update("avatar_url", url)
}
