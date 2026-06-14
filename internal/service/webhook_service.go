package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type WebhookPayload struct {
	Event       string    `json:"event"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	AccessedAt  time.Time `json:"accessed_at"`
	IP          string    `json:"ip"`
}

type WebhookService interface {
	FireWebhook(webhookURL string, payload WebhookPayload)
}

type webhookService struct {
	client *http.Client
}

func NewWebhookService() WebhookService {
	return &webhookService{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *webhookService) FireWebhook(webhookURL string, payload WebhookPayload) {
	go func() {
		data, err := json.Marshal(payload)
		if err != nil {
			return
		}
		req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(data))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()
	}()
}
