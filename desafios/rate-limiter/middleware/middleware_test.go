package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rate-limiter/config"
	"rate-limiter/limiter"
	"rate-limiter/storage"
)

func TestRateLimiterMiddleware_WithIP(t *testing.T) {
	memStorage := storage.NewMemoryStorage()
	cfg := &config.Config{
		RateLimitIP:          5,
		RateLimitIPBlockTime: 5 * time.Second,
	}

	rateLimiter := limiter.NewLimiter(memStorage, cfg)
	middleware := NewRateLimiterMiddleware(rateLimiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		middleware.Handler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d should return 200, got %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", rec.Code)
	}

	expectedBody := `{"error": "you have reached the maximum number of requests or actions allowed within a certain time frame"}`
	if rec.Body.String() != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, rec.Body.String())
	}
}

func TestRateLimiterMiddleware_WithToken(t *testing.T) {
	memStorage := storage.NewMemoryStorage()
	cfg := &config.Config{
		RateLimitTokenDefault:   3,
		RateLimitTokenBlockTime: 5 * time.Second,
	}

	rateLimiter := limiter.NewLimiter(memStorage, cfg)
	middleware := NewRateLimiterMiddleware(rateLimiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	token := "test-token-abc"

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", token)
		rec := httptest.NewRecorder()

		middleware.Handler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d should return 200, got %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("API_KEY", token)
	rec := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", rec.Code)
	}
}

func TestRateLimiterMiddleware_TokenOverridesIP(t *testing.T) {
	memStorage := storage.NewMemoryStorage()
	cfg := &config.Config{
		RateLimitIP:             2,
		RateLimitIPBlockTime:    5 * time.Second,
		RateLimitTokenDefault:   5,
		RateLimitTokenBlockTime: 5 * time.Second,
	}

	rateLimiter := limiter.NewLimiter(memStorage, cfg)
	middleware := NewRateLimiterMiddleware(rateLimiter)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	token := "override-token"

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", token)
		rec := httptest.NewRecorder()

		middleware.Handler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d with token should return 200 (using token limit), got %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("API_KEY", token)
	rec := httptest.NewRecorder()

	middleware.Handler(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429, got %d", rec.Code)
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name          string
		headers       map[string]string
		expectedToken string
	}{
		{
			name:          "API_KEY header",
			headers:       map[string]string{"API_KEY": "token123"},
			expectedToken: "token123",
		},
		{
			name:          "Authorization header with API_KEY prefix",
			headers:       map[string]string{"Authorization": "API_KEY token456"},
			expectedToken: "token456",
		},
		{
			name:          "No token",
			headers:       map[string]string{},
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			token := extractToken(req)
			if token != tt.expectedToken {
				t.Errorf("Expected token '%s', got '%s'", tt.expectedToken, token)
			}
		})
	}
}
