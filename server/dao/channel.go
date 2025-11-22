package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"server/common/pkg/db"
	rediscli "server/common/pkg/redis"
	"server/models"
	"time"

	"github.com/redis/go-redis/v9"
)

func GetChannelByID(id string) *models.Channel {
	var channel models.Channel
	if err := db.DB.Where("id = ?", id).First(&channel).Error; err != nil {
		return nil
	}
	return &channel
}

func IsMember(channelID, userID string) bool {
	var count int64
	db.DB.Model(&models.ChannelMember{}).
		Where("channel_id = ? AND user_id = ?", channelID, userID).
		Count(&count)
	return count > 0
}

func AddMember(channelID string, userID string) error {
	member := models.ChannelMember{
		ChannelID: channelID,
		UserID:    userID,
	}
	return db.DB.Create(&member).Error
}

func RemoveMember(channelID string, userID string) error {
	return db.DB.Where("channel_id = ? AND user_id = ?", channelID, userID).
		Delete(&models.ChannelMember{}).Error
}

func GetChannelIDsByUser(userID string) ([]string, error) {
	var links []models.ChannelMember

	if err := db.DB.Where("user_id = ?", userID).Find(&links).Error; err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(links))
	for _, m := range links {
		ids = append(ids, m.ChannelID)
	}

	return ids, nil
}

func GetChannelsByIDs(ids []string) ([]models.Channel, error) {
	var channels []models.Channel

	err := db.DB.
		Where("id IN ?", ids).
		Where("visible = ?", true).
		Find(&channels).Error

	if err != nil {
		return nil, err
	}

	return channels, nil
}

func GetMemberUserIDs(channelID string) ([]string, error) {
	var members []models.ChannelMember
	if err := db.DB.Where("channel_id = ?", channelID).Find(&members).Error; err != nil {
		return nil, err
	}
	userIDs := make([]string, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}
	return userIDs, nil
}

func GetUsersByIDs(userIDs []string) ([]models.User, error) {
	var users []models.User
	err := db.DB.Where("id IN ?", userIDs).Find(&users).Error
	return users, err
}

// 创建频道信息
func CreateMessage(msg *models.Message) {
	db.DB.Create(msg)
}

// 查询历史信息
func GetCachedMessages(channelID string) ([]models.Message, error) {
	ctx := context.Background()
	key := fmt.Sprintf("chat:history:%s", channelID)

	list, err := rediscli.Rds.LRange(ctx, key, 0, 99).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	msgs := make([]models.Message, 0, len(list))
	for _, item := range list {
		var msg models.Message
		if err := json.Unmarshal([]byte(item), &msg); err == nil {
			msgs = append(msgs, msg)
		}
	}

	reverseMessages(msgs)
	return msgs, nil
}

func GetOldMessages(channelID string, before time.Time) ([]models.Message, error) {
	var msgs []models.Message

	err := db.DB.
		Where("channel_id = ? AND sent_at < ?", channelID, before).
		Order("sent_at DESC").
		Limit(100).
		Find(&msgs).Error

	if err != nil {
		return nil, err
	}

	reverseMessages(msgs)
	return msgs, nil
}

func reverseMessages(m []models.Message) {
	for i, j := 0, len(m)-1; i < j; i, j = i+1, j-1 {
		m[i], m[j] = m[j], m[i]
	}
}
