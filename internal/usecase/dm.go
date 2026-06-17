package usecase

import (
	"fmt"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type DMUsecase interface {
	GetOrCreateRoom(userID, partnerID int64, setID *int64) (*domain.GetOrCreateRoomResponse, error)
	ListRooms(userID int64) ([]domain.DMRoomResponse, error)
	GetMessages(userID, roomID int64) ([]domain.DMMessageResponse, error)
	SendMessage(userID, roomID int64, body string) (*domain.DMMessageResponse, error)
	MarkRead(userID, roomID int64) error
}

type dmUsecase struct {
	dmRepo   repository.DMRepository
	userRepo repository.UserRepository
	setRepo  repository.SetRepository
}

func NewDMUsecase(dmRepo repository.DMRepository, userRepo repository.UserRepository, setRepo repository.SetRepository) DMUsecase {
	return &dmUsecase{dmRepo: dmRepo, userRepo: userRepo, setRepo: setRepo}
}

func (u *dmUsecase) GetOrCreateRoom(userID, partnerID int64, setID *int64) (*domain.GetOrCreateRoomResponse, error) {
	if userID == partnerID {
		return nil, fmt.Errorf("cannot DM yourself")
	}
	room, err := u.dmRepo.GetOrCreateRoom(userID, partnerID, setID)
	if err != nil {
		return nil, err
	}
	return &domain.GetOrCreateRoomResponse{RoomID: room.ID}, nil
}

func (u *dmUsecase) ListRooms(userID int64) ([]domain.DMRoomResponse, error) {
	rooms, err := u.dmRepo.FindRoomsByUser(userID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.DMRoomResponse, 0, len(rooms))
	for _, room := range rooms {
		partnerID := room.User2ID
		if room.User2ID == userID {
			partnerID = room.User1ID
		}

		partner, err := u.userRepo.FindByID(partnerID)
		if err != nil || partner == nil {
			continue
		}

		unread, _ := u.dmRepo.CountUnread(room.ID, userID)

		var lastMsgAt *string
		if room.LastMessageAt != nil {
			s := room.LastMessageAt.Format("2006-01-02T15:04:05Z07:00")
			lastMsgAt = &s
		}

		resp := domain.DMRoomResponse{
			ID:            room.ID,
			PartnerID:     partnerID,
			PartnerName:   partner.DisplayName,
			PartnerAvatar: partner.AvatarURL,
			SetID:         room.SetID,
			LastMessage:   room.LastMessage,
			LastMessageAt: lastMsgAt,
			UnreadCount:   unread,
		}

		if room.SetID != nil {
			set, err := u.setRepo.FindByID(*room.SetID)
			if err == nil && set != nil {
				resp.SetTitle = set.Title
			}
		}

		result = append(result, resp)
	}
	return result, nil
}

func (u *dmUsecase) GetMessages(userID, roomID int64) ([]domain.DMMessageResponse, error) {
	room, err := u.dmRepo.FindRoomByID(roomID)
	if err != nil || room == nil {
		return nil, fmt.Errorf("room not found")
	}
	if room.User1ID != userID && room.User2ID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	msgs, err := u.dmRepo.FindMessages(roomID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.DMMessageResponse, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, domain.DMMessageResponse{
			ID:        m.ID,
			RoomID:    m.RoomID,
			SenderID:  m.SenderID,
			Body:      m.Body,
			IsRead:    m.IsRead,
			CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return result, nil
}

func (u *dmUsecase) SendMessage(userID, roomID int64, body string) (*domain.DMMessageResponse, error) {
	if body == "" {
		return nil, fmt.Errorf("body is required")
	}

	room, err := u.dmRepo.FindRoomByID(roomID)
	if err != nil || room == nil {
		return nil, fmt.Errorf("room not found")
	}
	if room.User1ID != userID && room.User2ID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	msg := &domain.DMMessage{
		RoomID:   roomID,
		SenderID: userID,
		Body:     body,
	}
	id, err := u.dmRepo.CreateMessage(msg)
	if err != nil {
		return nil, err
	}

	_ = u.dmRepo.UpdateLastMessage(roomID, body)

	return &domain.DMMessageResponse{
		ID:       id,
		RoomID:   roomID,
		SenderID: userID,
		Body:     body,
		IsRead:   false,
	}, nil
}

func (u *dmUsecase) MarkRead(userID, roomID int64) error {
	room, err := u.dmRepo.FindRoomByID(roomID)
	if err != nil || room == nil {
		return fmt.Errorf("room not found")
	}
	if room.User1ID != userID && room.User2ID != userID {
		return fmt.Errorf("unauthorized")
	}
	return u.dmRepo.MarkRead(roomID, userID)
}
