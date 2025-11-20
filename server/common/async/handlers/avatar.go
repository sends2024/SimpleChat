package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"server/common/pkg/redislock"
	"server/service"
)

type ChangeAvatarPayload struct {
	UserID string `json:"user_id"`
	NewURL string `json:"new_url"`
	Token  string `json:"token"`
}

func processChangeAvatarTask(p ChangeAvatarPayload) error {
	lockKey := fmt.Sprintf("lock:user:%s:change_avatar", p.UserID)
	token := p.Token

	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in ChangeAvatar worker: %v", r)
				err = fmt.Errorf("panic: %v", r)
			}
		}()

		return service.ChangeAvatar(p.UserID, p.NewURL)
	}()

	redislock.ReleaseLock(context.Background(), lockKey, token)
	if err != nil {
		log.Println("[Task] ChangeAvatar failed:", err)
		return err
	}

	return nil
}

func HandleChangeAvatar(_ context.Context, raw json.RawMessage) error {
	var p ChangeAvatarPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return err
	}
	return processChangeAvatarTask(p)
}
