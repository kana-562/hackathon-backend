package mysql

import (
	"database/sql"
	"time"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type favoriteRepository struct {
	db *sql.DB
}

func NewFavoriteRepository(db *sql.DB) repository.FavoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) Add(userID, setID int64) error {
	_, err := r.db.Exec(
		`INSERT IGNORE INTO favorites (user_id, starter_set_id, created_at) VALUES (?, ?, ?)`,
		userID, setID, time.Now(),
	)
	return err
}

func (r *favoriteRepository) Remove(userID, setID int64) error {
	_, err := r.db.Exec(
		`DELETE FROM favorites WHERE user_id = ? AND starter_set_id = ?`,
		userID, setID,
	)
	return err
}

func (r *favoriteRepository) Exists(userID, setID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM favorites WHERE user_id = ? AND starter_set_id = ?`,
		userID, setID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *favoriteRepository) CountBySet(setID int64) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM favorites WHERE starter_set_id = ?`, setID).Scan(&count)
	return count, err
}

type startPlanRepository struct {
	db *sql.DB
}

func NewStartPlanRepository(db *sql.DB) repository.StartPlanRepository {
	return &startPlanRepository{db: db}
}

func (r *startPlanRepository) Create(plan *domain.StartPlan, steps []domain.StartPlanStep) (int64, error) {
	dbTx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = dbTx.Rollback()
		}
	}()

	now := time.Now()
	result, err := dbTx.Exec(`
		INSERT INTO start_plans (transaction_id, starter_set_id, user_id, title, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		plan.TransactionID, plan.StarterSetID, plan.UserID, plan.Title, now,
	)
	if err != nil {
		return 0, err
	}

	planID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, step := range steps {
		_, err = dbTx.Exec(`
			INSERT INTO start_plan_steps (start_plan_id, day_no, title, body)
			VALUES (?, ?, ?, ?)`,
			planID, step.DayNo, step.Title, step.Body,
		)
		if err != nil {
			return 0, err
		}
	}

	if err = dbTx.Commit(); err != nil {
		return 0, err
	}

	return planID, nil
}

func (r *startPlanRepository) FindByTransaction(transactionID int64) (*domain.StartPlan, error) {
	plan := &domain.StartPlan{}
	err := r.db.QueryRow(`
		SELECT id, transaction_id, starter_set_id, user_id, title, created_at
		FROM start_plans WHERE transaction_id = ?`, transactionID,
	).Scan(&plan.ID, &plan.TransactionID, &plan.StarterSetID, &plan.UserID, &plan.Title, &plan.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(`
		SELECT id, start_plan_id, day_no, title, body FROM start_plan_steps WHERE start_plan_id = ? ORDER BY day_no ASC`, plan.ID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var step domain.StartPlanStep
		if err := rows.Scan(&step.ID, &step.StartPlanID, &step.DayNo, &step.Title, &step.Body); err != nil {
			return nil, err
		}
		plan.Steps = append(plan.Steps, step)
	}

	return plan, rows.Err()
}
