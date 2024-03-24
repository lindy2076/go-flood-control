package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	fc "task/floodControl"

	"github.com/redis/go-redis/v9"
)

var (
	FC_REDIS_HOST     = os.Getenv("FC_REDIS_HOST")
	FC_REDIS_PORT     = os.Getenv("FC_REDIS_PORT")
	FC_REDIS_PASSWORD = os.Getenv("FC_REDIS_PASSWORD")
	FC_RETENTION      = StrToUInt(os.Getenv("FC_RETENTION"), 10)
	FC_MAXCHECKS      = StrToUInt(os.Getenv("FC_MAXCHECKS"), 10)
)

func StrToUInt(s string, def_val uint64) uint64 {
	if i, err := strconv.ParseUint(s, 10, 64); err == nil {
		return i
	}
	return def_val
}

func main() {
	fmt.Printf("Some debug info. Redis host: %s, port: %s, pswd: %s\n",
		FC_REDIS_HOST, FC_REDIS_PORT, FC_REDIS_PASSWORD)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", FC_REDIS_HOST, FC_REDIS_PORT),
		Password: FC_REDIS_PASSWORD,
		DB:       0,
	})
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}
	var rfc FloodControl = &fc.RedisFloodController{
		Client:           redisClient,
		RetentionSeconds: FC_RETENTION,
		MaxChecks:        FC_MAXCHECKS,
	}

	fmt.Printf("MaxChecks: %d\n", FC_MAXCHECKS)

	for i := 0; i < int(FC_MAXCHECKS)+1; i++ {
		v, err := rfc.Check(ctx, 0)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(v)
	}
}

// FloodControl интерфейс, который нужно реализовать.
// Рекомендуем создать директорию-пакет, в которой будет находиться реализация.
type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}
