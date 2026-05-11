package auth

import (
	"testing"
	"time"
)

func TestRateLimiterAllowsUntilCapHit(t *testing.T) {
	now := time.Unix(0, 0)
	rl := NewRateLimiter(3, time.Minute, func() time.Time { return now })
	for i := 0; i < 3; i++ {
		if !rl.Allow("k") {
			t.Fatalf("call %d should be allowed", i)
		}
	}
	if rl.Allow("k") {
		t.Fatal("4th call should be denied")
	}
	now = now.Add(2 * time.Minute)
	if !rl.Allow("k") {
		t.Fatal("after window elapsed, should be allowed")
	}
}
