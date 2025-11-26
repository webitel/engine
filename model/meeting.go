package model

import "time"

type Meeting struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	TTL       time.Duration     `json:"ttl"`
	CreatedAt int64             `json:"created_at"`
	ExpiresAt int64             `json:"expires_at"`
	Metadata  map[string]string `json:"metadata"`
	DomainId  int64             `json:"domain_id"`
}
