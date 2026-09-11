package messaging

import "testing"

func TestDecideOutcome_RateZero_NeverFails(t *testing.T) {
	for _, draw := range []float64{0, 0.001, 0.5, 0.999} {
		if err := decideOutcome(0, func() float64 { return draw }); err != nil {
			t.Fatalf("rate=0 should never fail (draw=%v), got %v", draw, err)
		}
	}
}

func TestDecideOutcome_RateOne_AlwaysFails(t *testing.T) {
	for _, draw := range []float64{0, 0.001, 0.5, 0.999} {
		if err := decideOutcome(1, func() float64 { return draw }); err == nil {
			t.Fatalf("rate=1 should always fail (draw=%v), got nil error", draw)
		}
	}
}

func TestDecideOutcome_DrawBelowRate_Fails(t *testing.T) {
	if err := decideOutcome(0.5, func() float64 { return 0.4 }); err == nil {
		t.Fatal("expected a draw below the failure rate to fail")
	}
}

func TestDecideOutcome_DrawAtOrAboveRate_Succeeds(t *testing.T) {
	if err := decideOutcome(0.5, func() float64 { return 0.6 }); err != nil {
		t.Fatalf("expected a draw above the failure rate to succeed, got %v", err)
	}
}
