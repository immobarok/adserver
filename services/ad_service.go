package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"adserver/config"
	"adserver/db"
	"adserver/models"
)

const adCacheKey = "ad:candidates"

// GetAdCandidate selects the best active creative for the given IP.
// It enforces frequency capping (max N impressions/IP/day) and daily budget limits.
// Results are cached in Redis when available.
func GetAdCandidate(ctx context.Context, ip string) (*models.AdResponse, error) {
	// Check frequency cap
	if capped, err := isFrequencyCapped(ctx, ip); err != nil {
		log.Printf("freq cap check error: %v", err)
	} else if capped {
		return nil, fmt.Errorf("frequency cap reached for IP %s", ip)
	}

	// Try Redis cache for candidate list
	candidates, err := getCachedCandidates(ctx)
	if err != nil || len(candidates) == 0 {
		candidates, err = fetchCandidatesFromDB(ctx)
		if err != nil {
			return nil, err
		}
		_ = setCachedCandidates(ctx, candidates)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no eligible ads available")
	}

	// Pick the highest bid_price candidate
	best := candidates[0]
	return &models.AdResponse{
		CreativeID:    best.ID,
		CampaignID:    best.CampaignID,
		Title:         best.Title,
		ImageURL:      best.ImageURL,
		ClickURL:      best.ClickURL,
		ImpressionURL: fmt.Sprintf("/track/impression?creative_id=%s", best.ID),
	}, nil
}

