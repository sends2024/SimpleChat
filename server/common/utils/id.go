package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

var entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

func NewULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

// 锁KEY
func BuildLockKey(entity, id, action string) string {
	return fmt.Sprintf("lock:%s:%s:%s", entity, id, action)
}

func GenerateCode10() string {
	digits := "0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = digits[rand.Intn(10)]
	}
	return string(b)
}
