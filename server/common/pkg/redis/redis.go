package rediscli

import (
	"os"

	"github.com/redis/go-redis/v9"
)

var Rds *redis.Client

func Init() {
	Rds = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}
