package async

import (
	"context"
	"encoding/json"
	"log"
	rediscli "server/common/pkg/redis"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	StreamName = "async_tasks"
	GroupName  = "async_group"
	RetryIdle  = 30 * time.Second
	ConsumerID = "worker"
)

func initConsumerGroup(ctx context.Context) {
	err := rediscli.Rds.XGroupCreateMkStream(ctx, StreamName, GroupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Println("XGroupCreate error:", err)
	} else {
		log.Println("Consumer Group ready")
	}
}

// retry机制
func retryPending(ctx context.Context) {
	res, err := rediscli.Rds.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: StreamName,
		Group:  GroupName,
		Count:  10,
		Start:  "-",
		End:    "+",
	}).Result()
	if err != nil {
		log.Println("XPENDING error:", err)
		return
	}

	for _, p := range res {
		if p.Idle < RetryIdle {
			continue
		}

		msgs, err := rediscli.Rds.XClaim(ctx, &redis.XClaimArgs{
			Stream:   StreamName,
			Group:    GroupName,
			Consumer: ConsumerID,
			MinIdle:  RetryIdle,
			Messages: []string{p.ID},
		}).Result()

		if err != nil {
			log.Println("XCLAIM error:", err)
			continue
		}

		// 对每条消息重试执行
		for _, msg := range msgs {
			raw, ok := msg.Values["data"].(string)
			if !ok {
				continue
			}

			var envelope Envelope
			if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
				continue
			}

			err := DispatchTask(context.Background(), envelope)
			if err != nil {
				log.Printf("Pending retry failed: %v", err)
				continue
			}

			// ACK
			ackErr := rediscli.Rds.XAck(context.Background(), StreamName, GroupName, msg.ID).Err()
			if ackErr != nil {
				log.Println("XACK failed:", ackErr)
			}
		}
	}
}

func pollPendingLoop(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			retryPending(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func StartWorker(ctx context.Context) {
	initConsumerGroup(ctx)
	go pollPendingLoop(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		default:
		}

		wctx, cancel := context.WithTimeout(ctx, 5*time.Second)

		res, err := rediscli.Rds.XReadGroup(wctx, &redis.XReadGroupArgs{
			Group:    GroupName,
			Consumer: ConsumerID,
			Streams:  []string{StreamName, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()

		cancel()

		if err != nil {
			if err == context.DeadlineExceeded || err == redis.Nil {
				continue
			}
			log.Println("XReadGroup error:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, stream := range res {
			for _, msg := range stream.Messages {

				raw, ok := msg.Values["data"].(string)
				if !ok {
					log.Println("Invalid msg: missing 'data'")
					continue
				}

				var envelope Envelope
				if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
					log.Println("Invalid JSON:", err)
					continue
				}

				err := DispatchTask(ctx, envelope)
				if err != nil {
					log.Printf("Task failed (will retry later): %v", err)
					continue
				}

				if ackErr := rediscli.Rds.XAck(ctx, StreamName, GroupName, msg.ID).Err(); ackErr != nil {
					log.Println("XACK failed:", ackErr)
				}
			}
		}
	}
}
