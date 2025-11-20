package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	rediscli "server/common/pkg/redis"
	"server/common/response"
	"server/common/utils"
	"server/dao"
	"server/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type ChangeChannelNamePayload struct {
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
}

type DeleteChannelPayload struct {
	ChannelID string `json:"channel_id"`
}

func CreateChannel(channelName, userID string) *response.AppError {
	exist := dao.GetChannelByName(channelName)
	if exist != nil {
		return response.NewAppError(http.StatusBadRequest, "ChannelName already exists")
	}

	channelID := utils.NewULID()
	dao.CreateChannel(&models.Channel{
		ID:        channelID,
		Name:      channelName,
		CreatedBy: userID,
	})

	user := dao.GetUserByID(userID)
	payload := JoinChannelPayload{
		Username:    user.ID,
		AvatarURL:   user.AvatarURL,
		ChannelID:   channelID,
		ChannelName: channelName,
		IsOwner:     false,
		UserID:      userID,
	}

	p, _ := json.Marshal(payload)

	evt := Envelope{
		TaskType: "JOIN",
		Payload:  p,
	}

	b, _ := json.Marshal(evt)
	rediscli.Rds.Publish(context.Background(), "channel_event", b)

	return nil
}

func DeleteChannel(channelID string) {
	dao.DeleteChannel(channelID)

	payload := DeleteChannelPayload{
		ChannelID: channelID,
	}

	p, _ := json.Marshal(payload)

	evt := Envelope{
		TaskType: "DELETE",
		Payload:  p,
	}

	b, _ := json.Marshal(evt)
	rediscli.Rds.Publish(context.Background(), "channel_event", b)
}

func RemoveMember(channelID, memberID string) *response.AppError {
	err := dao.RemoveMember(channelID, memberID)
	if err != nil {
		return response.NewAppError(http.StatusInternalServerError, "你永远是中国人")
	}

	payload := LeaveChannelPayload{
		ChannelID: channelID,
		UserID:    memberID,
	}

	p, _ := json.Marshal(payload)

	evt := Envelope{
		TaskType: "KICK",
		Payload:  p,
	}

	b, _ := json.Marshal(evt)
	rediscli.Rds.Publish(context.Background(), "channel_event", b)

	return nil
}

func GenerateInvite(channelID string) (string, error) {
	key := fmt.Sprintf("invite:channel:%s", channelID)

	code, err := rediscli.Rds.Get(context.Background(), key).Result()
	if err == redis.Nil {
		newCode := utils.GenerateCode10()
		err = rediscli.Rds.Set(context.Background(), key, newCode, 7*24*time.Hour).Err()
		if err != nil {
			return "", err
		}
		revKey := fmt.Sprintf("invite:code:%s", newCode)
		err = rediscli.Rds.Set(context.Background(), revKey, channelID, 7*24*time.Hour).Err()
		if err != nil {
			return "", err
		}

		return newCode, nil
	} else if err != nil {
		return "", err
	}

	return code, nil
}

func ChangeChannelName(newName string, channelID string) *response.AppError {
	exist := dao.GetChannelByName(newName)
	if exist != nil {
		return response.NewAppError(http.StatusBadRequest, "ChannelName already exists")
	}

	dao.UpdateChannelName(channelID, newName)

	payload := ChangeChannelNamePayload{
		ChannelID:   channelID,
		ChannelName: newName,
	}

	p, _ := json.Marshal(payload)

	evt := Envelope{
		TaskType: "CHANGE",
		Payload:  p,
	}

	b, _ := json.Marshal(evt)
	rediscli.Rds.Publish(context.Background(), "channel_event", b)

	return nil
}
