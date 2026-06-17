package mysql

import (
	"database/sql"
	"time"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(
		`INSERT INTO users (display_name, email, password_hash, avatar_url, rating_average, rating_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.DisplayName, user.Email, user.PasswordHash, user.AvatarURL,
		user.RatingAverage, user.RatingCount, now, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(
		`SELECT id, display_name, email, password_hash, avatar_url, rating_average, rating_count, created_at, updated_at
		FROM users WHERE email = ?`, email,
	).Scan(
		&u.ID, &u.DisplayName, &u.Email, &u.PasswordHash, &u.AvatarURL,
		&u.RatingAverage, &u.RatingCount, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) FindByID(id int64) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(
		`SELECT id, display_name, email, password_hash, avatar_url, rating_average, rating_count, created_at, updated_at
		FROM users WHERE id = ?`, id,
	).Scan(
		&u.ID, &u.DisplayName, &u.Email, &u.PasswordHash, &u.AvatarURL,
		&u.RatingAverage, &u.RatingCount, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepository) Update(user *domain.User) error {
	_, err := r.db.Exec(
		`UPDATE users SET display_name=?, avatar_url=?, rating_average=?, rating_count=?, updated_at=? WHERE id=?`,
		user.DisplayName, user.AvatarURL, user.RatingAverage, user.RatingCount, time.Now(), user.ID,
	)
	return err
}
