package models

import (
	"time"
)

// PublicIpUpdate represents a public IP update record
type PublicIpUpdate struct {
	ID                  int       `json:"id"`
	NewPublicIpAddress  string    `json:"new_public_ip_address"`
	OldPublicIpAddress  string    `json:"old_public_ip_address"`
	ChangedAt           time.Time `json:"changed_at"`
}