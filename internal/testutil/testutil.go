package testutil

import (
	"errors"
	"fmt"
	"sync"

	"hobby-relay-backend/internal/ai"
	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

// ---- Mock User Repository ----

type MockUserRepository struct {
	mu      sync.RWMutex
	users   map[int64]*domain.User
	byEmail map[string]*domain.User
	nextID  int64
}

func NewMockUserRepository() repository.UserRepository {
	return &MockUserRepository{
		users:   make(map[int64]*domain.User),
		byEmail: make(map[string]*domain.User),
		nextID:  1,
	}
}

func (r *MockUserRepository) Create(user *domain.User) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	user.ID = id
	cp := *user
	r.users[id] = &cp
	cp2 := *user
	r.byEmail[user.Email] = &cp2
	return id, nil
}

func (r *MockUserRepository) FindByEmail(email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.byEmail[email]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (r *MockUserRepository) FindByID(id int64) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, nil
	}
	cp := *u
	return &cp, nil
}

func (r *MockUserRepository) Update(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	cp := *user
	r.users[user.ID] = &cp
	cp2 := *user
	r.byEmail[user.Email] = &cp2
	return nil
}

// ---- Mock Set Repository ----

type MockSetRepository struct {
	mu     sync.RWMutex
	sets   map[int64]*domain.StarterSet
	nextID int64
}

func NewMockSetRepository() repository.SetRepository {
	return &MockSetRepository{
		sets:   make(map[int64]*domain.StarterSet),
		nextID: 1,
	}
}

// NewMockSetRepositoryPublic returns the concrete type for test seeding
func NewMockSetRepositoryPublic() *MockSetRepository {
	return &MockSetRepository{
		sets:   make(map[int64]*domain.StarterSet),
		nextID: 1,
	}
}

// Create on *MockSetRepository returns (int64, error) for test seeding convenience
func (r *MockSetRepository) CreateReturningID(set *domain.StarterSet) (int64, error) {
	return r.Create(set)
}

func (r *MockSetRepository) FindAll(filter repository.SetFilter) ([]domain.StarterSet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.StarterSet
	for _, s := range r.sets {
		if filter.Status != "" && s.Status != filter.Status {
			continue
		}
		if filter.SellerID > 0 && s.SellerID != filter.SellerID {
			continue
		}
		if filter.CategoryID > 0 && s.CategoryID != filter.CategoryID {
			continue
		}
		if filter.HobbyID > 0 && s.HobbyID != filter.HobbyID {
			continue
		}
		if filter.MaxPrice > 0 && s.Price > filter.MaxPrice {
			continue
		}
		if filter.MinBeginnerScore > 0 && s.BeginnerScore < filter.MinBeginnerScore {
			continue
		}
		if filter.MinReadinessScore > 0 && s.ReadinessScore < filter.MinReadinessScore {
			continue
		}
		cp := *s
		result = append(result, cp)
	}
	return result, nil
}

func (r *MockSetRepository) FindByID(id int64) (*domain.StarterSet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sets[id]
	if !ok {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}

func (r *MockSetRepository) Create(set *domain.StarterSet) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	set.ID = id
	cp := *set
	r.sets[id] = &cp
	return id, nil
}

func (r *MockSetRepository) Update(set *domain.StarterSet) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.sets[set.ID]; !ok {
		return errors.New("set not found")
	}
	cp := *set
	r.sets[set.ID] = &cp
	return nil
}

func (r *MockSetRepository) UpdateStatus(id int64, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sets[id]
	if !ok {
		return errors.New("set not found")
	}
	s.Status = status
	return nil
}

func (r *MockSetRepository) FindBySeller(sellerID int64) ([]domain.StarterSet, error) {
	return r.FindAll(repository.SetFilter{SellerID: sellerID})
}

func (r *MockSetRepository) FindFavorites(_ int64) ([]domain.StarterSet, error) {
	return nil, nil
}

func (r *MockSetRepository) AddImage(_ *domain.SetImage) error            { return nil }
func (r *MockSetRepository) AddItem(_ *domain.SetItem) error              { return nil }
func (r *MockSetRepository) AddRecommendedItem(_ *domain.RecommendedItem) error { return nil }
func (r *MockSetRepository) DeleteItems(_ int64) error                    { return nil }
func (r *MockSetRepository) DeleteRecommendedItems(_ int64) error         { return nil }

