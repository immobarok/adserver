package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"adserver/services"
)

// ServeAd godoc
// GET /ad
// Returns the best eligible ad for the requesting IP.
func ServeAd(c *gin.Context) {
	ip := c.ClientIP()
	ad, err := services.GetAdCandidate(c.Request.Context(), ip)
	if err != nil {
		c.JSON(http.StatusNoContent, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ad)
}

// TrackImpression godoc
// GET /track/impression?creative_id=xxx
// Records an impression and returns a 1x1 transparent GIF.
func TrackImpression(c *gin.Context) {
	creativeID := c.Query("creative_id")
	if creativeID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	referrer := c.Request.Referer()

	// Fire-and-forget in background (don't block pixel response)
	go func() {
		_ = services.RecordImpression(
			c.Copy().Request.Context(),
			creativeID, ip, userAgent, referrer,
		)
	}()

	// Return 1x1 transparent GIF
	pixel := []byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61,
		0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00,
		0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00,
		0x21, 0xF9, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x2C, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
		0x02, 0x02, 0x44, 0x01, 0x00,
		0x3B,
	}
	c.Data(http.StatusOK, "image/gif", pixel)
}

// TrackClick godoc
// GET /track/click?creative_id=xxx&redirect=https://...
// Records a click and redirects the user to the ad destination.
func TrackClick(c *gin.Context) {
	creativeID := c.Query("creative_id")
	redirect := c.Query("redirect")

	if creativeID == "" || redirect == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "creative_id and redirect are required"})
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	referrer := c.Request.Referer()

	go func() {
		_ = services.RecordClick(
			c.Copy().Request.Context(),
			creativeID, ip, userAgent, referrer,
		)
	}()

	c.Redirect(http.StatusFound, redirect)
}

// HealthCheck godoc
// GET /health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "adserver",
	})
}
