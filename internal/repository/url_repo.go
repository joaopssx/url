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
	ListByUser(userID string) ([]model.URL, error)
	Update(id string, updates map[string]interface{}) error
	SoftDelete(id, userID string) error
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

func (r *urlRepository) ListByUser(userID string) ([]model.URL, error) {
	query := `SELECT id, short_code, original_url, user_id, created_at, expires_at, access_count, deleted_at FROM urls WHERE user_id = ? AND deleted_at IS NULL`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []model.URL
	for rows.Next() {
		var u model.URL
		err := rows.Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.UserID, &u.CreatedAt, &u.ExpiresAt, &u.AccessCount, &u.DeletedAt)
		if err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}

func (r *urlRepository) Update(id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	query := "UPDATE urls SET "
	args := []interface{}{}
	i := 1
	for k, v := range updates {
		if i > 1 {
			query += ", "
		}
		query += k + " = ?"
		args = append(args, v)
		i++
	}
	query += " WHERE id = ?"
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *urlRepository) SoftDelete(id, userID string) error {
	query := `UPDATE urls SET deleted_at = ? WHERE id = ? AND user_id = ? AND deleted_at IS NULL`
	res, err := r.db.Exec(query, time.Now().UTC(), id, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
