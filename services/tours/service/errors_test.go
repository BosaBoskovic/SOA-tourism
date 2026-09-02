package service

import "testing"

func TestIsOwner(t *testing.T) {
	tests := []struct {
		name           string
		authorID       string
		callerUsername string
		callerRole     string
		want           bool
	}{
		{"owner matches", "ana", "ana", "guide", true},
		{"admin always allowed", "ana", "marko", "admin", true},
		{"different guide is not the owner", "ana", "marko", "guide", false},
		{"empty caller is never the owner", "ana", "", "guide", false},
		{"empty caller is not saved by an empty authorId either", "", "", "guide", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isOwner(tt.authorID, tt.callerUsername, tt.callerRole)
			if got != tt.want {
				t.Errorf("isOwner(%q, %q, %q) = %v, want %v", tt.authorID, tt.callerUsername, tt.callerRole, got, tt.want)
			}
		})
	}
}
