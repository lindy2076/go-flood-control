package floodControl

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisFloodController struct {
	Client           *redis.Client
	RetentionSeconds uint64
	MaxChecks        uint64
}

func (rfc *RedisFloodController) Check(ctx context.Context, userID int64) (bool, error) {
	timestamp := time.Now().UnixMilli()
	fromTimestamp := timestamp - int64(rfc.RetentionSeconds)*1000

	_, err := rfc.Client.TSAddWithArgs(ctx, strconv.Itoa(int(userID)), timestamp, 1, &redis.TSOptions{
		DuplicatePolicy: redis.Sum.String(),
		Retention:       int(rfc.RetentionSeconds) * 1000,
	}).Result()
	if err != nil {
		return false, err
	}

	val, err := rfc.Client.TSRevRangeWithArgs(ctx, strconv.Itoa(int(userID)),
		int(fromTimestamp), int(timestamp), &redis.TSRevRangeOptions{
			Aggregator:     redis.Sum,
			BucketDuration: int(rfc.RetentionSeconds) * 1000,
		}).Result()
	if err != nil {
		return false, err
	}
	if len(val) == 0 {
		return true, fmt.Errorf("unexpected: no clicks at all")
	}

	var lastChunk redis.TSTimestampValue = val[0]
	fmt.Println(lastChunk.Value, rfc.MaxChecks)
	if uint64(lastChunk.Value) > rfc.MaxChecks {
		return false, fmt.Errorf("max checks exceeded")
	}

	return true, nil
}
