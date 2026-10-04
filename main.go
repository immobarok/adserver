package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"adserver/config"
	"adserver/db"
	"adserver/handlers"
	"adserver/middleware"
)

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

	// ── Static files (publisher demo page) ──────────────────────────────
	r.Static("/static", "./static")

	// ── Health ──────────────────────────────────────────────────────────
	r.GET("/health", handlers.HealthCheck)

	// ── Ad serving & tracking ───────────────────────────────────────────
	r.GET("/ad", handlers.ServeAd)
	r.GET("/track/impression", handlers.TrackImpression)
	r.GET("/track/click", handlers.TrackClick)

	// ── REST API v1 ─────────────────────────────────────────────────────
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
