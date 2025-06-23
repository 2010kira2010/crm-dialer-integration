package dialer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
	logger     *zap.Logger
}

type Contact struct {
	Phone      string                 `json:"phone"`
	Name       string                 `json:"name"`
	Email      string                 `json:"email,omitempty"`
	CustomData map[string]interface{} `json:"custom_data,omitempty"`
}

type Campaign struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Bucket struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	Name       string `json:"name"`
}

type Scheduler struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewClient(apiURL, apiKey string, logger *zap.Logger) *Client {
	return &Client{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.apiURL, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c *Client) SendContact(ctx context.Context, schedulerID, campaignID, bucketID string, contact Contact) error {
	endpoint := fmt.Sprintf("/api/v1/campaigns/%s/buckets/%s/contacts", campaignID, bucketID)

	payload := map[string]interface{}{
		"scheduler_id": schedulerID,
		"contact":      contact,
	}

	resp, err := c.makeRequest(ctx, "POST", endpoint, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send contact: %s", string(body))
	}

	return nil
}

func (c *Client) GetCampaigns(ctx context.Context) ([]Campaign, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/v1/campaigns", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get campaigns: %s", string(body))
	}

	var campaigns []Campaign
	if err := json.NewDecoder(resp.Body).Decode(&campaigns); err != nil {
		return nil, err
	}

	return campaigns, nil
}

func (c *Client) GetBuckets(ctx context.Context, campaignID string) ([]Bucket, error) {
	endpoint := fmt.Sprintf("/api/v1/campaigns/%s/buckets", campaignID)

	resp, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get buckets: %s", string(body))
	}

	var buckets []Bucket
	if err := json.NewDecoder(resp.Body).Decode(&buckets); err != nil {
		return nil, err
	}

	return buckets, nil
}
