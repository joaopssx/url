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
	GetAccessStats(urlID string) ([]model.DayStat, []model.WeekStat, []model.MonthlyStat, error)
}

type urlRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) URLRepository {
	return &urlRepository{db: db}
}

func (r *urlRepository) Create(url *model.URL) error {
	query := `INSERT INTO urls (id, short_code, original_url, user_id, created_at, expires_at, access_count, deleted_at, webhook_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, url.ID, url.ShortCode, url.OriginalURL, url.UserID, url.CreatedAt, url.ExpiresAt, url.AccessCount, url.DeletedAt, url.WebhookURL)
	return err
}

func (r *urlRepository) FindByCode(code string) (*model.URL, error) {
	query := `SELECT id, short_code, original_url, user_id, created_at, expires_at, access_count, deleted_at, webhook_url FROM urls WHERE short_code = ?`
	row := r.db.QueryRow(query, code)

	var u model.URL
	err := row.Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.UserID, &u.CreatedAt, &u.ExpiresAt, &u.AccessCount, &u.DeletedAt, &u.WebhookURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *urlRepository) FindByOriginalURLAndUser(originalURL, userID string) (*model.URL, error) {
	query := `SELECT id, short_code, original_url, user_id, created_at, expires_at, access_count, deleted_at, webhook_url FROM urls WHERE original_url = ? AND user_id = ?`
	row := r.db.QueryRow(query, originalURL, userID)

	var u model.URL
	err := row.Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.UserID, &u.CreatedAt, &u.ExpiresAt, &u.AccessCount, &u.DeletedAt, &u.WebhookURL)
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
	query := `SELECT id, short_code, original_url, user_id, created_at, expires_at, access_count, deleted_at, webhook_url FROM urls WHERE user_id = ? AND deleted_at IS NULL`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []model.URL
	for rows.Next() {
		var u model.URL
		err := rows.Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.UserID, &u.CreatedAt, &u.ExpiresAt, &u.AccessCount, &u.DeletedAt, &u.WebhookURL)
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

func (r *urlRepository) GetAccessStats(urlID string) ([]model.DayStat, []model.WeekStat, []model.MonthlyStat, error) {
	dailyQuery := `SELECT strftime('%Y-%m-%d', accessed_at) as day, COUNT(*) as count FROM url_accesses WHERE url_id = ? GROUP BY day ORDER BY day DESC LIMIT 30`
	dailyRows, err := r.db.Query(dailyQuery, urlID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer dailyRows.Close()

	var daily []model.DayStat
	for dailyRows.Next() {
		var s model.DayStat
		if err := dailyRows.Scan(&s.Day, &s.Count); err != nil {
			return nil, nil, nil, err
		}
		daily = append(daily, s)
	}

	weeklyQuery := `SELECT strftime('%Y-W%W', accessed_at) as week, COUNT(*) as count FROM url_accesses WHERE url_id = ? GROUP BY week ORDER BY week DESC LIMIT 12`
	weeklyRows, err := r.db.Query(weeklyQuery, urlID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer weeklyRows.Close()

	var weekly []model.WeekStat
	for weeklyRows.Next() {
		var s model.WeekStat
		if err := weeklyRows.Scan(&s.Week, &s.Count); err != nil {
			return nil, nil, nil, err
		}
		weekly = append(weekly, s)
	}

	monthlyQuery := `SELECT strftime('%Y-%m', accessed_at) as month, COUNT(*) as count FROM url_accesses WHERE url_id = ? GROUP BY month ORDER BY month DESC LIMIT 12`
	monthlyRows, err := r.db.Query(monthlyQuery, urlID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer monthlyRows.Close()

	var monthly []model.MonthlyStat
	for monthlyRows.Next() {
		var s model.MonthlyStat
		if err := monthlyRows.Scan(&s.Month, &s.Count); err != nil {
			return nil, nil, nil, err
		}
		monthly = append(monthly, s)
	}

	if daily == nil {
		daily = []model.DayStat{}
	}
	if weekly == nil {
		weekly = []model.WeekStat{}
	}
	if monthly == nil {
		monthly = []model.MonthlyStat{}
	}

	return daily, weekly, monthly, nil
}
