package handlers

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userContextKey contextKey = "user"

type rateCounter struct {
	count       int
	windowStart time.Time
}

var (
	rateMu       sync.Mutex
	rateCounters = map[string]rateCounter{}
)

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := parseBearerToken(r.Header.Get("Authorization"))
		if tokenString == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (any, error) {
			return []byte(h.cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(*AuthClaims)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid claims"})
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requireRole(allowed ...string) func(http.Handler) http.Handler {
	set := map[string]bool{}
	for _, role := range allowed {
		set[role] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, _ := r.Context().Value(userContextKey).(*AuthClaims)
			if claims == nil || !set[claims.Role] {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) RateLimitMiddleware(next http.Handler) http.Handler {
	limit := h.cfg.RateLimitRPM
	window := time.Minute
	if limit <= 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		}
		now := time.Now()

		rateMu.Lock()
		rc, ok := rateCounters[ip]
		if !ok || now.Sub(rc.windowStart) >= window {
			rc = rateCounter{count: 0, windowStart: now}
		}
		rc.count++
		rateCounters[ip] = rc
		count := rc.count
		resetIn := window - now.Sub(rc.windowStart)
		rateMu.Unlock()

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(maxInt(0, limit-count)))
		w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(resetIn.Seconds())))

		if count > limit {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
