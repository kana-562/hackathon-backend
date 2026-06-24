package mysql

import (
	"database/sql"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) repository.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) FindAll() ([]domain.HobbyCategory, error) {
	rows, err := r.db.Query(
		`SELECT id, name, COALESCE(slug,''), COALESCE(description,''), COALESCE(icon_name,''), sort_order, created_at FROM hobby_categories ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.HobbyCategory
	for rows.Next() {
		var c domain.HobbyCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.IconName, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *categoryRepository) FindByID(id int64) (*domain.HobbyCategory, error) {
	c := &domain.HobbyCategory{}
	err := r.db.QueryRow(
		`SELECT id, name, COALESCE(slug,''), COALESCE(description,''), COALESCE(icon_name,''), sort_order, created_at FROM hobby_categories WHERE id = ?`, id,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.IconName, &c.SortOrder, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *categoryRepository) FindHobbiesByCategoryID(categoryID int64) ([]domain.Hobby, error) {
	rows, err := r.db.Query(
		`SELECT id, category_id, name, slug, description, sort_order, created_at FROM hobbies WHERE category_id = ? ORDER BY sort_order ASC`, categoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hobbies []domain.Hobby
	for rows.Next() {
		var h domain.Hobby
		if err := rows.Scan(&h.ID, &h.CategoryID, &h.Name, &h.Slug, &h.Description, &h.SortOrder, &h.CreatedAt); err != nil {
			return nil, err
		}
		hobbies = append(hobbies, h)
	}
	return hobbies, rows.Err()
}

func (r *categoryRepository) FindHobbyByID(id int64) (*domain.Hobby, error) {
	h := &domain.Hobby{}
	err := r.db.QueryRow(
		`SELECT id, category_id, name, slug, description, sort_order, created_at FROM hobbies WHERE id = ?`, id,
	).Scan(&h.ID, &h.CategoryID, &h.Name, &h.Slug, &h.Description, &h.SortOrder, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (r *categoryRepository) FindHobbyByName(name string) (*domain.Hobby, error) {
	h := &domain.Hobby{}
	err := r.db.QueryRow(
		`SELECT id, category_id, name, slug, description, sort_order, created_at FROM hobbies WHERE name = ?`, name,
	).Scan(&h.ID, &h.CategoryID, &h.Name, &h.Slug, &h.Description, &h.SortOrder, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return h, nil
}
