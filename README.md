# AdServer

A high-performance programmatic ad server built with **Go + Gin**, **PostgreSQL (Neon)**, and **Redis**.

---

## Architecture Overview & Request Flow

```
                        ┌──────────────────────────────────────┐
                        │         Publisher Page (HTML)         │
                        │  static/publisher.html               │
                        └──────────────┬───────────────────────┘
                                       │ 1. GET /ad  (async fetch)
                                       ▼
┌──────────────────────────────────────────────────────────────────┐
│                        Go + Gin HTTP Server                       │
│                                                                    │
│  ┌──────────────┐   ┌──────────────────┐   ┌──────────────────┐  │
│  │  Middleware  │──▶│    Handlers      │──▶│    Services      │  │
│  │  Logger      │   │  /ad             │   │  GetAdCandidate  │  │
│  │  CORS        │   │  /track/         │   │  RecordImpression│  │
│  └──────────────┘   │  impression      │   │  RecordClick     │  │
│                      │  /track/click    │   │  GetReport       │  │
│                      │  /api/v1/...     │   └────────┬─────────┘  │
│                      └──────────────────┘            │            │
└─────────────────────────────────────────────────────┼────────────┘
                                                       │
                    ┌──────────────────────────────────┤
                    │                                  │
                    ▼                                  ▼
         ┌──────────────────┐              ┌──────────────────┐
         │   PostgreSQL      │              │   Redis Cache    │
         │   (Neon Cloud)    │              │   (optional)     │
         │                  │              │                  │
         │  campaigns        │              │  ad:candidates   │
         │  creatives        │              │  freq:{ip}:{date}│
         │  ad_events        │              │                  │
         └──────────────────┘              └──────────────────┘
```

### Request Flow — Serving an Ad

```
Client (Browser)
  │
  │── GET /ad ──────────────────────────────────────────────────────────▶ Gin Router
  │                                                                            │
  │                                                              Check Redis frequency cap
  │                                                              key: freq:{ip}:{date}
  │                                                                            │
  │                                                         ┌─── cap reached? ─┤
  │                                                         │                  │ No
  │                                                    204 No Content    Check Redis cache
  │                                                                      key: ad:candidates
  │                                                                            │
  │                                                              ┌── cache hit? ─┤
  │                                                              │               │ Miss
  │                                                        Use cached        Query PostgreSQL
  │                                                        candidates        (active creatives,
  │                                                                           daily budget check)
  │                                                                                │
  │                                                                         Cache result (60s TTL)
  │                                                                                │
  │◀──── 200 JSON { creative_id, image_url, click_url, impression_url } ──────────┘
  │
  │── <img src="/track/impression?creative_id=xxx"> ──────────────────▶ Fire-and-forget
  │                                                                       INSERT ad_events
  │                                                                       INCR Redis freq key
  │◀──── 200 image/gif (1×1 transparent GIF) ─────────────────────────────────────┘
  │
  │── GET /track/click?creative_id=xxx&redirect=https://... ──────────▶ INSERT ad_events
  │◀──── 302 Redirect → Advertiser Landing Page ───────────────────────────────────┘
```

---

## Features

- 🎯 **Ad Serving** — Selects the highest-bid eligible creative per request
- 📊 **Impression Tracking** — 1×1 transparent GIF pixel, fire-and-forget
- 🖱️ **Click Tracking** — Server-side redirect with event recording
- 💰 **Daily Budget Caps** — Per-campaign daily spend limits enforced in SQL
- 🔒 **Frequency Capping** — Max 3 impressions/IP/day (Redis-backed, DB fallback)
- ⚡ **Redis Caching** — Ad candidate list cached with configurable TTL (default 60s)
- 📈 **Reporting API** — Impressions, clicks, CTR%, and spend per creative
- 🌐 **Publisher Demo** — Dark-mode HTML page with async ad loading + status panel

---

## Tech Stack

| Layer      | Technology              |
|------------|-------------------------|
| Backend    | Go 1.21+ / Gin v1.12    |
| Database   | PostgreSQL (Neon Cloud) |
| Cache      | Redis (optional)        |
| Frontend   | Vanilla HTML + JS       |

---

## Project Structure

```
adserver/
├── config/
│   └── config.go            # Environment config loader
├── db/
│   ├── db.go                # PostgreSQL & Redis connection pools
│   ├── migrate/
│   │   └── main.go          # Go-based migration runner (no psql needed)
│   └── migrations/
│       └── 001_create_tables.sql  # Schema + seed data
├── handlers/
│   ├── ad_handler.go        # /ad, /track/impression, /track/click, /health
│   └── handlers.go          # Campaign/Creative CRUD + reporting endpoints
├── middleware/
│   └── middleware.go        # Request logger + CORS
├── models/
│   └── models.go            # Domain structs (Campaign, Creative, AdEvent, etc.)
├── services/
│   ├── ad_service.go        # Ad selection, frequency cap, Redis cache, tracking
│   └── crud_service.go      # Campaign & creative CRUD, event listing
├── static/
│   └── publisher.html       # Live publisher demo page
├── main.go                  # Entry point — wires everything together
├── .env.example             # Environment variable template
└── README.md
```

---

## Setup & Run Instructions

### Prerequisites

