package limiter

import (
	"fmt"
	"time"

	"rate-limiter/config"
	"rate-limiter/storage"
)

type Limiter struct {
	storage storage.Storage
	config  *config.Config
}

type Result struct {
	Allowed bool
	Reason  string
}

func NewLimiter(storage storage.Storage, config *config.Config) *Limiter {
	return &Limiter{
		storage: storage,
		config:  config,
	}
}

func (l *Limiter) CheckLimit(identifier string, limit int, blockTime time.Duration) (*Result, error) {
	blocked, err := l.storage.IsBlocked(identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to check block status: %w", err)
	}
	if blocked {
		return &Result{
			Allowed: false,
			Reason:  "blocked",
		}, nil
	}

	count, err := l.storage.Increment(identifier, time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to increment counter: %w", err)
	}

	if count == -1 {
		return &Result{
			Allowed: false,
			Reason:  "blocked",
		}, nil
	}

	if count > int64(limit) {
		err = l.storage.SetBlock(identifier, blockTime)
		if err != nil {
			return nil, fmt.Errorf("failed to set block: %w", err)
		}

		return &Result{
			Allowed: false,
			Reason:  "limit_exceeded",
		}, nil
	}

	return &Result{
		Allowed: true,
		Reason:  "allowed",
	}, nil
}

func (l *Limiter) CheckIPLimit(ip string) (*Result, error) {
	return l.CheckLimit(
		fmt.Sprintf("ip:%s", ip),
		l.config.RateLimitIP,
		l.config.RateLimitIPBlockTime,
	)
}

func (l *Limiter) CheckTokenLimit(token string) (*Result, error) {
	var tokenLimit int
	if tokenLimiter, ok := l.storage.(storage.TokenLimiter); ok {
		customLimit, err := tokenLimiter.GetTokenLimit(token)
		if err == nil && customLimit > 0 {
			tokenLimit = customLimit
		} else {
			tokenLimit = l.config.RateLimitTokenDefault
		}
	} else {
		tokenLimit = l.config.RateLimitTokenDefault
	}

	return l.CheckLimit(
		fmt.Sprintf("token:%s", token),
		tokenLimit,
		l.config.RateLimitTokenBlockTime,
	)
}
