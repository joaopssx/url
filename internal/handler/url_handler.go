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

func (h *URLHandler) GetUserURLs(c *gin.Context) {
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := uid.(string)

	urls, err := h.urlService.GetUserURLs(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get urls"})
		return
	}

	type resp struct {
		ShortCode   string     `json:"short_code"`
		ShortURL    string     `json:"short_url"`
		OriginalURL string     `json:"original_url"`
		AccessCount int        `json:"access_count"`
		CreatedAt   time.Time  `json:"created_at"`
		ExpiresAt   *time.Time `json:"expires_at"`
	}

	result := make([]resp, 0, len(urls))
	for _, u := range urls {
		result = append(result, resp{
			ShortCode:   u.ShortCode,
			ShortURL:    u.ShortURL,
			OriginalURL: u.OriginalURL,
			AccessCount: u.AccessCount,
			CreatedAt:   u.CreatedAt,
			ExpiresAt:   u.ExpiresAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

func (h *URLHandler) DeleteURL(c *gin.Context) {
	code := c.Param("code")
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := uid.(string)

	err := h.urlService.DeleteURL(code, userID)
	if err != nil {
		switch err {
		case service.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
		case service.ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete url"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

type updateRequest struct {
	URL       *string    `json:"url"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (h *URLHandler) UpdateURL(c *gin.Context) {
	code := c.Param("code")
	uid, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := uid.(string)

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.URL == nil && req.ExpiresAt == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field is required"})
		return
	}

	u, err := h.urlService.UpdateURL(code, userID, req.URL, req.ExpiresAt)
	if err != nil {
		switch err {
		case service.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
		case service.ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case service.ErrInvalidURL:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update url"})
		}
		return
	}

	c.JSON(http.StatusOK, u)
}
