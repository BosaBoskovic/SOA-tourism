package repository

import (
	"sync"
	"testing"
)

func TestPurchaseRepository_SaveThenHasTokenHitsTheCache(t *testing.T) {
	repo := NewPurchaseRepository()
	repo.SaveToken("tourist1", "tour1")

	// A saved token is served from the in-memory cache and must never fall
	// through to the payments HTTP fallback - if it did, this call would
	// depend on a real payments service being reachable.
	got, err := repo.HasToken("tourist1", "tour1")
	if err != nil {
		t.Fatalf("HasToken returned an unexpected error: %v", err)
	}
	if !got {
		t.Fatal("expected HasToken to return true for a token that was just saved")
	}
}

// Regression test for guarding PurchaseRepository.tokens with a sync.RWMutex:
// concurrent SaveToken/HasToken on a plain Go map is a fatal "concurrent map
// read and map write" crash, not just a benign data race. Run with
// `go test -race` to have the race detector actually catch a regression.
func TestPurchaseRepository_ConcurrentAccessDoesNotCrash(t *testing.T) {
	repo := NewPurchaseRepository()
	repo.SaveToken("seed", "seed") // ensure HasToken below always hits the cache, never the network

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			repo.SaveToken("concurrent", "writer")
		}()
		go func() {
			defer wg.Done()
			_, _ = repo.HasToken("seed", "seed")
		}()
	}
	wg.Wait()
}
