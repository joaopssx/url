package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"urlshortener/internal/model"
)

type URLRepository interface {
	Create(url *model.URL) error
	FindByCode(code string) (*model.URL, error)
	FindByOriginalURLAndUser(originalURL, userID string) (*model.URL, error)
	IncrementAccessCount(id string) error
	RecordAccess(urlID, ip string) error
}

type urlRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) URLRepository {
	return &urlRepository{db: db}
}

func (r *urlRepository) Create(url *model.URL) error {
	query := `INSERT INTO urls (id, short_code, original_url, user_id, created_at, expires_at, access_count, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, url.ID, url.ShortCode, url.OriginalURL, url.UserID, url.CreatedAt, url.ExpiresAt, url.AccessCount, url.DeletedAt)
	return err
}

func (r *urlRepository) FindByCode(code string) (*model.URL, error) {
	query := `SELECT id, short_code, original_url, user_id, created_at, expires_at, access_count, deleted_at FROM urls WHERE short_code = ?`
	row := r.db.QueryRow(query, code)

	var u model.URL
	err := row.Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.UserID, &u.CreatedAt, &u.ExpiresAt, &u.AccessCount, &u.DeletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *urlRepository) FindByOriginalURLAndUser(originalURL, userID string) (*model.URL, error) {
	query := `SELECT id, short_code, original_url, user_id, created_at, expires_at, access_count, deleted_at FROM urls WHERE original_url = ? AND user_id = ?`
	row := r.db.QueryRow(query, originalURL, userID)

	var u model.URL
	err := row.Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.UserID, &u.CreatedAt, &u.ExpiresAt, &u.AccessCount, &u.DeletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *urlRepository) IncrementAccessCount(id string) error {
	query := `UPDATE urls SET access_count = access_count + 1 WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *urlRepository) RecordAccess(urlID, ip string) error {
	query := `INSERT INTO url_accesses (id, url_id, accessed_at, ip) VALUES (?, ?, ?, ?)`
	id := uuid.New().String()
	_, err := r.db.Exec(query, id, urlID, time.Now().UTC(), ip)
	return err
}