// ---- Mock Favorite Repository ----

type MockFavoriteRepository struct {
	mu        sync.RWMutex
	favorites map[string]bool
}

func NewMockFavoriteRepository() repository.FavoriteRepository {
	return &MockFavoriteRepository{
		favorites: make(map[string]bool),
	}
}

func favKey(userID, setID int64) string {
	return fmt.Sprintf("%d:%d", userID, setID)
}

func (r *MockFavoriteRepository) Add(userID, setID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.favorites[favKey(userID, setID)] = true
	return nil
}

func (r *MockFavoriteRepository) Remove(userID, setID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.favorites, favKey(userID, setID))
	return nil
}

func (r *MockFavoriteRepository) Exists(userID, setID int64) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.favorites[favKey(userID, setID)], nil
}

func (r *MockFavoriteRepository) CountBySet(_ int64) (int, error) {
	return 0, nil
}

// ---- Mock Transaction Repository ----

type MockTransactionRepository struct {
	mu     sync.RWMutex
	txs    map[int64]*domain.Transaction
	nextID int64
}

func NewMockTransactionRepository() repository.TransactionRepository {
	return &MockTransactionRepository{
		txs:    make(map[int64]*domain.Transaction),
		nextID: 1,
	}
}

func (r *MockTransactionRepository) Create(tx *domain.Transaction) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	tx.ID = id
	cp := *tx
	r.txs[id] = &cp
	return id, nil
}

func (r *MockTransactionRepository) FindByID(id int64) (*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, ok := r.txs[id]
	if !ok {
		return nil, nil
	}
	cp := *tx
	return &cp, nil
}

func (r *MockTransactionRepository) FindByBuyer(buyerID int64) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Transaction
	for _, tx := range r.txs {
		if tx.BuyerID == buyerID {
			result = append(result, *tx)
		}
	}
	return result, nil
}

func (r *MockTransactionRepository) FindBySeller(sellerID int64) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Transaction
	for _, tx := range r.txs {
		if tx.SellerID == sellerID {
			result = append(result, *tx)
		}
	}
	return result, nil
}

func (r *MockTransactionRepository) UpdateStatus(id int64, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, ok := r.txs[id]
	if !ok {
		return errors.New("transaction not found")
	}
	tx.Status = status
	return nil
}

// ---- Mock StartPlan Repository ----

type MockStartPlanRepository struct {
	mu     sync.RWMutex
	plans  map[int64]*domain.StartPlan
	byTx   map[int64]*domain.StartPlan
	nextID int64
}

func NewMockStartPlanRepository() repository.StartPlanRepository {
	return &MockStartPlanRepository{
		plans:  make(map[int64]*domain.StartPlan),
		byTx:   make(map[int64]*domain.StartPlan),
		nextID: 1,
	}
}

func (r *MockStartPlanRepository) Create(plan *domain.StartPlan, steps []domain.StartPlanStep) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	plan.ID = id
	plan.Steps = steps
	cp := *plan
	r.plans[id] = &cp
	r.byTx[plan.TransactionID] = &cp
	return id, nil
}

func (r *MockStartPlanRepository) FindByTransaction(transactionID int64) (*domain.StartPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byTx[transactionID]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

// ---- Mock AI Chat Session Repository ----

type MockAIChatSessionRepository struct {
	mu       sync.RWMutex
	sessions map[int64]*domain.AIChatSession
	nextID   int64
}

func NewMockAIChatSessionRepository() repository.AIChatSessionRepository {
	return &MockAIChatSessionRepository{
		sessions: make(map[int64]*domain.AIChatSession),
		nextID:   1,
	}
}

func (r *MockAIChatSessionRepository) Create(session *domain.AIChatSession) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	session.ID = id
	cp := *session
	r.sessions[id] = &cp
	return id, nil
}

func (r *MockAIChatSessionRepository) FindByID(id int64) (*domain.AIChatSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, nil
	}
	cp := *s
	return &cp, nil
}

