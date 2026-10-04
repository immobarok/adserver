# AdServer

A high-performance programmatic ad server built with **Go + Gin**, **PostgreSQL (Neon)**, and **Redis**.

## Features

- 🎯 **Ad Serving** – Selects the highest-bid eligible creative per request
- 📊 **Impression Tracking** – 1×1 transparent GIF pixel tracking
- 🖱️ **Click Tracking** – Server-side redirect with event recording
- 💰 **Daily Budget Caps** – Campaign daily spend limits enforced at query time
- 🔒 **Frequency Capping** – Max 3 impressions per IP per day (Redis-backed)
- ⚡ **Redis Caching** – Ad candidate list cached with configurable TTL
- 📈 **Reporting API** – Impressions, clicks, CTR, and spend per creative
- 🌐 **Publisher Demo** – Live HTML page showcasing async ad tag integration

## Tech Stack

| Layer      | Technology          |
|------------|---------------------|
| Backend    | Go 1.21+ / Gin      |
| Database   | PostgreSQL (Neon)   |
| Cache      | Redis (optional)    |
| Deploy     | Any (Render, Fly.io)|

## Project Structure

```
adserver/
├── config/          # Environment config loader
├── db/              # PostgreSQL & Redis connections
├── handlers/        # Gin HTTP handlers
│   ├── ad_handler.go   # /ad, /track/impression, /track/click
│   └── handlers.go     # CRUD & reporting endpoints
├── middleware/      # Logger, CORS
├── models/          # Domain structs
├── services/        # Business logic
│   ├── ad_service.go   # Ad selection, tracking, reporting
│   └── crud_service.go # Campaign & creative CRUD
├── static/          # Publisher demo page
│   └── publisher.html
├── db/migrations/
│   └── 001_create_tables.sql
├── main.go
├── .env.example
└── README.md
```

## Getting Started

### 1. Clone & Install

```bash
git clone https://github.com/immobarok/adserver.git
cd adserver
go mod download
```

### 2. Configure

```bash
cp .env.example .env
# Edit .env with your DB credentials
```

### 3. Run Database Migration

```bash
psql "$DATABASE_URL" -f db/migrations/001_create_tables.sql
```

### 4. Start the Server

```bash
go run main.go
```

Server starts on `http://localhost:8080`

## API Endpoints

### Ad Serving

| Method | Endpoint                | Description                          |
|--------|-------------------------|--------------------------------------|
| GET    | `/ad`                   | Serve the best eligible ad (JSON)    |
| GET    | `/track/impression`     | Record impression + return 1×1 GIF   |
| GET    | `/track/click`          | Record click + redirect to ad URL    |
| GET    | `/health`               | Health check                         |

### Campaigns

| Method | Endpoint               | Description          |
|--------|------------------------|----------------------|
| GET    | `/api/v1/campaigns`    | List all campaigns   |
| POST   | `/api/v1/campaigns`    | Create campaign      |
| GET    | `/api/v1/campaigns/:id`| Get campaign detail  |
| PATCH  | `/api/v1/campaigns/:id`| Update campaign      |
| DELETE | `/api/v1/campaigns/:id`| Delete campaign      |

### Creatives

| Method | Endpoint               | Description          |
|--------|------------------------|----------------------|
| GET    | `/api/v1/creatives`    | List creatives       |
| POST   | `/api/v1/creatives`    | Create creative      |
| DELETE | `/api/v1/creatives/:id`| Delete creative      |

### Reporting

| Method | Endpoint               | Query Params                              |
|--------|------------------------|-------------------------------------------|
| GET    | `/api/v1/report`       | `from`, `to` (YYYY-MM-DD), `campaign_id` |
| GET    | `/api/v1/events`       | `from`, `to`, `campaign_id`, `event_type`|

## Publisher Demo

Open `http://localhost:8080/static/publisher.html` to see live ad loading with impression pixel and click tracking in action.

## Environment Variables

| Variable           | Default       | Description                        |
|--------------------|---------------|------------------------------------|
| `PORT`             | `8080`        | Server port                        |
| `GIN_MODE`         | `debug`       | `debug` or `release`               |
| `DB_HOST`          | `localhost`   | PostgreSQL host                    |
| `DB_PORT`          | `5432`        | PostgreSQL port                    |
| `DB_USER`          | `postgres`    | PostgreSQL user                    |
| `DB_PASSWORD`      | –             | PostgreSQL password                |
| `DB_NAME`          | `adserver`    | Database name                      |
| `DB_SSLMODE`       | `disable`     | `disable` or `require`             |
| `REDIS_ADDR`       | `localhost:6379` | Redis address (optional)        |
| `FREQUENCY_CAP`    | `3`           | Max impressions per IP per day     |
| `CACHE_TTL_SECONDS`| `60`          | Redis cache TTL for ad candidates  |

## License

MIT
