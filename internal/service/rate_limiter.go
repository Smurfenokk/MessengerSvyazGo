package service

import (
	"context"
	"sync"
	"time"
)

// TokenBucket implements a simple token bucket rate limiter for WebSocket
type TokenBucket struct {
	capacity    int
	tokens      int
 refillRate  time.Duration
	lastRefill  time.Time
	mu          sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity int, refillRate time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed based on token bucket
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Refill tokens based on elapsed time
	if elapsed >= tb.refillRate {
		tb.tokens = tb.capacity
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// WebSocketRateLimiter manages rate limiters per client
type WebSocketRateLimiter struct {
	clients map[string]*TokenBucket
	mu      sync.RWMutex
}

func NewWebSocketRateLimiter() *WebSocketRateLimiter {
	return &WebSocketRateLimiter{
		clients: make(map[string]*TokenBucket),
	}
}

// GetOrCreate gets or creates a rate limiter for a client
func (wrl *WebSocketRateLimiter) GetOrCreate(username string) *TokenBucket {
	wrl.mu.RLock()
	bucket, exists := wrl.clients[username]
	wrl.mu.RUnlock()

	if exists {
		return bucket
	}

	wrl.mu.Lock()
	defer wrl.mu.Unlock()

	// Double-check after acquiring write lock
	if bucket, exists := wrl.clients[username]; exists {
		return bucket
	}

	// Create new bucket: 10 messages per second
	bucket = NewTokenBucket(10, time.Second)
	wrl.clients[username] = bucket
	return bucket
}

// Remove removes a rate limiter for a client
func (wrl *WebSocketRateLimiter) Remove(username string) {
	wrl.mu.Lock()
	defer wrl.mu.Unlock()
	delete(wrl.clients, username)
}
