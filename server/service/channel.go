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

type ChannelDTO struct {
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	IsOwner     bool   `json:"is_owner"`
}

type MemberDTO struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type MessageDTO struct {
	SenderID string    `json:"sender_id"`
	Content  string    `json:"content"`
	SentAt   time.Time `json:"sent_at"`
}

type HistoryDTO struct {
	Messages []MessageDTO `json:"messages"`
	Cursor   string       `json:"cursor"`
}

type Envelope struct {
	TaskType string          `json:"task_type"`
	Payload  json.RawMessage `json:"payload"`
}

type JoinChannelPayload struct {
	Username    string `json:"username"`
	AvatarURL   string `json:"avatar_url"`
	ChannelID   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	IsOwner     bool   `json:"is_owner"`
	UserID      string `json:"user_id"`
}

type LeaveChannelPayload struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
}

func JoinChannel(inviteCode string, userID string) *response.AppError {
	revKey := fmt.Sprintf("invite:code:%s", inviteCode)
	channelID, err := rediscli.Rds.Get(context.Background(), revKey).Result()
	if err == redis.Nil {
		return response.NewAppError(http.StatusNotFound, "Invite not exist")
	}

	channel := dao.GetChannelByID(channelID)
	if channel == nil {
		return response.NewAppError(http.StatusNotFound, "Channel has been dissolved")
	}

	exists := dao.IsMember(channelID, userID)
	if exists {
		return response.NewAppError(http.StatusConflict, "Member already exists")
	}

	dao.AddMember(channelID, userID)

	user := dao.GetUserByID(userID)
	payload := JoinChannelPayload{
		Username:    user.ID,
		AvatarURL:   user.AvatarURL,
		ChannelID:   channel.ID,
		ChannelName: channel.Name,
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

func LeaveChannel(channelID string, userID string) *response.AppError {
	err := dao.RemoveMember(channelID, userID)
	if err != nil {
		return response.NewAppError(http.StatusInternalServerError, "你永远是中国人")
	}

	payload := LeaveChannelPayload{
		ChannelID: channelID,
		UserID:    userID,
	}

	p, _ := json.Marshal(payload)

	evt := Envelope{
		TaskType: "LEAVE",
		Payload:  p,
	}

	b, _ := json.Marshal(evt)
	rediscli.Rds.Publish(context.Background(), "channel_event", b)

	return nil
}

func GetChannels(userID string) []ChannelDTO {
	channelIDs, err := dao.GetChannelIDsByUser(userID)
	if err != nil || len(channelIDs) == 0 {
		return []ChannelDTO{}
	}

	channels, err := dao.GetChannelsByIDs(channelIDs)
	if err != nil {
		return []ChannelDTO{}
	}

	result := make([]ChannelDTO, 0, len(channels))
	for _, ch := range channels {
		result = append(result, ChannelDTO{
			ChannelID:   ch.ID,
			ChannelName: ch.Name,
			IsOwner:     ch.CreatedBy == userID,
		})
	}

	return result
}

func GetMembers(channelID string) []MemberDTO {
	userIDs, err := dao.GetMemberUserIDs(channelID)
	if err != nil || len(userIDs) == 0 {
		return []MemberDTO{}
	}

	users, err := dao.GetUsersByIDs(userIDs)
	if err != nil {
		return []MemberDTO{}
	}

	result := make([]MemberDTO, 0, len(users))
	for _, u := range users {
		result = append(result, MemberDTO{
			UserID:    u.ID,
			Username:  u.Name,
			AvatarURL: u.AvatarURL,
		})
	}

	return result
}

func SaveMessage(channelID, userID, message string, sendTime time.Time) error {
	msgID := utils.NewULID()
	dao.CreateMessage(&models.Message{
		ID:        msgID,
		ChannelID: channelID,
		SenderID:  userID,
		Content:   message,
		SentAt:    sendTime,
	})

	return nil
}

func GetHistory(before string, channelID string) HistoryDTO {
	var msgs []models.Message

	if before == "" {
		msgs, _ = dao.GetCachedMessages(channelID)
	} else {
		t, _ := time.Parse(time.RFC3339, before)

		msgs, _ = dao.GetOldMessages(channelID, t)
	}

	result := make([]MessageDTO, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, MessageDTO{
			SenderID: m.SenderID,
			Content:  m.Content,
			SentAt:   m.SentAt,
		})
	}

	cursor := ""
	if len(result) > 0 {
		cursor = result[0].SentAt.Format(time.RFC3339)
	}

	return HistoryDTO{
		Messages: result,
		Cursor:   cursor,
	}
}
