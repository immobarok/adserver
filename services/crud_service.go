package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"adserver/db"
	"adserver/models"
)

// ─── Campaigns ───────────────────────────────────────────────────────────────

func ListCampaigns(ctx context.Context) ([]models.Campaign, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT id, name, budget, daily_limit, is_active, created_at, updated_at
		 FROM campaigns ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var c models.Campaign
		if err := rows.Scan(&c.ID, &c.Name, &c.Budget, &c.DailyLimit,
			&c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, rows.Err()
}

func CreateCampaign(ctx context.Context, req models.CreateCampaignRequest) (*models.Campaign, error) {
	var c models.Campaign
	err := db.DB.QueryRowContext(ctx,
		`INSERT INTO campaigns (name, budget, daily_limit)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, budget, daily_limit, is_active, created_at, updated_at`,
		req.Name, req.Budget, req.DailyLimit,
	).Scan(&c.ID, &c.Name, &c.Budget, &c.DailyLimit, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func GetCampaign(ctx context.Context, id string) (*models.Campaign, error) {
	var c models.Campaign
	err := db.DB.QueryRowContext(ctx,
		`SELECT id, name, budget, daily_limit, is_active, created_at, updated_at
		 FROM campaigns WHERE id = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Budget, &c.DailyLimit, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("campaign not found")
	}
	return &c, err
}

func UpdateCampaign(ctx context.Context, id string, req models.UpdateCampaignRequest) (*models.Campaign, error) {
	var c models.Campaign
	err := db.DB.QueryRowContext(ctx,
		`UPDATE campaigns
		 SET name        = COALESCE($2, name),
		     budget      = COALESCE($3, budget),
		     daily_limit = COALESCE($4, daily_limit),
		     is_active   = COALESCE($5, is_active),
		     updated_at  = NOW()
		 WHERE id = $1
		 RETURNING id, name, budget, daily_limit, is_active, created_at, updated_at`,
		id, req.Name, req.Budget, req.DailyLimit, req.IsActive,
	).Scan(&c.ID, &c.Name, &c.Budget, &c.DailyLimit, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("campaign not found")
	}
	return &c, err
}

func DeleteCampaign(ctx context.Context, id string) error {
	res, err := db.DB.ExecContext(ctx, `DELETE FROM campaigns WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("campaign not found")
	}
	return nil
}

// ─── Creatives ───────────────────────────────────────────────────────────────

func ListCreatives(ctx context.Context, campaignID string) ([]models.Creative, error) {
	query := `SELECT id, campaign_id, title, image_url, click_url, bid_price, is_active, created_at, updated_at
	          FROM creatives`
	args := []interface{}{}
	if campaignID != "" {
		query += ` WHERE campaign_id = $1`
		args = append(args, campaignID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creatives []models.Creative
	for rows.Next() {
		var c models.Creative
		if err := rows.Scan(&c.ID, &c.CampaignID, &c.Title, &c.ImageURL, &c.ClickURL,
			&c.BidPrice, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		creatives = append(creatives, c)
	}
	return creatives, rows.Err()
}

func CreateCreative(ctx context.Context, req models.CreateCreativeRequest) (*models.Creative, error) {
	var c models.Creative
	err := db.DB.QueryRowContext(ctx,
		`INSERT INTO creatives (campaign_id, title, image_url, click_url, bid_price)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, campaign_id, title, image_url, click_url, bid_price, is_active, created_at, updated_at`,
		req.CampaignID, req.Title, req.ImageURL, req.ClickURL, req.BidPrice,
	).Scan(&c.ID, &c.CampaignID, &c.Title, &c.ImageURL, &c.ClickURL,
		&c.BidPrice, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	return &c, err
}

func DeleteCreative(ctx context.Context, id string) error {
	res, err := db.DB.ExecContext(ctx, `DELETE FROM creatives WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("creative not found")
	}
	// Invalidate cache
	if db.RedisClient != nil {
		_ = db.RedisClient.Del(context.Background(), adCacheKey).Err()
	}
	return nil
}

// CampaignDailySpend returns today's total spend for a campaign.
func CampaignDailySpend(ctx context.Context, campaignID string) (float64, error) {
	var spend float64
	err := db.DB.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(cost), 0) FROM ad_events
		 WHERE campaign_id = $1 AND created_at >= CURRENT_DATE`,
		campaignID,
	).Scan(&spend)
	return spend, err
}

// ListAdEvents returns raw ad events with optional filters.
func ListAdEvents(ctx context.Context, campaignID, eventType string, from, to time.Time) ([]models.AdEvent, error) {
	query := `SELECT id, creative_id, campaign_id, event_type,
	          COALESCE(ip_address::text, ''), COALESCE(user_agent, ''),
	          COALESCE(referrer, ''), cost, created_at
	          FROM ad_events WHERE created_at BETWEEN $1 AND $2`
	args := []interface{}{from, to}

	if campaignID != "" {
		query += fmt.Sprintf(" AND campaign_id = $%d", len(args)+1)
		args = append(args, campaignID)
	}
	if eventType != "" {
		query += fmt.Sprintf(" AND event_type = $%d", len(args)+1)
		args = append(args, eventType)
	}
	query += " ORDER BY created_at DESC LIMIT 500"

	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.AdEvent
	for rows.Next() {
		var e models.AdEvent
		if err := rows.Scan(&e.ID, &e.CreativeID, &e.CampaignID, &e.EventType,
			&e.IPAddress, &e.UserAgent, &e.Referrer, &e.Cost, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
