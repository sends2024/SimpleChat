package dao

import (
	"server/common/pkg/db"
	"server/models"
)

func GetChannelByName(channelName string) *models.Channel {
	var channel models.Channel
	if err := db.DB.Where("name = ?", channelName).First(&channel).Error; err != nil {
		return nil
	}
	return &channel
}

func CreateChannel(channel *models.Channel) {
	db.DB.Create(channel)
}

func UpdateChannelName(channelID string, newName string) error {
	return db.DB.
		Model(&models.Channel{}).
		Where("id = ?", channelID).
		Update("name", newName).
		Error
}

func DeleteChannel(channelID string) error {
	return db.DB.
		Model(&models.Channel{}).
		Where("id = ?", channelID).
		Update("visible", false).
		Error
}
