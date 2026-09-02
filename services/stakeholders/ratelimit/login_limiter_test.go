package ratelimit

import (
	"testing"
	"time"
)

func TestLoginLimiter_LocksOutAfterMaxAttempts(t *testing.T) {
	l := NewLoginLimiter(3, time.Minute, time.Minute)

	if !l.Allow("ana") {
		t.Fatal("expected an unknown key to be allowed")
	}

	l.RecordFailure("ana")
	l.RecordFailure("ana")
	if !l.Allow("ana") {
		t.Fatal("expected the key to still be allowed below maxAttempts")
	}

	l.RecordFailure("ana") // 3rd failure hits maxAttempts
	if l.Allow("ana") {
		t.Fatal("expected the key to be locked out after maxAttempts failures")
	}
}

func TestLoginLimiter_LockoutIsPerKey(t *testing.T) {
	l := NewLoginLimiter(1, time.Minute, time.Minute)

	l.RecordFailure("ana")
	if l.Allow("ana") {
		t.Fatal("expected ana to be locked out")
	}
	if !l.Allow("marko") {
		t.Fatal("expected a different key to be unaffected by ana's lockout")
	}
}

func TestLoginLimiter_RecordSuccessClearsLockout(t *testing.T) {
	l := NewLoginLimiter(1, time.Minute, time.Minute)

	l.RecordFailure("ana")
	if l.Allow("ana") {
		t.Fatal("expected ana to be locked out")
	}

	l.RecordSuccess("ana")
	if !l.Allow("ana") {
		t.Fatal("expected RecordSuccess to clear the lockout")
	}
}

func TestLoginLimiter_LockoutExpiresAfterTheLockoutDuration(t *testing.T) {
	l := NewLoginLimiter(1, time.Hour, 20*time.Millisecond)

	l.RecordFailure("ana")
	if l.Allow("ana") {
		t.Fatal("expected ana to be locked out immediately after the failure")
	}

	time.Sleep(40 * time.Millisecond)
	if !l.Allow("ana") {
		t.Fatal("expected the lockout to have expired")
	}
}

func TestLoginLimiter_AttemptsOutsideTheWindowDontCount(t *testing.T) {
	l := NewLoginLimiter(2, 20*time.Millisecond, time.Minute)

	l.RecordFailure("ana")
	time.Sleep(30 * time.Millisecond) // the first failure ages out of the window
	l.RecordFailure("ana")

	if !l.Allow("ana") {
		t.Fatal("expected only one recent failure within the window - not enough to lock out")
	}
}
