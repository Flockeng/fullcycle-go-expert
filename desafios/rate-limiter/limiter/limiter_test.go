package limiter

import (
	"testing"
	"time"

	"rate-limiter/config"
	"rate-limiter/storage"
)

func TestCheckIPLimit(t *testing.T) {
	memStorage := storage.NewMemoryStorage()
	cfg := &config.Config{
		RateLimitIP:          5,
		RateLimitIPBlockTime: 5 * time.Second,
	}

	limiter := NewLimiter(memStorage, cfg)
	ip := "192.168.1.1"

	for i := 0; i < 5; i++ {
		result, err := limiter.CheckIPLimit(ip)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	result, err := limiter.CheckIPLimit(ip)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Allowed {
		t.Error("6th request should be blocked")
	}
	if result.Reason != "limit_exceeded" {
		t.Errorf("Expected reason 'limit_exceeded', got '%s'", result.Reason)
	}

	result, err = limiter.CheckIPLimit(ip)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Allowed {
		t.Error("Request should be blocked")
	}
	if result.Reason != "blocked" {
		t.Errorf("Expected reason 'blocked', got '%s'", result.Reason)
	}
}

func TestCheckTokenLimit(t *testing.T) {
	memStorage := storage.NewMemoryStorage()
	cfg := &config.Config{
		RateLimitTokenDefault:   10,
		RateLimitTokenBlockTime: 5 * time.Second,
	}

	limiter := NewLimiter(memStorage, cfg)
	token := "test-token-123"

	for i := 0; i < 10; i++ {
		result, err := limiter.CheckTokenLimit(token)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	result, err := limiter.CheckTokenLimit(token)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Allowed {
		t.Error("11th request should be blocked")
	}
}

func TestCheckTokenLimitWithCustomLimit(t *testing.T) {
	memStorage := storage.NewMemoryStorage()

	memStorage.SetTokenLimit("custom-token", 3)

	cfg := &config.Config{
		RateLimitTokenDefault:   10,
		RateLimitTokenBlockTime: 5 * time.Second,
	}

	limiter := NewLimiter(memStorage, cfg)

	for i := 0; i < 3; i++ {
		result, err := limiter.CheckTokenLimit("custom-token")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	result, err := limiter.CheckTokenLimit("custom-token")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result.Allowed {
		t.Error("4th request should be blocked")
	}
}

func TestDifferentIPs(t *testing.T) {
	memStorage := storage.NewMemoryStorage()
	cfg := &config.Config{
		RateLimitIP:          5,
		RateLimitIPBlockTime: 5 * time.Second,
	}

	limiter := NewLimiter(memStorage, cfg)
	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	for i := 0; i < 5; i++ {
		result, err := limiter.CheckIPLimit(ip1)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Allowed {
			t.Errorf("IP1 request %d should be allowed", i+1)
		}
	}

	for i := 0; i < 5; i++ {
		result, err := limiter.CheckIPLimit(ip2)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Allowed {
			t.Errorf("IP2 request %d should be allowed", i+1)
		}
	}

	result1, _ := limiter.CheckIPLimit(ip1)
	if result1.Allowed {
		t.Error("IP1 should be blocked")
	}

	result2, _ := limiter.CheckIPLimit(ip2)
	if result2.Allowed {
		t.Error("IP2 should also be blocked after 5 requests")
	}

	ip3 := "192.168.1.3"
	result3, err := limiter.CheckIPLimit(ip3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result3.Allowed {
		t.Error("IP3 should be allowed (different IP)")
	}
}
