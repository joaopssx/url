package model

import "time"

type URL struct {
	ID          string     `json:"id"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	UserID      *string    `json:"user_id"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	AccessCount int        `json:"access_count"`
	DeletedAt   *time.Time `json:"deleted_at"`
	WebhookURL  *string    `json:"webhook_url,omitempty"`
	ShortURL    string     `json:"short_url,omitempty" db:"-"`
}
