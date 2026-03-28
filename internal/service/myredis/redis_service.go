package myredis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mychat_server/internal/config"
	"mychat_server/pkg/script"
	"mychat_server/pkg/utils/zlog"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client
var ctx = context.Background()

func init() {
	conf := config.GetConfig()
	addr := conf.RedisConfig.Host + ":" + strconv.Itoa(conf.RedisConfig.Port)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: conf.RedisConfig.Password,
		DB:       conf.RedisConfig.Db,
	})
}

func SetKeyEx(key string, value string, timeout time.Duration) error {
	return redisClient.Set(ctx, key, value, timeout).Err()
}

func GetKey(key string) (string, error) {
	value, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			zlog.Info("key does not exist")
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func GetKeyNilIsErr(key string) (string, error) {
	return redisClient.Get(ctx, key).Result()
}

func GetKeyWithPrefixNilIsErr(prefix string) (string, error) {
	keys, err := redisClient.Keys(ctx, prefix+"*").Result()
	if err != nil {
		return "", err
	}
	if len(keys) == 0 {
		return "", redis.Nil
	}
	if len(keys) > 1 {
		return "", fmt.Errorf("multiple keys matched prefix %q", prefix)
	}
	return keys[0], nil
}

func GetKeyWithSuffixNilIsErr(suffix string) (string, error) {
	keys, err := redisClient.Keys(ctx, "*"+suffix).Result()
	if err != nil {
		return "", err
	}
	if len(keys) == 0 {
		return "", redis.Nil
	}
	if len(keys) > 1 {
		return "", fmt.Errorf("multiple keys matched suffix %q", suffix)
	}
	return keys[0], nil
}

func DelKeyIfExists(key string) error {
	exists, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists == 1 {
		return redisClient.Del(ctx, key).Err()
	}
	return nil
}

// ConsumeKeyOnceAtomic 使用 Lua 原子地消费一个 key：
// 存在则删除并返回 true，不存在返回 false。
func ConsumeKeyOnceAtomic(key string) (bool, error) {
	result, err := redisClient.Eval(ctx, script.ConsumeKeyOnceAtomic, []string{key}).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func DelKeysWithPattern(pattern string) error {
	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		log.Println("no key found for pattern", pattern)
		return nil
	}
	_, err = redisClient.Del(ctx, keys...).Result()
	return err
}

func DelKeysWithPrefix(prefix string) error {
	return DelKeysWithPattern(prefix + "*")
}

func DelKeysWithSuffix(suffix string) error {
	return DelKeysWithPattern("*" + suffix)
}

func DeleteAllRedisKeys() error {
	var cursor uint64
	for {
		keys, nextCursor, err := redisClient.Scan(ctx, cursor, "*", 0).Result()
		if err != nil {
			return err
		}
		cursor = nextCursor

		if len(keys) > 0 {
			if _, err := redisClient.Del(ctx, keys...).Result(); err != nil {
				return err
			}
		}

		if cursor == 0 {
			break
		}
	}
	return nil
}
