package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"time"

	"github.com/google/uuid"

	"urlshortener/internal/model"
	"urlshortener/internal/repository"
)

var (
	ErrDuplicateURL = errors.New("duplicate url")
	ErrNotFound     = errors.New("url not found")
	ErrExpired      = errors.New("url expired")
	ErrInvalidURL   = errors.New("invalid url")
	ErrForbidden    = errors.New("forbidden")
)

type URLService interface {
	Shorten(originalURL string, userID *string, expiresAt *time.Time, customCode *string) (*model.URL, error)
	Redirect(code, ip string) (*model.URL, error)
	GetUserURLs(userID string) ([]model.URL, error)
	UpdateURL(code, userID string, newURL *string, newExpiry *time.Time) (*model.URL, error)
	DeleteURL(code, userID string) error
}

type urlService struct {
	urlRepo repository.URLRepository
	baseURL string
}

func NewURLService(urlRepo repository.URLRepository, baseURL string) URLService {
	return &urlService{
		urlRepo: urlRepo,
		baseURL: baseURL,
	}
}

func (s *urlService) Shorten(originalURL string, userID *string, expiresAt *time.Time, customCode *string) (*model.URL, error) {
	_, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return nil, ErrInvalidURL
	}

	if userID != nil {
		existing, err := s.urlRepo.FindByOriginalURLAndUser(originalURL, *userID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrDuplicateURL
		}
	}

	code := ""
	if customCode != nil && *customCode != "" {
		code = *customCode
	} else {
		code, err = generateRandomCode(7)
		if err != nil {
			return nil, err
		}
	}

	u := &model.URL{
		ID:          uuid.New().String(),
		ShortCode:   code,
		OriginalURL: originalURL,
		UserID:      userID,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   expiresAt,
		AccessCount: 0,
	}

	err = s.urlRepo.Create(u)
	if err != nil {
		return nil, err
	}

	u.ShortURL = s.baseURL + "/" + u.ShortCode
	return u, nil
}

func (s *urlService) Redirect(code, ip string) (*model.URL, error) {
	u, err := s.urlRepo.FindByCode(code)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}

	if u.DeletedAt != nil {
		return nil, ErrNotFound
	}

	if u.ExpiresAt != nil && u.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpired
	}

	go func() {
		_ = s.urlRepo.IncrementAccessCount(u.ID)
		_ = s.urlRepo.RecordAccess(u.ID, ip)
	}()

	return u, nil
}

func generateRandomCode(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

func (s *urlService) GetUserURLs(userID string) ([]model.URL, error) {
	urls, err := s.urlRepo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	for i := range urls {
		urls[i].ShortURL = s.baseURL + "/" + urls[i].ShortCode
	}
	if urls == nil {
		urls = []model.URL{}
	}
	return urls, nil
}

func (s *urlService) UpdateURL(code, userID string, newURL *string, newExpiry *time.Time) (*model.URL, error) {
	u, err := s.urlRepo.FindByCode(code)
	if err != nil {
		return nil, err
	}
	if u == nil || u.DeletedAt != nil {
		return nil, ErrNotFound
	}

	if u.UserID == nil || *u.UserID != userID {
		return nil, ErrForbidden
	}

	updates := make(map[string]interface{})
	if newURL != nil {
		_, err := url.ParseRequestURI(*newURL)
		if err != nil {
			return nil, ErrInvalidURL
		}
		updates["original_url"] = *newURL
	}
	if newExpiry != nil {
		updates["expires_at"] = newExpiry
	}

	if len(updates) > 0 {
		err = s.urlRepo.Update(u.ID, updates)
		if err != nil {
			return nil, err
		}
	}

	updated, err := s.urlRepo.FindByCode(code)
	if err != nil {
		return nil, err
	}
	updated.ShortURL = s.baseURL + "/" + updated.ShortCode
	return updated, nil
}

func (s *urlService) DeleteURL(code, userID string) error {
	u, err := s.urlRepo.FindByCode(code)
	if err != nil {
		return err
	}
	if u == nil || u.DeletedAt != nil {
		return ErrNotFound
	}
	if u.UserID == nil || *u.UserID != userID {
		return ErrForbidden
	}

	err = s.urlRepo.SoftDelete(u.ID, userID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return ErrNotFound
		}
		return err
	}
	return nil
}
