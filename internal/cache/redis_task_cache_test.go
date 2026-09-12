package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) (
	*miniredis.Miniredis,
	*RedisTaskListCache,
) {
	t.Helper()

	server := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	t.Cleanup(func() {
		_ = client.Close()
	})

	return server, NewRedisTaskListCache(client)
}

func TestRedisTaskListCache_GetSet(t *testing.T) {
	_, taskCache := newTestRedis(t)

	ctx := context.Background()

	value := []byte(`{"tasks":[],"total":0}`)

	if err := taskCache.Set(
		ctx,
		"task_manager:tasks:list:test",
		value,
		time.Minute,
	); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := taskCache.Get(
		ctx,
		"task_manager:tasks:list:test",
	)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if string(got) != string(value) {
		t.Fatalf(
			"Get() = %q, want %q",
			got,
			value,
		)
	}
}

func TestRedisTaskListCache_GetMiss(t *testing.T) {
	_, taskCache := newTestRedis(t)

	_, err := taskCache.Get(
		context.Background(),
		"task_manager:tasks:list:missing",
	)

	if err != ErrCacheMiss {
		t.Fatalf(
			"Get() error = %v, want %v",
			err,
			ErrCacheMiss,
		)
	}
}

func TestRedisTaskListCache_InvalidateTaskLists(t *testing.T) {
	server, taskCache := newTestRedis(t)

	ctx := context.Background()

	server.Set(
		"task_manager:tasks:list:first",
		"first",
	)

	server.Set(
		"task_manager:tasks:list:second",
		"second",
	)

	server.Set(
		"task_manager:other",
		"keep",
	)

	if err := taskCache.InvalidateTaskLists(ctx); err != nil {
		t.Fatalf(
			"InvalidateTaskLists() error = %v",
			err,
		)
	}

	if server.Exists("task_manager:tasks:list:first") {
		t.Fatal("first task-list cache key still exists")
	}

	if server.Exists("task_manager:tasks:list:second") {
		t.Fatal("second task-list cache key still exists")
	}

	if !server.Exists("task_manager:other") {
		t.Fatal("non-task-list cache key was deleted")
	}
}