func (r *MockAIChatSessionRepository) FindBySetID(setID int64) (*domain.AIChatSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.sessions {
		if s.StarterSetID != nil && *s.StarterSetID == setID {
			cp := *s
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *MockAIChatSessionRepository) Update(session *domain.AIChatSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.sessions[session.ID]; !ok {
		return errors.New("session not found")
	}
	cp := *session
	r.sessions[session.ID] = &cp
	return nil
}

// ---- Mock AI Message Repository ----

type MockAIMessageRepository struct {
	mu       sync.RWMutex
	messages map[int64][]domain.AIMessage
	nextID   int64
}

func NewMockAIMessageRepository() repository.AIMessageRepository {
	return &MockAIMessageRepository{
		messages: make(map[int64][]domain.AIMessage),
		nextID:   1,
	}
}

func (r *MockAIMessageRepository) Create(msg *domain.AIMessage) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := r.nextID
	r.nextID++
	msg.ID = id
	cp := *msg
	r.messages[msg.SessionID] = append(r.messages[msg.SessionID], cp)
	return id, nil
}

func (r *MockAIMessageRepository) FindBySession(sessionID int64) ([]domain.AIMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	msgs := r.messages[sessionID]
	result := make([]domain.AIMessage, len(msgs))
	for i, m := range msgs {
		result[i] = m
	}
	return result, nil
}

// ---- Controlled Mock AI Client (for tests that need predictable behavior) ----

type MockAIClientControlled struct{}

func NewMockAIClientControlled() ai.Client {
	return &MockAIClientControlled{}
}

func (m *MockAIClientControlled) StartListingSupport(hobbyText string) (*ai.ListingSupportResult, error) {
	return &ai.ListingSupportResult{
		Message:        "何のアイテムが含まれていますか？",
		SuggestedChips: []string{"ギター", "チューナー", "ピック"},
		Progress:       domain.ProgressDTO{Current: 1, Total: 5},
		Done:           false,
	}, nil
}

func (m *MockAIClientControlled) NextListingStep(sessionMessages []ai.SessionMessage, userMessage string) (*ai.ListingSupportResult, error) {
	userMsgCount := 0
	for _, msg := range sessionMessages {
		if msg.Sender == "user" {
			userMsgCount++
		}
	}
	step := userMsgCount + 1

	if step >= 5 {
		return &ai.ListingSupportResult{
			Message:           "出品準備完了です！",
			Progress:          domain.ProgressDTO{Current: 5, Total: 5},
			Done:              true,
			Title:             "ギター スターターセット一式",
			Description:       "ギターを始めるためのセットです。",
			Items:             []ai.ItemInput{{Name: "ギター本体", ConditionLabel: "good", IsEssential: true}},
			RecommendedItems:  []ai.RecommendedInput{{Name: "替え弦", Importance: "required", Reason: "消耗品"}},
			BeginnerScore:     4,
			ReadinessScore:    85,
			PreviousOwnerNote: userMessage,
			StartableSummary:  "このセットだけで始められます。",
		}, nil
	}

	return &ai.ListingSupportResult{
		Message:        "次の質問です。",
		SuggestedChips: []string{"選択肢1", "選択肢2"},
		Progress:       domain.ProgressDTO{Current: step, Total: 5},
		Done:           false,
	}, nil
}

func (m *MockAIClientControlled) AnswerSetQuestion(setTitle string, _ []string, _ string) (string, error) {
	return "こちらは" + setTitle + "に関する回答です。", nil
}

func (m *MockAIClientControlled) InterpretSearchQuery(query string) (*ai.SearchInterpretation, error) {
	result := &ai.SearchInterpretation{}
	switch query {
	case "1万円以内", "10000円以内":
		result.MaxPrice = 10000
		result.SmartMessage = "「1万円以内」のセットを表示しています。"
	case "家でできる":
		result.RelatedHobbies = []string{"コーヒー", "イラスト", "ヨガ", "ウクレレ"}
		result.SmartMessage = "「家でできる」趣味のセットを表示しています。"
	default:
		result.SmartMessage = "「" + query + "」に近いセットを表示しています。"
	}
	return result, nil
}

func (m *MockAIClientControlled) GenerateStartPlan(_ string, _ string) ([]domain.StartPlanStepDTO, error) {
	return []domain.StartPlanStepDTO{
		{DayNo: 1, Title: "Day 1", Body: "Begin"},
		{DayNo: 2, Title: "Day 2", Body: "Practice"},
		{DayNo: 7, Title: "Day 7", Body: "Review"},
	}, nil
}
