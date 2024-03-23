package floodControl

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisFloodController struct {
	Client           *redis.Client
	RetentionSeconds uint64
	MaxChecks        uint64
}

func (rfc *RedisFloodController) Check(ctx context.Context, userID int64) (bool, error) {
	err := rfc.Client.Set(ctx, "foo", "bar", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rfc.Client.Get(ctx, "foo").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("foo", val)
	return true, nil
}
