package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type setRepository struct {
	db *sql.DB
}

func NewSetRepository(db *sql.DB) repository.SetRepository {
	return &setRepository{db: db}
}

func (r *setRepository) FindAll(filter repository.SetFilter) ([]domain.StarterSet, error) {
	query := `
		SELECT s.id, s.seller_id, s.hobby_id, s.category_id, s.title, s.description, s.price,
		       s.status, s.beginner_score, s.readiness_score, s.value_score, s.estimated_new_price,
		       s.previous_owner_note, s.startable_summary, s.published_at, s.created_at, s.updated_at,
		       h.name AS hobby_name, hc.name AS category_name,
		       COALESCE(si.image_url, '') AS image_url
		FROM starter_sets s
		LEFT JOIN hobbies h ON s.hobby_id = h.id
		LEFT JOIN hobby_categories hc ON s.category_id = hc.id
		LEFT JOIN (
			SELECT starter_set_id, image_url FROM set_images WHERE sort_order = (
				SELECT MIN(sort_order) FROM set_images si2 WHERE si2.starter_set_id = set_images.starter_set_id
			)
		) si ON s.id = si.starter_set_id`

	var conditions []string
	var args []interface{}

	if filter.Status != "" {
		conditions = append(conditions, "s.status = ?")
		args = append(args, filter.Status)
	}
	if filter.CategoryID > 0 {
		conditions = append(conditions, "s.category_id = ?")
		args = append(args, filter.CategoryID)
	}
	if filter.HobbyID > 0 {
		conditions = append(conditions, "s.hobby_id = ?")
		args = append(args, filter.HobbyID)
	}
	if filter.MaxPrice > 0 {
		conditions = append(conditions, "s.price <= ?")
		args = append(args, filter.MaxPrice)
	}
	if filter.MinBeginnerScore > 0 {
		conditions = append(conditions, "s.beginner_score >= ?")
		args = append(args, filter.MinBeginnerScore)
	}
	if filter.MinReadinessScore > 0 {
		conditions = append(conditions, "s.readiness_score >= ?")
		args = append(args, filter.MinReadinessScore)
	}
	if filter.SellerID > 0 {
		conditions = append(conditions, "s.seller_id = ?")
		args = append(args, filter.SellerID)
	}
	if filter.Q != "" {
		like := "%" + filter.Q + "%"
		conditions = append(conditions, "(s.title LIKE ? OR h.name LIKE ? OR s.description LIKE ?)")
		args = append(args, like, like, like)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	switch filter.Sort {
	case "price_asc":
		query += " ORDER BY s.price ASC"
	case "price_desc":
		query += " ORDER BY s.price DESC"
	case "beginner":
		query += " ORDER BY s.beginner_score DESC"
	case "readiness":
		query += " ORDER BY s.readiness_score DESC"
	default:
		query += " ORDER BY s.published_at DESC, s.created_at DESC"
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []domain.StarterSet
	for rows.Next() {
		var s domain.StarterSet
		var hobbyName, categoryName, imageURL string
		var publishedAt sql.NullTime

		err := rows.Scan(
			&s.ID, &s.SellerID, &s.HobbyID, &s.CategoryID,
			&s.Title, &s.Description, &s.Price, &s.Status,
			&s.BeginnerScore, &s.ReadinessScore, &s.ValueScore, &s.EstimatedNewPrice,
			&s.PreviousOwnerNote, &s.StartableSummary, &publishedAt,
			&s.CreatedAt, &s.UpdatedAt,
			&hobbyName, &categoryName, &imageURL,
		)
		if err != nil {
			return nil, err
		}
		if publishedAt.Valid {
			s.PublishedAt = &publishedAt.Time
		}
		s.Hobby = &domain.Hobby{Name: hobbyName}
		s.Category = &domain.HobbyCategory{Name: categoryName}
		if imageURL != "" {
			s.Images = []domain.SetImage{{ImageURL: imageURL}}
		}
		sets = append(sets, s)
	}
	return sets, rows.Err()
}

func (r *setRepository) FindByID(id int64) (*domain.StarterSet, error) {
	s := &domain.StarterSet{}
	var publishedAt sql.NullTime
	var hobbyName, hobbySlug, categoryName, categorySlug string
	var hobbyID, categoryID int64

	err := r.db.QueryRow(`
		SELECT s.id, s.seller_id, s.hobby_id, s.category_id, s.title, s.description, s.price,
		       s.status, s.beginner_score, s.readiness_score, s.value_score, s.estimated_new_price,
		       s.previous_owner_note, s.startable_summary, s.published_at, s.created_at, s.updated_at,
		       COALESCE(h.name,'') AS hobby_name, COALESCE(h.slug,'') AS hobby_slug,
		       COALESCE(hc.name,'') AS category_name, COALESCE(hc.slug,'') AS category_slug,
		       u.id, u.display_name, u.email, u.avatar_url, u.rating_average
		FROM starter_sets s
		LEFT JOIN hobbies h ON s.hobby_id = h.id
		LEFT JOIN hobby_categories hc ON s.category_id = hc.id
		LEFT JOIN users u ON s.seller_id = u.id
		WHERE s.id = ?`, id,
	).Scan(
		&s.ID, &s.SellerID, &hobbyID, &categoryID,
		&s.Title, &s.Description, &s.Price, &s.Status,
		&s.BeginnerScore, &s.ReadinessScore, &s.ValueScore, &s.EstimatedNewPrice,
		&s.PreviousOwnerNote, &s.StartableSummary, &publishedAt,
		&s.CreatedAt, &s.UpdatedAt,
		&hobbyName, &hobbySlug, &categoryName, &categorySlug,
		&s.SellerID, // re-scan seller
		new(string), new(string), new(string), new(float64),
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		// fallback query without joins
		return r.findByIDSimple(id)
	}
	_ = hobbyID
	_ = categoryID
	_ = hobbySlug
	_ = categorySlug
	if publishedAt.Valid {
		s.PublishedAt = &publishedAt.Time
	}

	// Fetch with proper joins
	return r.findByIDFull(id)
}

func (r *setRepository) findByIDSimple(id int64) (*domain.StarterSet, error) {
	s := &domain.StarterSet{}
	var publishedAt sql.NullTime
	err := r.db.QueryRow(`
		SELECT id, seller_id, hobby_id, category_id, title, description, price,
		       status, beginner_score, readiness_score, value_score, estimated_new_price,
		       previous_owner_note, startable_summary, published_at, created_at, updated_at
		FROM starter_sets WHERE id = ?`, id,
	).Scan(
		&s.ID, &s.SellerID, &s.HobbyID, &s.CategoryID,
		&s.Title, &s.Description, &s.Price, &s.Status,
		&s.BeginnerScore, &s.ReadinessScore, &s.ValueScore, &s.EstimatedNewPrice,
		&s.PreviousOwnerNote, &s.StartableSummary, &publishedAt,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if publishedAt.Valid {
		s.PublishedAt = &publishedAt.Time
	}
	return s, nil
}

func (r *setRepository) findByIDFull(id int64) (*domain.StarterSet, error) {
	s, err := r.findByIDSimple(id)
	if err != nil || s == nil {
		return s, err
	}

	// Fetch hobby
	if s.HobbyID > 0 {
		h := &domain.Hobby{}
		_ = r.db.QueryRow(`SELECT id, category_id, name, slug, description, sort_order, created_at FROM hobbies WHERE id = ?`, s.HobbyID).
			Scan(&h.ID, &h.CategoryID, &h.Name, &h.Slug, &h.Description, &h.SortOrder, &h.CreatedAt)
		s.Hobby = h
	}

	// Fetch category
	if s.CategoryID > 0 {
		c := &domain.HobbyCategory{}
		_ = r.db.QueryRow(`SELECT id, name, slug, description, icon_name, sort_order, created_at FROM hobby_categories WHERE id = ?`, s.CategoryID).
			Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.IconName, &c.SortOrder, &c.CreatedAt)
		s.Category = c
	}

	// Fetch seller
	if s.SellerID > 0 {
		u := &domain.User{}
		_ = r.db.QueryRow(`SELECT id, display_name, email, avatar_url, rating_average, rating_count, created_at, updated_at FROM users WHERE id = ?`, s.SellerID).
			Scan(&u.ID, &u.DisplayName, &u.Email, &u.AvatarURL, &u.RatingAverage, &u.RatingCount, &u.CreatedAt, &u.UpdatedAt)
		s.Seller = u
	}

	// Fetch images
	imgRows, err := r.db.Query(`SELECT id, starter_set_id, image_url, sort_order, created_at FROM set_images WHERE starter_set_id = ? ORDER BY sort_order ASC`, id)
	if err == nil {
		defer imgRows.Close()
		for imgRows.Next() {
			var img domain.SetImage
			if err := imgRows.Scan(&img.ID, &img.StarterSetID, &img.ImageURL, &img.SortOrder, &img.CreatedAt); err == nil {
				s.Images = append(s.Images, img)
			}
		}
	}

	// Fetch items
	itemRows, err := r.db.Query(`SELECT id, starter_set_id, name, condition_label, quantity, is_essential, note, created_at, updated_at FROM set_items WHERE starter_set_id = ? ORDER BY id ASC`, id)
	if err == nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var item domain.SetItem
			if err := itemRows.Scan(&item.ID, &item.StarterSetID, &item.Name, &item.ConditionLabel, &item.Quantity, &item.IsEssential, &item.Note, &item.CreatedAt, &item.UpdatedAt); err == nil {
				s.Items = append(s.Items, item)
			}
		}
	}

	// Fetch recommended items
	recRows, err := r.db.Query(`SELECT id, starter_set_id, name, importance, reason, created_at FROM recommended_items WHERE starter_set_id = ? ORDER BY id ASC`, id)
	if err == nil {
		defer recRows.Close()
		for recRows.Next() {
			var rec domain.RecommendedItem
			if err := recRows.Scan(&rec.ID, &rec.StarterSetID, &rec.Name, &rec.Importance, &rec.Reason, &rec.CreatedAt); err == nil {
				s.RecommendedItems = append(s.RecommendedItems, rec)
			}
		}
	}

	return s, nil
}

