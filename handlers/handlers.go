package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"adserver/models"
	"adserver/services"
)

// ─── Campaigns ───────────────────────────────────────────────────────────────

// ListCampaigns godoc
// GET /api/v1/campaigns
func ListCampaigns(c *gin.Context) {
	campaigns, err := services.ListCampaigns(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": campaigns, "count": len(campaigns)})
}

// CreateCampaign godoc
// POST /api/v1/campaigns
func CreateCampaign(c *gin.Context) {
	var req models.CreateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	campaign, err := services.CreateCampaign(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": campaign})
}

// GetCampaign godoc
// GET /api/v1/campaigns/:id
func GetCampaign(c *gin.Context) {
	campaign, err := services.GetCampaign(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	spend, _ := services.CampaignDailySpend(c.Request.Context(), campaign.ID)
	c.JSON(http.StatusOK, gin.H{"data": campaign, "today_spend": spend})
}

// UpdateCampaign godoc
// PATCH /api/v1/campaigns/:id
func UpdateCampaign(c *gin.Context) {
	var req models.UpdateCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	campaign, err := services.UpdateCampaign(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": campaign})
}

// DeleteCampaign godoc
// DELETE /api/v1/campaigns/:id
func DeleteCampaign(c *gin.Context) {
	if err := services.DeleteCampaign(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "campaign deleted"})
}

// ─── Creatives ───────────────────────────────────────────────────────────────

// ListCreatives godoc
// GET /api/v1/creatives?campaign_id=xxx
func ListCreatives(c *gin.Context) {
	creatives, err := services.ListCreatives(c.Request.Context(), c.Query("campaign_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": creatives, "count": len(creatives)})
}

// CreateCreative godoc
// POST /api/v1/creatives
func CreateCreative(c *gin.Context) {
	var req models.CreateCreativeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	creative, err := services.CreateCreative(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": creative})
}

// DeleteCreative godoc
// DELETE /api/v1/creatives/:id
func DeleteCreative(c *gin.Context) {
	if err := services.DeleteCreative(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "creative deleted"})
}

// ─── Reporting ───────────────────────────────────────────────────────────────

// GetReport godoc
// GET /api/v1/report?from=2026-01-01&to=2026-12-31&campaign_id=xxx
func GetReport(c *gin.Context) {
	fromStr := c.DefaultQuery("from", time.Now().Format("2006-01-02"))
	toStr := c.DefaultQuery("to", time.Now().Format("2006-01-02"))

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' date, use YYYY-MM-DD"})
		return
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' date, use YYYY-MM-DD"})
		return
	}
	// Extend 'to' to end of day
	to = to.Add(24*time.Hour - time.Second)

	report, err := services.GetReport(c.Request.Context(), from, to, c.Query("campaign_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  report,
		"count": len(report),
		"from":  fromStr,
		"to":    toStr,
	})
}

// ListAdEvents godoc
// GET /api/v1/events?campaign_id=xxx&event_type=impression&from=...&to=...
func ListAdEvents(c *gin.Context) {
	fromStr := c.DefaultQuery("from", time.Now().Format("2006-01-02"))
	toStr := c.DefaultQuery("to", time.Now().Format("2006-01-02"))

	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	to = to.Add(24*time.Hour - time.Second)

	events, err := services.ListAdEvents(
		c.Request.Context(),
		c.Query("campaign_id"),
		c.Query("event_type"),
		from, to,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events, "count": len(events)})
}
