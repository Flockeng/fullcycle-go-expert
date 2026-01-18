//go:build integration
// +build integration

package integration

import (
	"os"
	"testing"
	"time"

	"rate-limiter/config"
	"rate-limiter/limiter"
	"rate-limiter/storage"
)

func TestRedisStorage_Integration(t *testing.T) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisStorage, err := storage.NewRedisStorage(redisHost, redisPort, "", 0)
	if err != nil {
		t.Skipf("Redis não disponível: %v", err)
	}
	defer redisStorage.Close()

	key := "integration:test:1"
	redisStorage.Reset(key)

	count, err := redisStorage.Increment(key, time.Second)
	if err != nil {
		t.Fatalf("Erro ao incrementar: %v", err)
	}
	if count != 1 {
		t.Errorf("Esperado count 1, obteve %d", count)
	}

	err = redisStorage.SetBlock(key, 5*time.Second)
	if err != nil {
		t.Fatalf("Erro ao bloquear: %v", err)
	}

	blocked, err := redisStorage.IsBlocked(key)
	if err != nil {
		t.Fatalf("Erro ao verificar bloqueio: %v", err)
	}
	if !blocked {
		t.Error("Chave deveria estar bloqueada")
	}

	err = redisStorage.Reset(key)
	if err != nil {
		t.Fatalf("Erro ao resetar: %v", err)
	}

	blocked, err = redisStorage.IsBlocked(key)
	if err != nil {
		t.Fatalf("Erro ao verificar bloqueio: %v", err)
	}
	if blocked {
		t.Error("Chave não deveria estar bloqueada após reset")
	}
}

func TestLimiterWithRedis_Integration(t *testing.T) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisStorage, err := storage.NewRedisStorage(redisHost, redisPort, "", 0)
	if err != nil {
		t.Skipf("Redis não disponível: %v", err)
	}
	defer redisStorage.Close()

	cfg := &config.Config{
		RateLimitIP:          5,
		RateLimitIPBlockTime: 5 * time.Second,
	}

	limiter := limiter.NewLimiter(redisStorage, cfg)
	ip := "192.168.1.100"

	redisStorage.Reset("ip:" + ip)

	for i := 0; i < 5; i++ {
		result, err := limiter.CheckIPLimit(ip)
		if err != nil {
			t.Fatalf("Erro inesperado: %v", err)
		}
		if !result.Allowed {
			t.Errorf("Requisição %d deveria ser permitida", i+1)
		}
	}

	result, err := limiter.CheckIPLimit(ip)
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}
	if result.Allowed {
		t.Error("6ª requisição deveria ser bloqueada")
	}
}

func TestTokenLimitsWithRedis_Integration(t *testing.T) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisStorage, err := storage.NewRedisStorage(redisHost, redisPort, "", 0)
	if err != nil {
		t.Skipf("Redis não disponível: %v", err)
	}
	defer redisStorage.Close()

	token := "integration-token-123"
	err = redisStorage.SetTokenLimit(token, 3)
	if err != nil {
		t.Fatalf("Erro ao definir limite do token: %v", err)
	}

	cfg := &config.Config{
		RateLimitTokenDefault:   10,
		RateLimitTokenBlockTime: 5 * time.Second,
	}

	limiter := limiter.NewLimiter(redisStorage, cfg)

	redisStorage.Reset("token:" + token)

	for i := 0; i < 3; i++ {
		result, err := limiter.CheckTokenLimit(token)
		if err != nil {
			t.Fatalf("Erro inesperado: %v", err)
		}
		if !result.Allowed {
			t.Errorf("Requisição %d deveria ser permitida", i+1)
		}
	}

	result, err := limiter.CheckTokenLimit(token)
	if err != nil {
		t.Fatalf("Erro inesperado: %v", err)
	}
	if result.Allowed {
		t.Error("4ª requisição deveria ser bloqueada")
	}
}
