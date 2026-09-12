package service

import (
	"context"
	"errors"
	"time"
)

type mockTaskListCache struct {
	values map[string][]byte

	getFunc        func(context.Context, string) ([]byte, error)
	setFunc        func(context.Context, string, []byte, time.Duration) error
	invalidateFunc func(context.Context) error
}

func newMockTaskListCache() *mockTaskListCache {
	return &mockTaskListCache{
		values: make(map[string][]byte),
	}
}

func (m *mockTaskListCache) Get(
	ctx context.Context,
	key string,
) ([]byte, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, key)
	}

	value, ok := m.values[key]
	if !ok {
		return nil, errors.New("cache miss")
	}

	return value, nil
}

func (m *mockTaskListCache) Set(
	ctx context.Context,
	key string,
	value []byte,
	expiration time.Duration,
) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, key, value, expiration)
	}

	m.values[key] = value

	return nil
}

func (m *mockTaskListCache) InvalidateTaskLists(
	ctx context.Context,
) error {
	if m.invalidateFunc != nil {
		return m.invalidateFunc(ctx)
	}

	clear(m.values)

	return nil
}
