package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"urlshortener/internal/service"
)

type URLHandler struct {
	urlService service.URLService
}

func NewURLHandler(urlService service.URLService) *URLHandler {
	return &URLHandler{urlService: urlService}
}

type shortenRequest struct {
	URL       string     `json:"url" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
	Custom    *string    `json:"custom"`
}

func (h *URLHandler) Shorten(c *gin.Context) {
	var req shortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *string
	if uid, exists := c.Get("user_id"); exists {
		strUID := uid.(string)
		userID = &strUID
	}

	u, err := h.urlService.Shorten(req.URL, userID, req.ExpiresAt, req.Custom)
	if err != nil {
		switch err {
		case service.ErrInvalidURL:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case service.ErrDuplicateURL:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to shorten url"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"short_code":   u.ShortCode,
		"short_url":    u.ShortURL,
		"original_url": u.OriginalURL,
		"created_at":   u.CreatedAt,
		"expires_at":   u.ExpiresAt,
	})
}

func (h *URLHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	ip := c.ClientIP()

	u, err := h.urlService.Redirect(code, ip)
	if err != nil {
		switch err {
		case service.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
		case service.ErrExpired:
			c.JSON(http.StatusGone, gin.H{"error": "url expired"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.Redirect(http.StatusFound, u.OriginalURL)
}
