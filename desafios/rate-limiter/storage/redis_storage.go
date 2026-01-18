package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStorage(host string, port string, password string, db int) (*RedisStorage, error) {
	addr := fmt.Sprintf("%s:%s", host, port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisStorage{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisStorage) Increment(key string, expiration time.Duration) (int64, error) {
	blocked, err := r.IsBlocked(key)
	if err != nil {
		return 0, err
	}
	if blocked {
		return -1, nil
	}

	pipe := r.client.Pipeline()
	incr := pipe.Incr(r.ctx, key)
	pipe.Expire(r.ctx, key, expiration)
	_, err = pipe.Exec(r.ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

func (r *RedisStorage) SetBlock(key string, duration time.Duration) error {
	blockKey := fmt.Sprintf("block:%s", key)
	return r.client.Set(r.ctx, blockKey, "1", duration).Err()
}

func (r *RedisStorage) IsBlocked(key string) (bool, error) {
	blockKey := fmt.Sprintf("block:%s", key)
	val, err := r.client.Exists(r.ctx, blockKey).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

func (r *RedisStorage) Reset(key string) error {
	pipe := r.client.Pipeline()
	pipe.Del(r.ctx, key)
	pipe.Del(r.ctx, fmt.Sprintf("block:%s", key))
	_, err := pipe.Exec(r.ctx)
	return err
}

func (r *RedisStorage) GetTokenLimit(token string) (int, error) {
	key := fmt.Sprintf("token_limit:%s", token)
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	limit, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return limit, nil
}

func (r *RedisStorage) SetTokenLimit(token string, limit int) error {
	key := fmt.Sprintf("token_limit:%s", token)
	return r.client.Set(r.ctx, key, limit, 0).Err()
}

func (r *RedisStorage) Close() error {
	return r.client.Close()
}
