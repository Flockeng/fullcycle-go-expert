package storage

import (
	"testing"
	"time"
)

func TestMemoryStorage_Increment(t *testing.T) {
	storage := NewMemoryStorage()
	key := "test-key"

	count, err := storage.Increment(key, time.Second)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	count, err = storage.Increment(key, time.Second)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestMemoryStorage_Block(t *testing.T) {
	storage := NewMemoryStorage()
	key := "test-key"

	// Bloqueia por 1 segundo
	err := storage.SetBlock(key, time.Second)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	blocked, err := storage.IsBlocked(key)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !blocked {
		t.Error("Key should be blocked")
	}

	count, err := storage.Increment(key, time.Second)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != -1 {
		t.Errorf("Expected -1 (blocked), got %d", count)
	}
}

func TestMemoryStorage_Reset(t *testing.T) {
	storage := NewMemoryStorage()
	key := "test-key"

	storage.Increment(key, time.Second)
	storage.Increment(key, time.Second)

	storage.SetBlock(key, time.Second)

	err := storage.Reset(key)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	blocked, err := storage.IsBlocked(key)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if blocked {
		t.Error("Key should not be blocked after reset")
	}

	count, err := storage.Increment(key, time.Second)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1 after reset, got %d", count)
	}
}

func TestMemoryStorage_TokenLimits(t *testing.T) {
	storage := NewMemoryStorage()
	token := "test-token"

	err := storage.SetTokenLimit(token, 100)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	limit, err := storage.GetTokenLimit(token)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if limit != 100 {
		t.Errorf("Expected limit 100, got %d", limit)
	}

	limit, err = storage.GetTokenLimit("non-existent-token")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if limit != 0 {
		t.Errorf("Expected limit 0 for non-existent token, got %d", limit)
	}
}
