package floodControl

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"time"

	"github.com/redis/go-redis/v9"
)

var (
	FC_REDIS_HOST     = os.Getenv("FC_REDIS_HOST")
	FC_REDIS_PORT     = os.Getenv("FC_REDIS_PORT")
	FC_REDIS_PASSWORD = os.Getenv("FC_REDIS_PASSWORD")
	FC_RETENTION      = 5
	FC_MAXCHECKS      = 11
)

type fakeClock struct {
	fakeTime time.Time
}

func (f *fakeClock) Now() time.Time { return f.fakeTime }

func getNewConnection() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", FC_REDIS_HOST, FC_REDIS_PORT),
		Password: FC_REDIS_PASSWORD,
		DB:       0,
	})
}

func getNewRFC(retention uint64, checks uint64) *RedisFloodController {
	return &RedisFloodController{
		Client:           getNewConnection(),
		RetentionSeconds: retention,
		MaxChecks:        checks,
	}
}

func TestRedisFCPositive(t *testing.T) {
	var max uint64 = 5
	var ret uint64 = 11
	rfc := getNewRFC(ret, max)
	var uID int64 = 1000
	ctx := context.Background()
	defer rfc.Client.Del(ctx, strconv.Itoa(int(uID)))
	clock = &fakeClock{time.UnixMilli(int64(ret)*1000 + 1)}

	t.Run("single", func(t *testing.T) {
		res, err := rfc.Check(ctx, uID)
		if err != nil {
			t.Errorf("Error during TestRedisFCPositive: %s", err)
		}
		if res != true {
			t.Errorf("False somehow")
		}
	})
	t.Run("multiple1", func(t *testing.T) {
		for i := 0; i < int(max)-1; i++ {
			res, err := rfc.Check(ctx, uID)
			if err != nil {
				t.Errorf("Error during TestRedisFCPositive: %s", err)
			}
			if res != true {
				t.Errorf("False somehow")
			}
		}

	})
	clock = &fakeClock{time.UnixMilli(int64(ret)*1000 + int64(ret)*1000 + 1)}
	t.Run("multiple2", func(t *testing.T) {
		for i := 0; i < int(max); i++ {
			res, err := rfc.Check(ctx, uID)
			if err != nil {
				t.Errorf("Error during TestRedisFCPositive: %s", err)
			}
			if res != true {
				t.Errorf("False somehow")
			}
		}

	})

}

func TestRedisFCNegative(t *testing.T) {
	var max uint64 = 5
	var ret uint64 = 11
	rfc := getNewRFC(ret, max)
	var uID int64 = 1001
	ctx := context.Background()
	defer rfc.Client.Del(ctx, strconv.Itoa(int(uID)))
	clock = &fakeClock{time.UnixMilli(int64(ret)*1000 + 1)}

	t.Run("multiple1", func(t *testing.T) {
		for i := 0; i < int(max); i++ {
			res, err := rfc.Check(ctx, uID)
			if err != nil {
				t.Errorf("Error during TestRedisFCNegative: %s", err)
			}
			if res != true {
				t.Errorf("False somehow")
			}
		}
	})

	t.Run("multiple2", func(t *testing.T) {
		for i := 0; i < int(max); i++ {
			res, err := rfc.Check(ctx, uID)
			if err == nil {
				t.Errorf("wtf %v", res)
				return
			}
			if err.Error() != "max checks exceeded" {
				t.Errorf("Error during TestRedisFCNegative: %s", err)
			}
			if res == true {
				t.Errorf("true somehow")
			}
		}
	})
}

func TestRedisFCTwoClientsNegative(t *testing.T) {
	var max uint64 = 5
	var ret uint64 = 11
	rfc1 := getNewRFC(ret, max)
	rfc2 := getNewRFC(ret, max)
	var uID int64 = 1001
	ctx := context.Background()
	defer rfc1.Client.Del(ctx, strconv.Itoa(int(uID)))
	clock = &fakeClock{time.UnixMilli(int64(ret)*1000 + 1)}

	t.Run("multiple", func(t *testing.T) {
		ress := make(chan bool, max*2)
		errs := make(chan string, max*2)
		go func() {
			for i := 0; i < int(max); i++ {
				res, err := rfc1.Check(ctx, uID)
				ress <- res
				if err != nil {
					errs <- err.Error()
				} else {
					errs <- "nil"
				}
				time.Sleep(time.Millisecond * 1)
			}
		}()
		go func() {
			for i := 0; i < int(max); i++ {
				res, err := rfc2.Check(ctx, uID)
				ress <- res
				if err != nil {
					errs <- err.Error()
				} else {
					errs <- "nil"
				}
				time.Sleep(time.Millisecond * 2)
			}
		}()
		var bools [10]bool
		var errrs [10]string
		m := 0
		for i := 0; i < int(max)*2; i++ {
			bools[i] = <-ress
			errrs[i] = <-errs
			if bools[i] == true {
				m += 1
			}
		}

		if m != int(max) {
			t.Errorf("TwoClientsNegative failed: %d actual, %d expected %v %v", m, int(max), bools, errrs)
		}

	})

}
