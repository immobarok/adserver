package models

import "time"

// Campaign represents an advertising campaign.
type Campaign struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Budget     float64   `json:"budget"`
	DailyLimit float64   `json:"daily_limit"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateCampaignRequest is the payload for creating a campaign.
type CreateCampaignRequest struct {
	Name       string  `json:"name"        binding:"required"`
	Budget     float64 `json:"budget"      binding:"required,gt=0"`
	DailyLimit float64 `json:"daily_limit" binding:"required,gt=0"`
}

// UpdateCampaignRequest is the payload for updating a campaign.
type UpdateCampaignRequest struct {
	Name       *string  `json:"name"`
	Budget     *float64 `json:"budget"`
	DailyLimit *float64 `json:"daily_limit"`
	IsActive   *bool    `json:"is_active"`
}

// Creative represents a single ad creative belonging to a campaign.
type Creative struct {
	ID         string    `json:"id"`
	CampaignID string    `json:"campaign_id"`
	Title      string    `json:"title"`
	ImageURL   string    `json:"image_url"`
	ClickURL   string    `json:"click_url"`
	BidPrice   float64   `json:"bid_price"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateCreativeRequest is the payload for creating a creative.
type CreateCreativeRequest struct {
	CampaignID string  `json:"campaign_id" binding:"required,uuid"`
	Title      string  `json:"title"       binding:"required"`
	ImageURL   string  `json:"image_url"   binding:"required,url"`
	ClickURL   string  `json:"click_url"   binding:"required,url"`
	BidPrice   float64 `json:"bid_price"   binding:"required,gt=0"`
}

// AdEvent represents an impression or click event.
type AdEvent struct {
	ID         string    `json:"id"`
	CreativeID string    `json:"creative_id"`
	CampaignID string    `json:"campaign_id"`
	EventType  string    `json:"event_type"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Referrer   string    `json:"referrer"`
	Cost       float64   `json:"cost"`
	CreatedAt  time.Time `json:"created_at"`
}

// AdResponse is the data returned when serving an ad.
type AdResponse struct {
	CreativeID  string  `json:"creative_id"`
	CampaignID  string  `json:"campaign_id"`
	Title       string  `json:"title"`
	ImageURL    string  `json:"image_url"`
	ClickURL    string  `json:"click_url"`
	ImpressionURL string `json:"impression_url"`
}

// ReportRow holds aggregated stats per creative.
type ReportRow struct {
	CampaignID   string  `json:"campaign_id"`
	CampaignName string  `json:"campaign_name"`
	CreativeID   string  `json:"creative_id"`
	CreativeTitle string `json:"creative_title"`
	Impressions  int64   `json:"impressions"`
	Clicks       int64   `json:"clicks"`
	CTR          float64 `json:"ctr"`
	TotalCost    float64 `json:"total_cost"`
}
