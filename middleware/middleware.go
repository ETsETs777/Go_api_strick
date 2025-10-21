package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}
	
	return limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := rl.GetLimiter(ip)
		
		if !limiter.Allow() {
			http.Error(w, "Слишком много запросов. Попробуйте позже.", http.StatusTooManyRequests)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) CleanupOldVisitors() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			rl.mu.Lock()
			for ip, limiter := range rl.visitors {
				if limiter.Tokens() == float64(rl.burst) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

type RequestLogger struct {
	mu      sync.Mutex
	counter int
}

func NewRequestLogger() *RequestLogger {
	return &RequestLogger{}
}

func (rl *RequestLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		rl.mu.Lock()
		rl.counter++
		requestID := rl.counter
		rl.mu.Unlock()
		
		fmt.Printf("[%d] %s %s %s - Started\n", 
			requestID, time.Now().Format("15:04:05"), r.Method, r.RequestURI)
		
		next.ServeHTTP(w, r)
		
		fmt.Printf("[%d] %s %s %s - Completed in %v\n", 
			requestID, time.Now().Format("15:04:05"), r.Method, r.RequestURI, time.Since(start))
	})
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("PANIC: %v\n", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

