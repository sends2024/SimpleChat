package async

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"server/common/async/handlers"
	rediscli "server/common/pkg/redis"

	redis "github.com/redis/go-redis/v9"
)

type Envelope struct {
	TaskType string          `json:"task_type"`
	Payload  json.RawMessage `json:"payload"`
}

func EnqueueTask(taskType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	envelope := Envelope{
		TaskType: taskType,
		Payload:  payloadBytes,
	}
	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	// 入队
	_, err = rediscli.Rds.XAdd(context.Background(), &redis.XAddArgs{
		Stream: StreamName,
		Values: map[string]interface{}{
			"data": string(envelopeBytes),
		},
	}).Result()

	if err != nil {
		log.Printf("Failed to enqueue task")
		return err
	}

	return nil
}

func DispatchTask(ctx context.Context, envelope Envelope) error {
	switch envelope.TaskType {

	case "change_avatar":
		return handlers.HandleChangeAvatar(ctx, envelope.Payload)

	case "chat_message":
		return handlers.HandleChatMessage(ctx, envelope.Payload)

	default:
		log.Printf("unknown task type: %s", envelope.TaskType)

		return errors.New("unknown task type")
	}
}
