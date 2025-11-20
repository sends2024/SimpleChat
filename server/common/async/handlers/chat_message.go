package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"server/service"
)

type MessagePayload struct {
	ChannelID string    `json:"channel_id"`
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	SendTime  time.Time `json:"create_at"`
}

func processHandleChatMessageTask(p MessagePayload) error {
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in ChatMessage worker: %v", r)
				err = fmt.Errorf("panic: %v", r)
			}
		}()

		return service.SaveMessage(p.ChannelID, p.UserID, p.Message, p.SendTime)
	}()

	if err != nil {
		log.Println("[Task] ChatMessage failed:", err)
		return err
	}

	return nil
}

func HandleChatMessage(_ context.Context, raw json.RawMessage) error {
	var p MessagePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	return processHandleChatMessageTask(p)
}
