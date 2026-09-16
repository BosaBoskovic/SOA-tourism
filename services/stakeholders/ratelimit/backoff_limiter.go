// Package ratelimit provides a small in-memory progressive-lockout tracker
// for security-sensitive, credential-guessing-shaped endpoints (login,
// change-password, ...). Single-instance only (no shared store across
// replicas) - good enough to blunt credential-stuffing/brute-force against
// a specific account at this deployment's scale. If stakeholders is scaled
// to N replicas, each replica tracks its own counters, so the effective
// grace/lockout is looser than configured (roughly divided across
// replicas) - acceptable for this project, but worth knowing before
// relying on it as the only defense.
package ratelimit

import (
	"sync"
	"time"
)

// BackoffLimiter locks an identity out for progressively longer periods
// after each consecutive failure, instead of a flat "N strikes, fixed
// timeout" rule: the first graceAttempts failures are free (typos happen),
// then every failure after that doubles the lockout - baseDelay, 2x, 4x,
// 8x, ... capped at maxLockout - so a sustained guessing attempt against
// one account gets exponentially more expensive for whoever's doing it,
// while a legitimate user who mistypes a password once or twice is never
// punished. A successful attempt resets the streak to zero.
type BackoffLimiter struct {
	mu            sync.Mutex
	graceAttempts int
	baseDelay     time.Duration
	maxLockout    time.Duration
	failures      map[string]int
	lockedUntil   map[string]time.Time
}

func NewBackoffLimiter(graceAttempts int, baseDelay, maxLockout time.Duration) *BackoffLimiter {
	return &BackoffLimiter{
		graceAttempts: graceAttempts,
		baseDelay:     baseDelay,
		maxLockout:    maxLockout,
		failures:      make(map[string]int),
		lockedUntil:   make(map[string]time.Time),
	}
}

// Allow reports whether an attempt for key is currently permitted. When it
// isn't, retryAfter says how much longer the caller has to wait - callers
// typically surface this as an HTTP "Retry-After" header.
func (l *BackoffLimiter) Allow(key string) (allowed bool, retryAfter time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	until, locked := l.lockedUntil[key]
	if !locked {
		return true, 0
	}
	remaining := time.Until(until)
	if remaining <= 0 {
		return true, 0
	}
	return false, remaining
}

// RecordFailure records a failure for key. Once graceAttempts have been
// used up, every further failure re-locks key for an exponentially larger
// duration than the previous one.
func (l *BackoffLimiter) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.failures[key]++
	n := l.failures[key]
	if n <= l.graceAttempts {
		return
	}

	// n=graceAttempts+1 -> shift 0 -> baseDelay
	// n=graceAttempts+2 -> shift 1 -> baseDelay*2, and so on.
	shift := n - l.graceAttempts - 1
	if shift > 30 { // guard against 1<<shift overflow on a very long streak
		shift = 30
	}
	lockout := l.baseDelay * time.Duration(int64(1)<<uint(shift))
	if lockout > l.maxLockout || lockout <= 0 { // <=0 covers any overflow that slipped through
		lockout = l.maxLockout
	}
	l.lockedUntil[key] = time.Now().Add(lockout)
}

// RecordSuccess clears any tracked failures for key.
func (l *BackoffLimiter) RecordSuccess(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
	delete(l.lockedUntil, key)
}
