package storage

import "time"

type TokenLimiter interface {
	GetTokenLimit(token string) (int, error)
}

type Storage interface {
	Increment(key string, expiration time.Duration) (int64, error)
	SetBlock(key string, duration time.Duration) error
	IsBlocked(key string) (bool, error)
	Reset(key string) error
}