// fetchCandidatesFromDB queries active creatives respecting daily budget limits.
func fetchCandidatesFromDB(ctx context.Context) ([]models.Creative, error) {
	query := `
		SELECT
			c.id, c.campaign_id, c.title, c.image_url, c.click_url, c.bid_price,
			c.is_active, c.created_at, c.updated_at
		FROM creatives c
		JOIN campaigns cp ON cp.id = c.campaign_id
		WHERE c.is_active = TRUE
		  AND cp.is_active = TRUE
		  AND (
		      cp.daily_limit = 0
		      OR (
		          SELECT COALESCE(SUM(ae.cost), 0)
		          FROM ad_events ae
		          WHERE ae.campaign_id = cp.id
		            AND ae.created_at >= CURRENT_DATE
		      ) < cp.daily_limit
		  )
		ORDER BY c.bid_price DESC
		LIMIT 50
	`

	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creatives []models.Creative
	for rows.Next() {
		var c models.Creative
		if err := rows.Scan(
			&c.ID, &c.CampaignID, &c.Title, &c.ImageURL, &c.ClickURL,
			&c.BidPrice, &c.IsActive, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		creatives = append(creatives, c)
	}
	return creatives, rows.Err()
}

// RecordImpression persists an impression event to the database.
func RecordImpression(ctx context.Context, creativeID, ip, userAgent, referrer string) error {
	creative, err := GetCreativeByID(ctx, creativeID)
	if err != nil {
		return fmt.Errorf("creative not found: %w", err)
	}

	query := `
		INSERT INTO ad_events (creative_id, campaign_id, event_type, ip_address, user_agent, referrer, cost)
		VALUES ($1, $2, 'impression', $3::inet, $4, $5, $6)
	`
	_, err = db.DB.ExecContext(ctx, query,
		creative.ID, creative.CampaignID, ip, userAgent, referrer, creative.BidPrice,
	)
	if err != nil {
		return err
	}

	// Increment frequency cap counter in Redis
	_ = incrementFrequency(ctx, ip)
	return nil
}

// RecordClick persists a click event to the database.
func RecordClick(ctx context.Context, creativeID, ip, userAgent, referrer string) error {
	creative, err := GetCreativeByID(ctx, creativeID)
	if err != nil {
		return fmt.Errorf("creative not found: %w", err)
	}

	query := `
		INSERT INTO ad_events (creative_id, campaign_id, event_type, ip_address, user_agent, referrer, cost)
		VALUES ($1, $2, 'click', $3::inet, $4, $5, 0)
	`
	_, err = db.DB.ExecContext(ctx, query,
		creative.ID, creative.CampaignID, ip, userAgent, referrer,
	)
	return err
}

// GetCreativeByID fetches a single creative.
func GetCreativeByID(ctx context.Context, id string) (*models.Creative, error) {
	query := `SELECT id, campaign_id, title, image_url, click_url, bid_price, is_active, created_at, updated_at
	          FROM creatives WHERE id = $1`
	row := db.DB.QueryRowContext(ctx, query, id)
	var c models.Creative
	err := row.Scan(&c.ID, &c.CampaignID, &c.Title, &c.ImageURL, &c.ClickURL,
		&c.BidPrice, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("creative %s not found", id)
	}
	return &c, err
}

// GetReport returns aggregated impression/click stats per creative.
func GetReport(ctx context.Context, from, to time.Time, campaignID string) ([]models.ReportRow, error) {
	args := []interface{}{from, to}
	campaignFilter := ""
	if campaignID != "" {
		campaignFilter = "AND ae.campaign_id = $3"
		args = append(args, campaignID)
	}

	query := fmt.Sprintf(`
		SELECT
			cp.id                                                   AS campaign_id,
			cp.name                                                 AS campaign_name,
			cr.id                                                   AS creative_id,
			cr.title                                                AS creative_title,
			COUNT(*) FILTER (WHERE ae.event_type = 'impression')   AS impressions,
			COUNT(*) FILTER (WHERE ae.event_type = 'click')        AS clicks,
			ROUND(
				CASE
					WHEN COUNT(*) FILTER (WHERE ae.event_type = 'impression') = 0 THEN 0
					ELSE COUNT(*) FILTER (WHERE ae.event_type = 'click')::NUMERIC
					     / COUNT(*) FILTER (WHERE ae.event_type = 'impression') * 100
				END, 2
			)                                                       AS ctr,
			COALESCE(SUM(ae.cost), 0)                              AS total_cost
		FROM ad_events ae
		JOIN creatives cr  ON cr.id  = ae.creative_id
		JOIN campaigns cp  ON cp.id  = ae.campaign_id
		WHERE ae.created_at BETWEEN $1 AND $2
		%s
		GROUP BY cp.id, cp.name, cr.id, cr.title
		ORDER BY impressions DESC
	`, campaignFilter)

	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var report []models.ReportRow
	for rows.Next() {
		var r models.ReportRow
		if err := rows.Scan(&r.CampaignID, &r.CampaignName, &r.CreativeID,
			&r.CreativeTitle, &r.Impressions, &r.Clicks, &r.CTR, &r.TotalCost); err != nil {
			return nil, err
		}
		report = append(report, r)
	}
	return report, rows.Err()
}

// ─── Redis helpers ───────────────────────────────────────────────────────────

func freqKey(ip string) string {
	return fmt.Sprintf("freq:%s:%s", ip, time.Now().Format("2006-01-02"))
}

func isFrequencyCapped(ctx context.Context, ip string) (bool, error) {
	if db.RedisClient == nil {
		return false, nil
	}
	val, err := db.RedisClient.Get(ctx, freqKey(ip)).Int()
	if err != nil {
		return false, nil // key doesn't exist yet
	}
	return val >= config.App.FrequencyCap, nil
}

func incrementFrequency(ctx context.Context, ip string) error {
	if db.RedisClient == nil {
		return nil
	}
	key := freqKey(ip)
	pipe := db.RedisClient.Pipeline()
	pipe.Incr(ctx, key)
	pipe.ExpireAt(ctx, key, midnight())
	_, err := pipe.Exec(ctx)
	return err
}

func midnight() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}

func getCachedCandidates(ctx context.Context) ([]models.Creative, error) {
	if db.RedisClient == nil {
		return nil, fmt.Errorf("redis unavailable")
	}
	data, err := db.RedisClient.Get(ctx, adCacheKey).Bytes()
	if err != nil {
		return nil, err
	}
	var candidates []models.Creative
	return candidates, json.Unmarshal(data, &candidates)
}

func setCachedCandidates(ctx context.Context, candidates []models.Creative) error {
	if db.RedisClient == nil {
		return nil
	}
	data, err := json.Marshal(candidates)
	if err != nil {
		return err
	}
	ttl := time.Duration(config.App.CacheTTLSeconds) * time.Second
	return db.RedisClient.Set(ctx, adCacheKey, data, ttl).Err()
}
