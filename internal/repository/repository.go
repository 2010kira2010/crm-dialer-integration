package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"crm-dialer-integration/internal/models"
)

type Repository struct {
	db     *sql.DB
	logger *zap.Logger
}

func New(databaseURL string, logger *zap.Logger) (*Repository, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Repository{
		db:     db,
		logger: logger,
	}, nil
}

// AmoCRM Fields
func (r *Repository) GetAmoCRMFields(ctx context.Context, entityType string) ([]*models.AmoCRMField, error) {
	query := `
        SELECT id, name, type, created_at, updated_at
        FROM amocrm_fields
        WHERE entity_type = $1
        ORDER BY name
    `

	rows, err := r.db.QueryContext(ctx, query, entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to query fields: %w", err)
	}
	defer rows.Close()

	var fields []*models.AmoCRMField
	for rows.Next() {
		var field models.AmoCRMField
		if err := rows.Scan(&field.ID, &field.Name, &field.Type, &field.CreatedAt, &field.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan field: %w", err)
		}
		fields = append(fields, &field)
	}

	return fields, nil
}

func (r *Repository) SaveAmoCRMField(ctx context.Context, field *models.AmoCRMField) error {
	query := `
        INSERT INTO amocrm_fields (id, name, type, entity_type, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.ExecContext(ctx, query,
		field.ID, field.Name, field.Type, field.EntityType,
		time.Now(), time.Now())

	if err != nil {
		return fmt.Errorf("failed to save field: %w", err)
	}

	return nil
}

func (r *Repository) UpsertAmoCRMField(ctx context.Context, field *models.AmoCRMField) error {
	query := `
        INSERT INTO amocrm_fields (id, name, type, entity_type, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            type = EXCLUDED.type,
            updated_at = EXCLUDED.updated_at
    `

	_, err := r.db.ExecContext(ctx, query,
		field.ID, field.Name, field.Type, field.EntityType,
		time.Now(), time.Now())

	if err != nil {
		return fmt.Errorf("failed to upsert field: %w", err)
	}

	return nil
}

// Dialer entities
func (r *Repository) GetDialerSchedulers(ctx context.Context) ([]*models.DialerScheduler, error) {
	query := `
        SELECT id, name, created_at, updated_at
        FROM dialer_schedulers
        ORDER BY name
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query schedulers: %w", err)
	}
	defer rows.Close()

	var schedulers []*models.DialerScheduler
	for rows.Next() {
		var scheduler models.DialerScheduler
		if err := rows.Scan(&scheduler.ID, &scheduler.Name, &scheduler.CreatedAt, &scheduler.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan scheduler: %w", err)
		}
		schedulers = append(schedulers, &scheduler)
	}

	return schedulers, nil
}

