package repository

import "hobby-relay-backend/internal/domain"

type UserRepository interface {
	Create(user *domain.User) (int64, error)
	FindByEmail(email string) (*domain.User, error)
	FindByID(id int64) (*domain.User, error)
	Update(user *domain.User) error
}

type CategoryRepository interface {
	FindAll() ([]domain.HobbyCategory, error)
	FindByID(id int64) (*domain.HobbyCategory, error)
	FindHobbiesByCategoryID(categoryID int64) ([]domain.Hobby, error)
	FindHobbyByID(id int64) (*domain.Hobby, error)
	FindHobbyByName(name string) (*domain.Hobby, error)
}

type SetFilter struct {
	Q                 string
	CategoryID        int64
	HobbyID           int64
	Status            string
	MaxPrice          int
	MinBeginnerScore  int
	MinReadinessScore int
	Sort              string
	SellerID          int64
	Limit             int
	Offset            int
}

type SetRepository interface {
	FindAll(filter SetFilter) ([]domain.StarterSet, error)
	FindByID(id int64) (*domain.StarterSet, error)
	Create(set *domain.StarterSet) (int64, error)
	Update(set *domain.StarterSet) error
	UpdateStatus(id int64, status string) error
	FindBySeller(sellerID int64) ([]domain.StarterSet, error)
	FindFavorites(userID int64) ([]domain.StarterSet, error)
	AddImage(image *domain.SetImage) error
	AddItem(item *domain.SetItem) error
	AddRecommendedItem(item *domain.RecommendedItem) error
	DeleteItems(setID int64) error
	DeleteRecommendedItems(setID int64) error
}

type FavoriteRepository interface {
	Add(userID, setID int64) error
	Remove(userID, setID int64) error
	Exists(userID, setID int64) (bool, error)
	CountBySet(setID int64) (int, error)
}

type TransactionRepository interface {
	Create(tx *domain.Transaction) (int64, error)
	FindByID(id int64) (*domain.Transaction, error)
	FindByBuyer(buyerID int64) ([]domain.Transaction, error)
	FindBySeller(sellerID int64) ([]domain.Transaction, error)
	UpdateStatus(id int64, status string) error
}

type StartPlanRepository interface {
	Create(plan *domain.StartPlan, steps []domain.StartPlanStep) (int64, error)
	FindByTransaction(transactionID int64) (*domain.StartPlan, error)
}

type AIChatSessionRepository interface {
	Create(session *domain.AIChatSession) (int64, error)
	FindByID(id int64) (*domain.AIChatSession, error)
	FindBySetID(setID int64) (*domain.AIChatSession, error)
	Update(session *domain.AIChatSession) error
}

type AIMessageRepository interface {
	Create(msg *domain.AIMessage) (int64, error)
	FindBySession(sessionID int64) ([]domain.AIMessage, error)
}

type DMRepository interface {
	GetOrCreateRoom(user1ID, user2ID int64, setID *int64) (*domain.DMRoom, error)
	FindRoomByID(roomID int64) (*domain.DMRoom, error)
	FindRoomsByUser(userID int64) ([]domain.DMRoom, error)
	CreateMessage(msg *domain.DMMessage) (int64, error)
	FindMessages(roomID int64) ([]domain.DMMessage, error)
	MarkRead(roomID, readerID int64) error
	CountUnread(roomID, readerID int64) (int, error)
	UpdateLastMessage(roomID int64, body string) error
}
