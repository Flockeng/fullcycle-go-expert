package main

import (
	"fmt"
	"log"
	"net/http"

	"rate-limiter/config"
	"rate-limiter/limiter"
	"rate-limiter/middleware"
	"rate-limiter/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	redisStorage, err := storage.NewRedisStorage(
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisPassword,
		cfg.RedisDB,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Redis storage: %v", err)
	}
	defer redisStorage.Close()

	rateLimiter := limiter.NewLimiter(redisStorage, cfg)
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(rateLimiter)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Request successful"}`))
	})

	mux := http.NewServeMux()
	mux.Handle("/", rateLimiterMiddleware.Handler(handler))

	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Rate Limit IP: %d req/s\n", cfg.RateLimitIP)
	fmt.Printf("Rate Limit Token Default: %d req/s\n", cfg.RateLimitTokenDefault)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
