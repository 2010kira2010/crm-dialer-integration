package models

import (
	"encoding/json"
	"time"
)

// AmoCRMField represents a field from AmoCRM
type AmoCRMField struct {
	ID         int64     `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	Type       string    `db:"type" json:"type"`
	EntityType string    `db:"entity_type" json:"entity_type"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// DialerScheduler represents a scheduler from dialer system
type DialerScheduler struct {
	ID        string    `db:"id" json:"id"` // UUID
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// DialerBucket represents a bucket from dialer system
type DialerBucket struct {
	ID         string    `db:"id" json:"id"` // UUID
	CampaignID string    `db:"campaign_id" json:"campaign_id"`
	Name       string    `db:"name" json:"name"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// DialerCampaign represents a campaign from dialer system
type DialerCampaign struct {
	ID        string    `db:"id" json:"id"` // UUID
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// IntegrationFlow represents a React Flow configuration
type IntegrationFlow struct {
	ID        string          `db:"id" json:"id"`
	Name      string          `db:"name" json:"name"`
	FlowData  json.RawMessage `db:"flow_data" json:"flow_data"`
	IsActive  bool            `db:"is_active" json:"is_active"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}

// WebhookLog represents a webhook log entry
type WebhookLog struct {
	ID          string          `db:"id" json:"id"`
	WebhookType string          `db:"webhook_type" json:"webhook_type"`
	RawData     json.RawMessage `db:"raw_data" json:"raw_data"`
	ProcessedAt time.Time       `db:"processed_at" json:"processed_at"`
	Status      string          `db:"status" json:"status"`
}