- Go 1.21+
- PostgreSQL (or a [Neon](https://neon.tech) free account)
- Redis (optional — caching & frequency capping gracefully degrade without it)

### 1. Clone the Repository

```bash
git clone https://github.com/immobarok/adserver.git
cd adserver
go mod download
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env`:

```env
PORT=8080
GIN_MODE=debug

DB_HOST=your-neon-host.neon.tech
DB_PORT=5432
DB_USER=neondb_owner
DB_PASSWORD=your_password
DB_NAME=neondb
DB_SSLMODE=require

REDIS_ADDR=localhost:6379   # leave blank if no Redis
FREQUENCY_CAP=3
CACHE_TTL_SECONDS=60
```

### 3. Run Database Migration

No `psql` required — use the built-in Go runner:

```bash
go run db/migrate/main.go
```

This creates the `campaigns`, `creatives`, and `ad_events` tables and inserts 3 seed campaigns with 5 creatives.

### 4. Start the Server

```bash
go run main.go
```

Server starts at **http://localhost:8080**

### 5. Open the Publisher Demo

```
http://localhost:8080/static/publisher.html
```

---

## API Reference

### Ad Serving & Tracking

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/ad` | Returns best eligible ad as JSON |
| `GET` | `/track/impression?creative_id=<id>` | Records impression, returns 1×1 GIF |
| `GET` | `/track/click?creative_id=<id>&redirect=<url>` | Records click, 302 redirect |
| `GET` | `/health` | Health check |

### Campaigns

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/campaigns` | List all campaigns |
| `POST` | `/api/v1/campaigns` | Create campaign |
| `GET` | `/api/v1/campaigns/:id` | Get campaign + today's spend |
| `PATCH` | `/api/v1/campaigns/:id` | Update campaign fields |
| `DELETE` | `/api/v1/campaigns/:id` | Delete campaign |

**POST /api/v1/campaigns body:**
```json
{ "name": "Summer Sale", "budget": 5000, "daily_limit": 500 }
```

### Creatives

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/creatives?campaign_id=<id>` | List creatives (optional filter) |
| `POST` | `/api/v1/creatives` | Create creative |
| `DELETE` | `/api/v1/creatives/:id` | Delete creative |

**POST /api/v1/creatives body:**
```json
{
  "campaign_id": "uuid",
  "title": "Summer Banner",
  "image_url": "https://example.com/banner.jpg",
  "click_url": "https://example.com/landing",
  "bid_price": 0.05
}
```

### Reporting

| Method | Endpoint | Query Params |
|--------|----------|-------------|
| `GET` | `/api/v1/report` | `from`, `to` (YYYY-MM-DD), `campaign_id` |
| `GET` | `/api/v1/events` | `from`, `to`, `campaign_id`, `event_type` |

**Example:**
```
GET /api/v1/report?from=2026-10-01&to=2026-10-04
```

---

## Bonus Features Implemented

| Feature | Implementation |
|---------|---------------|
| ✅ **Redis Caching** | Ad candidates cached for 60s (configurable). Invalidated on creative delete. |
| ✅ **Frequency Capping** | Max 3 impressions/IP/day via Redis counter with midnight TTL. Falls back gracefully if Redis is down. |
| ✅ **Daily Budget Limits** | SQL subquery checks today's spend vs `daily_limit` before serving each ad. |
| ✅ **Publisher Demo Page** | Dark-mode HTML page with async ad loading, skeleton loaders, 1×1 pixel tracking, click redirect, and live status panel. |
| ✅ **Fire-and-forget Tracking** | Impression/click recording happens in a goroutine so it never blocks the pixel/redirect response. |
| ✅ **Graceful Redis Fallback** | Server starts and serves ads normally even if Redis is unavailable. |

---

## Assumptions Made

1. **Highest-bid wins** — Ad selection uses a simple first-price auction: the creative with the highest `bid_price` that passes all filters is served. No randomization or weighted selection.
2. **IP-based frequency cap** — Frequency capping is done per client IP. Behind NAT (e.g., shared office Wi-Fi), multiple users share one IP. A cookie/device-fingerprint approach would be more accurate but requires client-side JS.
3. **Impression cost = bid price** — The cost recorded per impression equals the creative's `bid_price`. In a real RTB system this would be the clearing price from an auction.
4. **Click cost = 0** — Clicks are tracked as events but carry no additional cost (CPM model assumed).
5. **No authentication** — The management API (`/api/v1/...`) has no auth layer. In production, JWT or API-key middleware would be required.
6. **Redis is optional** — Without Redis, frequency capping is disabled (all IPs serve indefinitely) and ad candidates are re-queried from PostgreSQL on every request.

---

## Potential Improvements

| Area | Improvement |
|------|-------------|
| **Auth** | Add JWT middleware to protect `/api/v1` endpoints |
| **Auction** | Implement second-price (Vickrey) auction logic |
| **Targeting** | Add geo-targeting, device-type, and time-of-day targeting |
| **Frequency Cap** | Use device fingerprint or cookie instead of raw IP |
| **Rate Limiting** | Add per-IP rate limiting on `/ad` to prevent abuse |
| **Metrics** | Expose Prometheus metrics endpoint (`/metrics`) |
| **Migrations** | Use a proper migration tool (e.g., `golang-migrate`) with versioning |
| **Ad Formats** | Support video ads, native ads, and rich media in addition to banners |
| **A/B Testing** | Weight-based creative rotation for split testing |
| **Dashboard** | Admin UI for campaign management and live reporting charts |

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `GIN_MODE` | `debug` | `debug` or `release` |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | PostgreSQL user |
| `DB_PASSWORD` | — | PostgreSQL password |
| `DB_NAME` | `adserver` | Database name |
| `DB_SSLMODE` | `disable` | `disable` or `require` |
| `REDIS_ADDR` | `localhost:6379` | Redis address (optional) |
| `FREQUENCY_CAP` | `3` | Max impressions per IP per day |
| `CACHE_TTL_SECONDS` | `60` | Redis cache TTL for ad candidates |

---

## License

MIT
