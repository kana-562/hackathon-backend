package domain

import "time"

type User struct {
	ID            int64     `json:"id"`
	DisplayName   string    `json:"displayName"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	AvatarURL     string    `json:"avatarUrl"`
	RatingAverage float64   `json:"ratingAverage"`
	RatingCount   int       `json:"ratingCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type HobbyCategory struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	IconName    string    `json:"iconName"`
	SortOrder   int       `json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Hobby struct {
	ID          int64     `json:"id"`
	CategoryID  int64     `json:"categoryId"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
}

type StarterSet struct {
	ID                int64          `json:"id"`
	SellerID          int64          `json:"sellerId"`
	HobbyID           int64          `json:"hobbyId"`
	CategoryID        int64          `json:"categoryId"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	Price             int            `json:"price"`
	Status            string         `json:"status"`
	BeginnerScore     int            `json:"beginnerScore"`
	ReadinessScore    int            `json:"readinessScore"`
	ValueScore        int            `json:"valueScore"`
	EstimatedNewPrice int            `json:"estimatedNewPrice"`
	PreviousOwnerNote string         `json:"previousOwnerNote"`
	StartableSummary  string         `json:"startableSummary"`
	PublishedAt       *time.Time     `json:"publishedAt"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	Images            []SetImage     `json:"images,omitempty"`
	Items             []SetItem      `json:"items,omitempty"`
	RecommendedItems  []RecommendedItem `json:"recommendedItems,omitempty"`
	Seller            *User          `json:"seller,omitempty"`
	Hobby             *Hobby         `json:"hobby,omitempty"`
	Category          *HobbyCategory `json:"category,omitempty"`
}

type SetImage struct {
	ID           int64     `json:"id"`
	StarterSetID int64     `json:"starterSetId"`
	ImageURL     string    `json:"imageUrl"`
	SortOrder    int       `json:"sortOrder"`
	CreatedAt    time.Time `json:"createdAt"`
}

type SetItem struct {
	ID             int64     `json:"id"`
	StarterSetID   int64     `json:"starterSetId"`
	Name           string    `json:"name"`
	ConditionLabel string    `json:"conditionLabel"`
	Quantity       int       `json:"quantity"`
	IsEssential    bool      `json:"isEssential"`
	Note           string    `json:"note"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type RecommendedItem struct {
	ID           int64     `json:"id"`
	StarterSetID int64     `json:"starterSetId"`
	Name         string    `json:"name"`
	Importance   string    `json:"importance"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Favorite struct {
	UserID       int64     `json:"userId"`
	StarterSetID int64     `json:"starterSetId"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Transaction struct {
	ID           int64       `json:"id"`
	StarterSetID int64       `json:"starterSetId"`
	BuyerID      int64       `json:"buyerId"`
	SellerID     int64       `json:"sellerId"`
	Price        int         `json:"price"`
	Status       string      `json:"status"`
	CreatedAt    time.Time   `json:"createdAt"`
	UpdatedAt    time.Time   `json:"updatedAt"`
	StarterSet   *StarterSet `json:"starterSet,omitempty"`
}

type AIChatSession struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"userId"`
	StarterSetID *int64     `json:"starterSetId"`
	SessionType  string     `json:"sessionType"`
	Status       string     `json:"status"`
	ProgressStep int        `json:"progressStep"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type AIMessage struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"sessionId"`
	Sender    string    `json:"sender"`
	Message   string    `json:"message"`
	Metadata  string    `json:"metadata"`
	CreatedAt time.Time `json:"createdAt"`
}

type StartPlan struct {
	ID            int64           `json:"id"`
	TransactionID int64           `json:"transactionId"`
	StarterSetID  int64           `json:"starterSetId"`
	UserID        int64           `json:"userId"`
	Title         string          `json:"title"`
	CreatedAt     time.Time       `json:"createdAt"`
	Steps         []StartPlanStep `json:"steps,omitempty"`
}

type StartPlanStep struct {
	ID          int64  `json:"id"`
	StartPlanID int64  `json:"startPlanId"`
	DayNo       int    `json:"dayNo"`
	Title       string `json:"title"`
	Body        string `json:"body"`
}

type DMRoom struct {
	ID            int64      `json:"id"`
	User1ID       int64      `json:"user1Id"`
	User2ID       int64      `json:"user2Id"`
	SetID         *int64     `json:"setId"`
	LastMessage   string     `json:"lastMessage"`
	LastMessageAt *time.Time `json:"lastMessageAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type DMMessage struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"roomId"`
	SenderID  int64     `json:"senderId"`
	Body      string    `json:"body"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}