func (r *setRepository) Create(set *domain.StarterSet) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(`
		INSERT INTO starter_sets (seller_id, hobby_id, category_id, title, description, price, status,
		                          beginner_score, readiness_score, value_score, estimated_new_price,
		                          previous_owner_note, startable_summary, published_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		set.SellerID, set.HobbyID, set.CategoryID, set.Title, set.Description, set.Price, set.Status,
		set.BeginnerScore, set.ReadinessScore, set.ValueScore, set.EstimatedNewPrice,
		set.PreviousOwnerNote, set.StartableSummary, set.PublishedAt, now, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *setRepository) Update(set *domain.StarterSet) error {
	_, err := r.db.Exec(`
		UPDATE starter_sets SET title=?, description=?, price=?, status=?,
		       beginner_score=?, readiness_score=?, value_score=?, estimated_new_price=?,
		       previous_owner_note=?, startable_summary=?, hobby_id=?, category_id=?,
		       published_at=?, updated_at=?
		WHERE id=?`,
		set.Title, set.Description, set.Price, set.Status,
		set.BeginnerScore, set.ReadinessScore, set.ValueScore, set.EstimatedNewPrice,
		set.PreviousOwnerNote, set.StartableSummary, set.HobbyID, set.CategoryID,
		set.PublishedAt, time.Now(), set.ID,
	)
	return err
}

func (r *setRepository) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE starter_sets SET status=?, updated_at=? WHERE id=?`, status, time.Now(), id)
	return err
}

