package mysql

import (
	"database/sql"
	"time"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type aiChatSessionRepository struct {
	db *sql.DB
}

func NewAIChatSessionRepository(db *sql.DB) repository.AIChatSessionRepository {
	return &aiChatSessionRepository{db: db}
}

func (r *aiChatSessionRepository) Create(session *domain.AIChatSession) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(`
		INSERT INTO ai_chat_sessions (user_id, starter_set_id, session_type, status, progress_step, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		session.UserID, session.StarterSetID, session.SessionType, session.Status, session.ProgressStep, now, now,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *aiChatSessionRepository) FindByID(id int64) (*domain.AIChatSession, error) {
	s := &domain.AIChatSession{}
	var setID sql.NullInt64
	err := r.db.QueryRow(`
		SELECT id, user_id, starter_set_id, session_type, status, progress_step, created_at, updated_at
		FROM ai_chat_sessions WHERE id = ?`, id,
	).Scan(&s.ID, &s.UserID, &setID, &s.SessionType, &s.Status, &s.ProgressStep, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if setID.Valid {
		s.StarterSetID = &setID.Int64
	}
	return s, nil
}

func (r *aiChatSessionRepository) FindBySetID(setID int64) (*domain.AIChatSession, error) {
	s := &domain.AIChatSession{}
	var sid sql.NullInt64
	err := r.db.QueryRow(`
		SELECT id, user_id, starter_set_id, session_type, status, progress_step, created_at, updated_at
		FROM ai_chat_sessions WHERE starter_set_id = ? ORDER BY created_at DESC LIMIT 1`, setID,
	).Scan(&s.ID, &s.UserID, &sid, &s.SessionType, &s.Status, &s.ProgressStep, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if sid.Valid {
		s.StarterSetID = &sid.Int64
	}
	return s, nil
}

func (r *aiChatSessionRepository) Update(session *domain.AIChatSession) error {
	_, err := r.db.Exec(`
		UPDATE ai_chat_sessions SET status=?, progress_step=?, updated_at=? WHERE id=?`,
		session.Status, session.ProgressStep, time.Now(), session.ID,
	)
	return err
}

type aiMessageRepository struct {
	db *sql.DB
}

func NewAIMessageRepository(db *sql.DB) repository.AIMessageRepository {
	return &aiMessageRepository{db: db}
}

func (r *aiMessageRepository) Create(msg *domain.AIMessage) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO ai_messages (session_id, sender, message, metadata, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		msg.SessionID, msg.Sender, msg.Message, msg.Metadata, time.Now(),
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *aiMessageRepository) FindBySession(sessionID int64) ([]domain.AIMessage, error) {
	rows, err := r.db.Query(`
		SELECT id, session_id, sender, message, metadata, created_at
		FROM ai_messages WHERE session_id = ? ORDER BY created_at ASC`, sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.AIMessage
	for rows.Next() {
		var m domain.AIMessage
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Sender, &m.Message, &m.Metadata, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}