func (r *Repository) GetDialerCampaigns(ctx context.Context) ([]*models.DialerCampaign, error) {
	query := `
        SELECT id, name, created_at, updated_at
        FROM dialer_campaigns
        ORDER BY name
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*models.DialerCampaign
	for rows.Next() {
		var campaign models.DialerCampaign
		if err := rows.Scan(&campaign.ID, &campaign.Name, &campaign.CreatedAt, &campaign.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		campaigns = append(campaigns, &campaign)
	}

	return campaigns, nil
}

func (r *Repository) GetDialerBuckets(ctx context.Context, campaignID string) ([]*models.DialerBucket, error) {
	query := `
        SELECT id, campaign_id, name, created_at, updated_at
        FROM dialer_buckets
        WHERE ($1 = '' OR campaign_id = $1)
        ORDER BY name
    `

	rows, err := r.db.QueryContext(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to query buckets: %w", err)
	}
	defer rows.Close()

	var buckets []*models.DialerBucket
	for rows.Next() {
		var bucket models.DialerBucket
		if err := rows.Scan(&bucket.ID, &bucket.CampaignID, &bucket.Name, &bucket.CreatedAt, &bucket.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan bucket: %w", err)
		}
		buckets = append(buckets, &bucket)
	}

	return buckets, nil
}

func (r *Repository) UpsertDialerScheduler(ctx context.Context, scheduler *models.DialerScheduler) error {
	query := `
        INSERT INTO dialer_schedulers (id, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            updated_at = EXCLUDED.updated_at
    `

	_, err := r.db.ExecContext(ctx, query,
		scheduler.ID, scheduler.Name, time.Now(), time.Now())

	if err != nil {
		return fmt.Errorf("failed to upsert scheduler: %w", err)
	}

	return nil
}

func (r *Repository) UpsertDialerCampaign(ctx context.Context, campaign *models.DialerCampaign) error {
	query := `
        INSERT INTO dialer_campaigns (id, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            updated_at = EXCLUDED.updated_at
    `

	_, err := r.db.ExecContext(ctx, query,
		campaign.ID, campaign.Name, time.Now(), time.Now())

	if err != nil {
		return fmt.Errorf("failed to upsert campaign: %w", err)
	}

	return nil
}

func (r *Repository) UpsertDialerBucket(ctx context.Context, bucket *models.DialerBucket) error {
	query := `
        INSERT INTO dialer_buckets (id, campaign_id, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (id) DO UPDATE SET
            campaign_id = EXCLUDED.campaign_id,
            name = EXCLUDED.name,
            updated_at = EXCLUDED.updated_at
    `

	_, err := r.db.ExecContext(ctx, query,
		bucket.ID, bucket.CampaignID, bucket.Name, time.Now(), time.Now())

	if err != nil {
		return fmt.Errorf("failed to upsert bucket: %w", err)
	}

	return nil
}

// Integration Flows
func (r *Repository) GetIntegrationFlows(ctx context.Context) ([]*models.IntegrationFlow, error) {
	query := `
        SELECT id, name, flow_data, is_active, created_at, updated_at
        FROM integration_flows
        ORDER BY name
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query flows: %w", err)
	}
	defer rows.Close()

	var flows []*models.IntegrationFlow
	for rows.Next() {
		var flow models.IntegrationFlow
		if err := rows.Scan(&flow.ID, &flow.Name, &flow.FlowData, &flow.IsActive, &flow.CreatedAt, &flow.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan flow: %w", err)
		}
		flows = append(flows, &flow)
	}

	return flows, nil
}

func (r *Repository) GetIntegrationFlowByID(ctx context.Context, id string) (*models.IntegrationFlow, error) {
	query := `
        SELECT id, name, flow_data, is_active, created_at, updated_at
        FROM integration_flows
        WHERE id = $1
    `

	var flow models.IntegrationFlow
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&flow.ID, &flow.Name, &flow.FlowData, &flow.IsActive, &flow.CreatedAt, &flow.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}

	return &flow, nil
}

func (r *Repository) CreateIntegrationFlow(ctx context.Context, flow *models.IntegrationFlow) error {
	query := `
        INSERT INTO integration_flows (id, name, flow_data, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	_, err := r.db.ExecContext(ctx, query,
		flow.ID, flow.Name, flow.FlowData, flow.IsActive, time.Now(), time.Now())

	if err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}

	return nil
}

func (r *Repository) UpdateIntegrationFlow(ctx context.Context, flow *models.IntegrationFlow) error {
	query := `
        UPDATE integration_flows
        SET name = $2, flow_data = $3, is_active = $4, updated_at = $5
        WHERE id = $1
    `

	_, err := r.db.ExecContext(ctx, query,
		flow.ID, flow.Name, flow.FlowData, flow.IsActive, time.Now())

	if err != nil {
		return fmt.Errorf("failed to update flow: %w", err)
	}

	return nil
}

func (r *Repository) DeleteIntegrationFlow(ctx context.Context, id string) error {
	query := `DELETE FROM integration_flows WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}

	return nil
}

// Webhook Logs
func (r *Repository) SaveWebhookLog(ctx context.Context, log *models.WebhookLog) error {
	query := `
        INSERT INTO webhook_logs (id, webhook_type, raw_data, processed_at, status)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.WebhookType, log.RawData, log.ProcessedAt, log.Status)

	if err != nil {
		return fmt.Errorf("failed to save webhook log: %w", err)
	}

	return nil
}
