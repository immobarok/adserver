package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"adserver/config"
	"adserver/db"
	"adserver/handlers"
	"adserver/middleware"
)

// Embed the entire static directory into the binary at compile time.
// This means the binary is self-contained — no filesystem dependency on Render.
//
//go:embed static
var staticFiles embed.FS

func main() {
	// 1. Load configuration
	config.Load()

	// 2. Connect to databases
	db.ConnectPostgres()
	db.ConnectRedis()

	// 3. Set Gin mode
	gin.SetMode(config.App.GinMode)

	// 4. Create router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// ── Embedded static files ────────────────────────────────────────────
	// Strips the leading "static" prefix so /static/admin.html works correctly
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("❌ Failed to create sub-filesystem: %v", err)
	}
	r.StaticFS("/static", http.FS(sub))

	// ── Health ───────────────────────────────────────────────────────────
	r.GET("/health", handlers.HealthCheck)

	// ── Ad serving & tracking ────────────────────────────────────────────
	r.GET("/ad", handlers.ServeAd)
	r.GET("/track/impression", handlers.TrackImpression)
	r.GET("/track/click", handlers.TrackClick)

	// ── REST API v1 ──────────────────────────────────────────────────────
	api := r.Group("/api/v1")
	{
		// Campaigns
		api.GET("/campaigns", handlers.ListCampaigns)
		api.POST("/campaigns", handlers.CreateCampaign)
		api.GET("/campaigns/:id", handlers.GetCampaign)
		api.PATCH("/campaigns/:id", handlers.UpdateCampaign)
		api.DELETE("/campaigns/:id", handlers.DeleteCampaign)

		// Creatives
		api.GET("/creatives", handlers.ListCreatives)
		api.POST("/creatives", handlers.CreateCreative)
		api.DELETE("/creatives/:id", handlers.DeleteCreative)

		// Reporting & Events
		api.GET("/report", handlers.GetReport)
		api.GET("/events", handlers.ListAdEvents)
	}

	addr := ":" + config.App.Port
	log.Printf("🚀 AdServer listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
