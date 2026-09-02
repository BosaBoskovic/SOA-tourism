// Package ratelimit provides a small in-memory lockout tracker for login
// attempts. Single-instance only (no shared store across replicas) - good
// enough to blunt credential-stuffing/brute-force against a specific
// account for this deployment's scale.
package ratelimit

import (
	"sync"
	"time"
)

type LoginLimiter struct {
	mu          sync.Mutex
	maxAttempts int
	window      time.Duration
	lockout     time.Duration
	attempts    map[string][]time.Time
	lockedUntil map[string]time.Time
}

func NewLoginLimiter(maxAttempts int, window, lockout time.Duration) *LoginLimiter {
	return &LoginLimiter{
		maxAttempts: maxAttempts,
		window:      window,
		lockout:     lockout,
		attempts:    make(map[string][]time.Time),
		lockedUntil: make(map[string]time.Time),
	}
}

// Allow reports whether a login attempt for key is currently permitted.
func (l *LoginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	until, locked := l.lockedUntil[key]
	if !locked {
		return true
	}
	if time.Now().Before(until) {
		return false
	}
	delete(l.lockedUntil, key)
	delete(l.attempts, key)
	return true
}

// RecordFailure records a failed attempt for key, locking it out once
// maxAttempts have happened within window.
func (l *LoginLimiter) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	old := l.attempts[key]
	recent := make([]time.Time, 0, len(old)+1)
	for _, t := range old {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	recent = append(recent, now)
	l.attempts[key] = recent

	if len(recent) >= l.maxAttempts {
		l.lockedUntil[key] = now.Add(l.lockout)
	}
}

// RecordSuccess clears any tracked failures for key.
func (l *LoginLimiter) RecordSuccess(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
	delete(l.lockedUntil, key)
}
