package storage

import (
	"sync"
	"time"
)

type MemoryStorage struct {
	mu          sync.RWMutex
	counters    map[string]int64
	blocks      map[string]time.Time
	tokenLimits map[string]int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		counters:    make(map[string]int64),
		blocks:      make(map[string]time.Time),
		tokenLimits: make(map[string]int),
	}
}

func (m *MemoryStorage) Increment(key string, expiration time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if blockTime, exists := m.blocks[key]; exists {
		if time.Now().Before(blockTime) {
			return -1, nil
		}
		delete(m.blocks, key)
	}

	m.counters[key]++
	return m.counters[key], nil
}

func (m *MemoryStorage) SetBlock(key string, duration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.blocks[key] = time.Now().Add(duration)
	return nil
}

func (m *MemoryStorage) IsBlocked(key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	blockTime, exists := m.blocks[key]
	if !exists {
		return false, nil
	}

	if time.Now().Before(blockTime) {
		return true, nil
	}

	return false, nil
}

func (m *MemoryStorage) Reset(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.counters, key)
	delete(m.blocks, key)
	return nil
}

func (m *MemoryStorage) GetTokenLimit(token string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit, exists := m.tokenLimits[token]; exists {
		return limit, nil
	}
	return 0, nil
}

func (m *MemoryStorage) SetTokenLimit(token string, limit int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokenLimits[token] = limit
	return nil
}
