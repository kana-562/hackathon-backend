package domain

// Auth
type SignupRequest struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
type UserResponse struct {
	ID          int64   `json:"id"`
	DisplayName string  `json:"displayName"`
	Email       string  `json:"email"`
	AvatarURL   string  `json:"avatarUrl"`
	RatingAvg   float64 `json:"ratingAverage"`
}

// Home
type HomeResponse struct {
	FeaturedSets []StarterSetCard `json:"featuredSets"`
	NewSets      []StarterSetCard `json:"newSets"`
	Categories   []HobbyCategory  `json:"categories"`
}

// Set card (list view)
type StarterSetCard struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Price          int    `json:"price"`
	BeginnerScore  int    `json:"beginnerScore"`
	ReadinessScore int    `json:"readinessScore"`
	ImageURL       string `json:"imageUrl"`
	HobbyName      string `json:"hobbyName"`
	CategoryName   string `json:"categoryName"`
	Status         string `json:"status"`
	IsFavorite     bool   `json:"isFavorite"`
}

// Set detail
type StarterSetDetail struct {
	ID                int64                `json:"id"`
	Title             string               `json:"title"`
	Price             int                  `json:"price"`
	Description       string               `json:"description"`
	Status            string               `json:"status"`
	BeginnerScore     int                  `json:"beginnerScore"`
	ReadinessScore    int                  `json:"readinessScore"`
	ValueScore        int                  `json:"valueScore"`
	EstimatedNewPrice int                  `json:"estimatedNewPrice"`
	PreviousOwnerNote string               `json:"previousOwnerNote"`
	StartableSummary  string               `json:"startableSummary"`
	Images            []string             `json:"images"`
	Items             []SetItemDTO         `json:"items"`
	RecommendedItems  []RecommendedItemDTO `json:"recommendedItems"`
	Seller            UserResponse         `json:"seller"`
	HobbyName         string               `json:"hobbyName"`
	CategoryName      string               `json:"categoryName"`
	IsFavorite        bool                 `json:"isFavorite"`
}

type SetItemDTO struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	ConditionLabel string `json:"conditionLabel"`
	Quantity       int    `json:"quantity"`
	IsEssential    bool   `json:"isEssential"`
	Note           string `json:"note"`
}

type RecommendedItemDTO struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Importance string `json:"importance"`
	Reason     string `json:"reason"`
}

// Search
type SearchQuery struct {
	Q                 string
	Smart             bool
	CategoryID        int64
	HobbyID           int64
	Sort              string
	MaxPrice          int
	MinBeginnerScore  int
	MinReadinessScore int
}
type SearchResponse struct {
	Sets         []StarterSetCard `json:"sets"`
	SmartMessage string           `json:"smartMessage,omitempty"`
	RelatedChips []string         `json:"relatedChips,omitempty"`
}

// Sell draft
type CreateDraftRequest struct {
	HobbyText string `json:"hobbyText"`
}
type CreateDraftResponse struct {
	DraftSetID     int64       `json:"draftSetId"`
	SessionID      int64       `json:"sessionId"`
	Message        string      `json:"message"`
	SuggestedChips []string    `json:"suggestedChips"`
	Progress       ProgressDTO `json:"progress"`
}
type ProgressDTO struct {
	Current int `json:"current"`
	Total   int `json:"total"`
}
type SellMessageRequest struct {
	Message string `json:"message"`
}
type SellMessageResponse struct {
	Message        string      `json:"message"`
	SuggestedChips []string    `json:"suggestedChips,omitempty"`
	Progress       ProgressDTO `json:"progress"`
	Done           bool        `json:"done"`
}

// Draft detail
type DraftDetail struct {
	ID                int64                `json:"id"`
	Title             string               `json:"title"`
	Description       string               `json:"description"`
	Price             int                  `json:"price"`
	CategoryID        int64                `json:"categoryId"`
	HobbyID           int64                `json:"hobbyId"`
	BeginnerScore     int                  `json:"beginnerScore"`
	ReadinessScore    int                  `json:"readinessScore"`
	StartableSummary  string               `json:"startableSummary"`
	PreviousOwnerNote string               `json:"previousOwnerNote"`
	Items             []SetItemDTO         `json:"items"`
	RecommendedItems  []RecommendedItemDTO `json:"recommendedItems"`
	ImageURL          string               `json:"imageUrl"`
}

// Update draft
type UpdateDraftRequest struct {
	Title             string `json:"title"`
	Description       string `json:"description"`
	Price             int    `json:"price"`
	BeginnerScore     int    `json:"beginnerScore"`
	ReadinessScore    int    `json:"readinessScore"`
	StartableSummary  string `json:"startableSummary"`
	PreviousOwnerNote string `json:"previousOwnerNote"`
}

// Transaction
type CreateTransactionRequest struct {
	StarterSetID int64 `json:"starterSetId"`
}
type CreateTransactionResponse struct {
	TransactionID int64  `json:"transactionId"`
	Status        string `json:"status"`
}
type TransactionDetail struct {
	ID         int64          `json:"id"`
	Status     string         `json:"status"`
	Price      int            `json:"price"`
	StarterSet StarterSetCard `json:"starterSet"`
	StartPlan  *StartPlanDTO  `json:"startPlan,omitempty"`
	CreatedAt  string         `json:"createdAt"`
}
type StartPlanDTO struct {
	ID    int64              `json:"id"`
	Title string             `json:"title"`
	Steps []StartPlanStepDTO `json:"steps"`
}
type StartPlanStepDTO struct {
	DayNo int    `json:"dayNo"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Set question
type SetQuestionRequest struct {
	Message string `json:"message"`
}
type SetQuestionResponse struct {
	Message string `json:"message"`
}

// MyPage
type MyPageResponse struct {
	User           UserResponse `json:"user"`
	SellingCount   int          `json:"sellingCount"`
	PurchasesCount int          `json:"purchasesCount"`
	FavoritesCount int          `json:"favoritesCount"`
}