func (r *setRepository) FindBySeller(sellerID int64) ([]domain.StarterSet, error) {
	return r.FindAll(repository.SetFilter{SellerID: sellerID})
}

func (r *setRepository) FindFavorites(userID int64) ([]domain.StarterSet, error) {
	query := `
		SELECT s.id, s.seller_id, s.hobby_id, s.category_id, s.title, s.description, s.price,
		       s.status, s.beginner_score, s.readiness_score, s.value_score, s.estimated_new_price,
		       s.previous_owner_note, s.startable_summary, s.published_at, s.created_at, s.updated_at,
		       COALESCE(h.name,'') AS hobby_name, COALESCE(hc.name,'') AS category_name,
		       COALESCE(si.image_url, '') AS image_url
		FROM favorites f
		JOIN starter_sets s ON f.starter_set_id = s.id
		LEFT JOIN hobbies h ON s.hobby_id = h.id
		LEFT JOIN hobby_categories hc ON s.category_id = hc.id
		LEFT JOIN set_images si ON s.id = si.starter_set_id AND si.sort_order = (
			SELECT MIN(sort_order) FROM set_images WHERE starter_set_id = s.id
		)
		WHERE f.user_id = ?
		ORDER BY f.created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []domain.StarterSet
	for rows.Next() {
		var s domain.StarterSet
		var hobbyName, categoryName, imageURL string
		var publishedAt sql.NullTime

		err := rows.Scan(
			&s.ID, &s.SellerID, &s.HobbyID, &s.CategoryID,
			&s.Title, &s.Description, &s.Price, &s.Status,
			&s.BeginnerScore, &s.ReadinessScore, &s.ValueScore, &s.EstimatedNewPrice,
			&s.PreviousOwnerNote, &s.StartableSummary, &publishedAt,
			&s.CreatedAt, &s.UpdatedAt,
			&hobbyName, &categoryName, &imageURL,
		)
		if err != nil {
			return nil, err
		}
		if publishedAt.Valid {
			s.PublishedAt = &publishedAt.Time
		}
		s.Hobby = &domain.Hobby{Name: hobbyName}
		s.Category = &domain.HobbyCategory{Name: categoryName}
		if imageURL != "" {
			s.Images = []domain.SetImage{{ImageURL: imageURL}}
		}
		sets = append(sets, s)
	}
	return sets, rows.Err()
}

func (r *setRepository) AddImage(image *domain.SetImage) error {
	_, err := r.db.Exec(
		`INSERT INTO set_images (starter_set_id, image_url, sort_order, created_at) VALUES (?, ?, ?, ?)`,
		image.StarterSetID, image.ImageURL, image.SortOrder, time.Now(),
	)
	return err
}

func (r *setRepository) AddItem(item *domain.SetItem) error {
	now := time.Now()
	_, err := r.db.Exec(
		`INSERT INTO set_items (starter_set_id, name, condition_label, quantity, is_essential, note, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		item.StarterSetID, item.Name, item.ConditionLabel, item.Quantity, item.IsEssential, item.Note, now, now,
	)
	return err
}

func (r *setRepository) AddRecommendedItem(item *domain.RecommendedItem) error {
	_, err := r.db.Exec(
		`INSERT INTO recommended_items (starter_set_id, name, importance, reason, created_at) VALUES (?, ?, ?, ?, ?)`,
		item.StarterSetID, item.Name, item.Importance, item.Reason, time.Now(),
	)
	return err
}

func (r *setRepository) DeleteItems(setID int64) error {
	_, err := r.db.Exec(`DELETE FROM set_items WHERE starter_set_id = ?`, setID)
	return err
}

func (r *setRepository) DeleteRecommendedItems(setID int64) error {
	_, err := r.db.Exec(`DELETE FROM recommended_items WHERE starter_set_id = ?`, setID)
	return err
}
