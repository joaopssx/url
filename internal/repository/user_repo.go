package repository

import (
	"database/sql"
	"urlshortener/internal/model"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByEmail(email string) (*model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	query := `INSERT INTO users (id, email, password_hash, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, user.ID, user.Email, user.PasswordHash, user.CreatedAt)
	return err
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	query := `SELECT id, email, password_hash, created_at FROM users WHERE email = ?`
	row := r.db.QueryRow(query, email)

	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
