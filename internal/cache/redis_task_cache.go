package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const taskListKeyPrefix = "task_manager:tasks:list:"

type RedisTaskListCache struct {
	client *redis.Client
}

func NewRedisTaskListCache(client *redis.Client) *RedisTaskListCache {
	return &RedisTaskListCache{
		client: client,
	}
}

func (c *RedisTaskListCache) Get(
	ctx context.Context,
	key string,
) ([]byte, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}

		return nil, err
	}

	return value, nil
}

func (c *RedisTaskListCache) Set(
	ctx context.Context,
	key string,
	value []byte,
	expiration time.Duration,
) error {
	return c.client.Set(
		ctx,
		key,
		value,
		expiration,
	).Err()
}

func (c *RedisTaskListCache) InvalidateTaskLists(
	ctx context.Context,
) error {
	var cursor uint64

	for {
		keys, nextCursor, err := c.client.Scan(
			ctx,
			cursor,
			taskListKeyPrefix+"*",
			100,
		).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	return nil
}
