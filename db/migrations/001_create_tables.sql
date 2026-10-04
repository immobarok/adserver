-- Migration: 001_create_tables.sql
-- Run this against your PostgreSQL database

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- CAMPAIGNS
-- ============================================================
CREATE TABLE IF NOT EXISTS campaigns (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    budget      NUMERIC(12,2) NOT NULL DEFAULT 0,
    daily_limit NUMERIC(12,2) NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- CREATIVES (ads belonging to a campaign)
-- ============================================================
CREATE TABLE IF NOT EXISTS creatives (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    image_url   TEXT NOT NULL,
    click_url   TEXT NOT NULL,
    bid_price   NUMERIC(10,4) NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_creatives_campaign_id ON creatives(campaign_id);
CREATE INDEX IF NOT EXISTS idx_creatives_is_active ON creatives(is_active);

-- ============================================================
-- AD EVENTS (impressions & clicks)
-- ============================================================
CREATE TABLE IF NOT EXISTS ad_events (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    creative_id  UUID NOT NULL REFERENCES creatives(id) ON DELETE CASCADE,
    campaign_id  UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    event_type   VARCHAR(20) NOT NULL CHECK (event_type IN ('impression', 'click')),
    ip_address   INET,
    user_agent   TEXT,
    referrer     TEXT,
    cost         NUMERIC(10,4) NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ad_events_creative_id  ON ad_events(creative_id);
CREATE INDEX IF NOT EXISTS idx_ad_events_campaign_id  ON ad_events(campaign_id);
CREATE INDEX IF NOT EXISTS idx_ad_events_event_type   ON ad_events(event_type);
CREATE INDEX IF NOT EXISTS idx_ad_events_created_at   ON ad_events(created_at);
CREATE INDEX IF NOT EXISTS idx_ad_events_ip_address   ON ad_events(ip_address);

-- ============================================================
-- SEED DATA
-- ============================================================
INSERT INTO campaigns (id, name, budget, daily_limit, is_active)
VALUES
    ('a1b2c3d4-0000-0000-0000-000000000001', 'Summer Sale 2026',      5000.00, 500.00, TRUE),
    ('a1b2c3d4-0000-0000-0000-000000000002', 'Back to School',        3000.00, 300.00, TRUE),
    ('a1b2c3d4-0000-0000-0000-000000000003', 'Brand Awareness Q4',    8000.00, 800.00, TRUE)
ON CONFLICT DO NOTHING;

INSERT INTO creatives (campaign_id, title, image_url, click_url, bid_price, is_active)
VALUES
    ('a1b2c3d4-0000-0000-0000-000000000001', 'Summer Banner 1',  'https://picsum.photos/seed/ad1/728/90',  'https://example.com/summer?ref=ad1', 0.05, TRUE),
    ('a1b2c3d4-0000-0000-0000-000000000001', 'Summer Banner 2',  'https://picsum.photos/seed/ad2/300/250', 'https://example.com/summer?ref=ad2', 0.04, TRUE),
    ('a1b2c3d4-0000-0000-0000-000000000002', 'School Ad 1',      'https://picsum.photos/seed/ad3/728/90',  'https://example.com/school?ref=ad3', 0.06, TRUE),
    ('a1b2c3d4-0000-0000-0000-000000000002', 'School Ad 2',      'https://picsum.photos/seed/ad4/300/250', 'https://example.com/school?ref=ad4', 0.03, TRUE),
    ('a1b2c3d4-0000-0000-0000-000000000003', 'Brand Ad 1',       'https://picsum.photos/seed/ad5/728/90',  'https://example.com/brand?ref=ad5',  0.08, TRUE)
ON CONFLICT DO NOTHING;
