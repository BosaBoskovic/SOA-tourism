package ratelimit

import (
	"testing"
	"time"
)

func TestBackoffLimiter_GraceAttemptsAreFree(t *testing.T) {
	l := NewBackoffLimiter(2, time.Hour, time.Hour) // long delays - a lockout here would fail the test loudly

	l.RecordFailure("ana")
	l.RecordFailure("ana")
	if allowed, _ := l.Allow("ana"); !allowed {
		t.Fatal("expected the first 2 (grace) failures to not lock the key out")
	}
}

func TestBackoffLimiter_LocksOutAfterGraceAttempts(t *testing.T) {
	l := NewBackoffLimiter(2, time.Hour, time.Hour)

	l.RecordFailure("ana")
	l.RecordFailure("ana")
	l.RecordFailure("ana") // 3rd failure - past grace
	allowed, retryAfter := l.Allow("ana")
	if allowed {
		t.Fatal("expected the key to be locked out after grace attempts are used up")
	}
	if retryAfter <= 0 {
		t.Fatalf("expected a positive retryAfter, got %v", retryAfter)
	}
}

func TestBackoffLimiter_EachFailureDoublesTheLockout(t *testing.T) {
	l := NewBackoffLimiter(0, time.Second, time.Hour)

	l.RecordFailure("ana") // 1st failure -> 1x base
	_, first := l.Allow("ana")

	l.RecordFailure("ana") // 2nd failure -> 2x base (recorded while still locked - matches an attacker hammering the endpoint)
	_, second := l.Allow("ana")

	l.RecordFailure("ana") // 3rd failure -> 4x base
	_, third := l.Allow("ana")

	if !(first < second && second < third) {
		t.Fatalf("expected strictly increasing lockouts, got %v -> %v -> %v", first, second, third)
	}
	// Roughly doubling each time (allow slack for the sub-millisecond it
	// takes the test itself to run between RecordFailure and Allow).
	if second < first*3/2 {
		t.Fatalf("expected the 2nd lockout to be roughly double the 1st: %v vs %v", first, second)
	}
	if third < second*3/2 {
		t.Fatalf("expected the 3rd lockout to be roughly double the 2nd: %v vs %v", second, third)
	}
}

func TestBackoffLimiter_LockoutIsCappedAtMax(t *testing.T) {
	l := NewBackoffLimiter(0, time.Second, 5*time.Second)

	for i := 0; i < 10; i++ { // far more than enough failures to blow past the cap uncapped
		l.RecordFailure("ana")
	}

	_, retryAfter := l.Allow("ana")
	if retryAfter > 5*time.Second {
		t.Fatalf("expected the lockout to be capped at 5s, got %v", retryAfter)
	}
}

func TestBackoffLimiter_LockoutIsPerKey(t *testing.T) {
	l := NewBackoffLimiter(0, time.Hour, time.Hour)

	l.RecordFailure("ana")
	if allowed, _ := l.Allow("ana"); allowed {
		t.Fatal("expected ana to be locked out")
	}
	if allowed, _ := l.Allow("marko"); !allowed {
		t.Fatal("expected a different key to be unaffected by ana's lockout")
	}
}

func TestBackoffLimiter_RecordSuccessResetsTheStreak(t *testing.T) {
	l := NewBackoffLimiter(0, time.Hour, time.Hour)

	l.RecordFailure("ana")
	if allowed, _ := l.Allow("ana"); allowed {
		t.Fatal("expected ana to be locked out")
	}

	l.RecordSuccess("ana")
	if allowed, _ := l.Allow("ana"); !allowed {
		t.Fatal("expected RecordSuccess to clear the lockout")
	}

	// The streak should be back to zero, not just the lockout cleared - the
	// very next failure should get the 1st-failure (short) lockout again,
	// not pick up where the pre-reset streak left off.
	l.RecordFailure("ana")
	_, retryAfter := l.Allow("ana")
	if retryAfter > 2*time.Hour { // 1st-failure lockout == baseDelay (1h); a resumed streak would be much larger
		t.Fatalf("expected a fresh 1st-failure lockout after reset, got %v", retryAfter)
	}
}

func TestBackoffLimiter_LockoutExpiresAfterItsOwnDuration(t *testing.T) {
	l := NewBackoffLimiter(0, 20*time.Millisecond, time.Minute)

	l.RecordFailure("ana")
	if allowed, _ := l.Allow("ana"); allowed {
		t.Fatal("expected ana to be locked out immediately after the failure")
	}

	time.Sleep(40 * time.Millisecond)
	if allowed, _ := l.Allow("ana"); !allowed {
		t.Fatal("expected the lockout to have expired")
	}
}
