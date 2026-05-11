package auth

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu     sync.Mutex
	cap    int
	window time.Duration
	now    func() time.Time
	hits   map[string][]time.Time
}

func NewRateLimiter(cap int, window time.Duration, now func() time.Time) *RateLimiter {
	return &RateLimiter{cap: cap, window: window, now: now, hits: map[string][]time.Time{}}
}

func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := r.now().Add(-r.window)
	kept := r.hits[key][:0]
	for _, t := range r.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= r.cap {
		r.hits[key] = kept
		return false
	}
	kept = append(kept, r.now())
	r.hits[key] = kept
	return true
}
