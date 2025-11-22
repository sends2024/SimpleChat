package main

import (
	"context"
	"os/signal"
	"server/common/async"
	"server/common/pkg/db"
	rediscli "server/common/pkg/redis"
	"server/router"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")

	db.Init()
	rediscli.Init()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go async.StartWorker(ctx)

	r := router.SetupRouter()
	r.Run(":8080")
}
