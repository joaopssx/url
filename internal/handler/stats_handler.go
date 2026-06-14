package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"urlshortener/internal/service"
)

type StatsHandler struct {
	statsService service.StatsService
}

func NewStatsHandler(statsService service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	code := c.Param("code")

	stats, err := h.statsService.GetStats(code)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
