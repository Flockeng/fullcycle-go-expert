package middleware

import (
	"net/http"
	"strings"

	"rate-limiter/limiter"
)

type RateLimiterMiddleware struct {
	limiter *limiter.Limiter
}

func NewRateLimiterMiddleware(limiter *limiter.Limiter) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: limiter,
	}
}

func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)

		if token != "" {
			result, err := m.limiter.CheckTokenLimit(token)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if !result.Allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "you have reached the maximum number of requests or actions allowed within a certain time frame"}`))
				return
			}
		}

		if token == "" {
			ip := getClientIP(r)
			result, err := m.limiter.CheckIPLimit(ip)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if !result.Allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "you have reached the maximum number of requests or actions allowed within a certain time frame"}`))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func extractToken(r *http.Request) string {
	token := r.Header.Get("API_KEY")
	if token != "" {
		return strings.TrimSpace(token)
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "API_KEY ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "API_KEY "))
	}

	return ""
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-Ip")
	if ip != "" {
		return ip
	}

	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
		return ip
	}

	ip = r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
