// Package notify makes a best-effort call to stakeholders' internal
// notification endpoint. A failure here (stakeholders down/slow) must never
// break whatever triggered the notification.
package notify

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

var client = &http.Client{Timeout: 3 * time.Second}

func stakeholdersURL() string {
	url := os.Getenv("STAKEHOLDERS_URL")
	if url == "" {
		url = "http://localhost:8081"
	}
	return url
}

type payload struct {
	Username        string `json:"username"`
	Type            string `json:"type"`
	Message         string `json:"message"`
	RelatedUsername string `json:"relatedUsername,omitempty"`
}

func send(p payload) {
	body, err := json.Marshal(p)
	if err != nil {
		return
	}
	resp, err := client.Post(stakeholdersURL()+"/stakeholders/notifications/internal", "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Warn("notification delivery failed", "username", p.Username, "error", err)
		return
	}
	defer resp.Body.Close()
}

// NewReview notifies tourAuthorUsername that touristUsername reviewed
// their tour. Fire-and-forget - call it in a goroutine.
func NewReview(tourAuthorUsername, touristUsername string) {
	if tourAuthorUsername == "" || tourAuthorUsername == touristUsername {
		return
	}
	go send(payload{
		Username:        tourAuthorUsername,
		Type:            "review",
		Message:         touristUsername + " je ostavio/la recenziju na tvoju turu.",
		RelatedUsername: touristUsername,
	})
}
