package dialer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"crm-dialer-integration/internal/models"
	"crm-dialer-integration/internal/repository"
	"crm-dialer-integration/pkg/config"
)

type Service struct {
	client *Client
	repo   *repository.Repository
	logger *zap.Logger
}

func NewService(cfg *config.Config, repo *repository.Repository, logger *zap.Logger) *Service {
	client := NewClient(cfg.DialerAPIURL, cfg.DialerAPIKey, logger)

	return &Service{
		client: client,
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) SendContact(ctx context.Context, schedulerID, campaignID, bucketID string, contact Contact) error {
	return s.client.SendContact(ctx, schedulerID, campaignID, bucketID, contact)
}

func (s *Service) SyncCampaigns(ctx context.Context) error {
	campaigns, err := s.client.GetCampaigns(ctx)
	if err != nil {
		return fmt.Errorf("failed to get campaigns: %w", err)
	}

	for _, campaign := range campaigns {
		model := &models.DialerCampaign{
			ID:   campaign.ID,
			Name: campaign.Name,
		}

		if err := s.repo.UpsertDialerCampaign(ctx, model); err != nil {
			s.logger.Error("Failed to upsert campaign",
				zap.String("campaign_id", campaign.ID),
				zap.Error(err))
		}
	}

	s.logger.Info("Synced campaigns", zap.Int("count", len(campaigns)))
	return nil
}

func (s *Service) SyncBuckets(ctx context.Context, campaignID string) error {
	buckets, err := s.client.GetBuckets(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to get buckets: %w", err)
	}

	for _, bucket := range buckets {
		model := &models.DialerBucket{
			ID:         bucket.ID,
			CampaignID: bucket.CampaignID,
			Name:       bucket.Name,
		}

		if err := s.repo.UpsertDialerBucket(ctx, model); err != nil {
			s.logger.Error("Failed to upsert bucket",
				zap.String("bucket_id", bucket.ID),
				zap.Error(err))
		}
	}

	s.logger.Info("Synced buckets",
		zap.String("campaign_id", campaignID),
		zap.Int("count", len(buckets)))
	return nil
}

func (s *Service) SyncSchedulers(ctx context.Context) error {
	schedulers, err := s.client.GetSchedulers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schedulers: %w", err)
	}

	for _, scheduler := range schedulers {
		model := &models.DialerScheduler{
			ID:   scheduler.ID,
			Name: scheduler.Name,
		}

		if err := s.repo.UpsertDialerScheduler(ctx, model); err != nil {
			s.logger.Error("Failed to upsert scheduler",
				zap.String("scheduler_id", scheduler.ID),
				zap.Error(err))
		}
	}

	s.logger.Info("Synced schedulers", zap.Int("count", len(schedulers)))
	return nil
}

// Add GetSchedulers method to dialer client
func (c *Client) GetSchedulers(ctx context.Context) ([]Scheduler, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/v1/schedulers", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get schedulers: %s", string(body))
	}

	var schedulers []Scheduler
	if err := json.NewDecoder(resp.Body).Decode(&schedulers); err != nil {
		return nil, err
	}

	return schedulers, nil
}
