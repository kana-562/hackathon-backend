package mysql

import (
	"database/sql"
	"time"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type dmRepository struct {
	db *sql.DB
}

func NewDMRepository(db *sql.DB) repository.DMRepository {
	return &dmRepository{db: db}
}

func (r *dmRepository) GetOrCreateRoom(user1ID, user2ID int64, setID *int64) (*domain.DMRoom, error) {
	// Canonical order: smaller ID first
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	room := &domain.DMRoom{}
	err := r.db.QueryRow(`
		SELECT id, user1_id, user2_id, set_id, COALESCE(last_message,''), last_message_at, created_at, updated_at
		FROM dm_rooms WHERE user1_id=? AND user2_id=?`, user1ID, user2ID,
	).Scan(&room.ID, &room.User1ID, &room.User2ID, &room.SetID,
		&room.LastMessage, &room.LastMessageAt, &room.CreatedAt, &room.UpdatedAt)

	if err == sql.ErrNoRows {
		now := time.Now()
		result, err := r.db.Exec(`
			INSERT INTO dm_rooms (user1_id, user2_id, set_id, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)`, user1ID, user2ID, setID, now, now)
		if err != nil {
			return nil, err
		}
		id, _ := result.LastInsertId()
		room = &domain.DMRoom{
			ID:        id,
			User1ID:   user1ID,
			User2ID:   user2ID,
			SetID:     setID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		return room, nil
	}
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (r *dmRepository) FindRoomByID(roomID int64) (*domain.DMRoom, error) {
	room := &domain.DMRoom{}
	err := r.db.QueryRow(`
		SELECT id, user1_id, user2_id, set_id, COALESCE(last_message,''), last_message_at, created_at, updated_at
		FROM dm_rooms WHERE id=?`, roomID,
	).Scan(&room.ID, &room.User1ID, &room.User2ID, &room.SetID,
		&room.LastMessage, &room.LastMessageAt, &room.CreatedAt, &room.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return room, err
}

func (r *dmRepository) FindRoomsByUser(userID int64) ([]domain.DMRoom, error) {
	rows, err := r.db.Query(`
		SELECT id, user1_id, user2_id, set_id, COALESCE(last_message,''), last_message_at, created_at, updated_at
		FROM dm_rooms
		WHERE user1_id=? OR user2_id=?
		ORDER BY COALESCE(last_message_at, created_at) DESC`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []domain.DMRoom
	for rows.Next() {
		var room domain.DMRoom
		if err := rows.Scan(&room.ID, &room.User1ID, &room.User2ID, &room.SetID,
			&room.LastMessage, &room.LastMessageAt, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (r *dmRepository) CreateMessage(msg *domain.DMMessage) (int64, error) {
	now := time.Now()
	result, err := r.db.Exec(`
		INSERT INTO dm_messages (room_id, sender_id, body, is_read, created_at)
		VALUES (?, ?, ?, FALSE, ?)`, msg.RoomID, msg.SenderID, msg.Body, now)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *dmRepository) FindMessages(roomID int64) ([]domain.DMMessage, error) {
	rows, err := r.db.Query(`
		SELECT id, room_id, sender_id, body, is_read, created_at
		FROM dm_messages WHERE room_id=? ORDER BY created_at ASC`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []domain.DMMessage
	for rows.Next() {
		var m domain.DMMessage
		if err := rows.Scan(&m.ID, &m.RoomID, &m.SenderID, &m.Body, &m.IsRead, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

func (r *dmRepository) MarkRead(roomID, readerID int64) error {
	_, err := r.db.Exec(`
		UPDATE dm_messages SET is_read=TRUE
		WHERE room_id=? AND sender_id!=? AND is_read=FALSE`, roomID, readerID)
	return err
}

func (r *dmRepository) CountUnread(roomID, readerID int64) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM dm_messages
		WHERE room_id=? AND sender_id!=? AND is_read=FALSE`, roomID, readerID).Scan(&count)
	return count, err
}

func (r *dmRepository) UpdateLastMessage(roomID int64, body string) error {
	_, err := r.db.Exec(`
		UPDATE dm_rooms SET last_message=?, last_message_at=?, updated_at=?
		WHERE id=?`, body, time.Now(), time.Now(), roomID)
	return err
}
