package usecase

import (
	"errors"
	"fmt"

	"hobby-relay-backend/internal/ai"
	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type SellUsecase interface {
	CreateDraft(userID int64, hobbyText string) (*domain.CreateDraftResponse, error)
	SendMessage(userID, draftID int64, message string) (*domain.SellMessageResponse, error)
	GetDraft(userID, draftID int64) (*domain.DraftDetail, error)
	UpdateDraft(userID, draftID int64, req domain.UpdateDraftRequest) error
	PublishDraft(userID, draftID int64) error
}

type sellUsecase struct {
	setRepo     repository.SetRepository
	sessionRepo repository.AIChatSessionRepository
	messageRepo repository.AIMessageRepository
	aiClient    ai.Client
}

func NewSellUsecase(
	setRepo repository.SetRepository,
	sessionRepo repository.AIChatSessionRepository,
	messageRepo repository.AIMessageRepository,
	aiClient ai.Client,
) SellUsecase {
	return &sellUsecase{
		setRepo:     setRepo,
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		aiClient:    aiClient,
	}
}

func (u *sellUsecase) CreateDraft(userID int64, hobbyText string) (*domain.CreateDraftResponse, error) {
	if hobbyText == "" {
		return nil, errors.New("hobbyText is required")
	}

	// Create draft set
	draft := &domain.StarterSet{
		SellerID: userID,
		Status:   domain.SetStatusDraft,
		Title:    fmt.Sprintf("%s スターターセット", hobbyText),
	}
	draftID, err := u.setRepo.Create(draft)
	if err != nil {
		return nil, err
	}

	// Create AI session
	session := &domain.AIChatSession{
		UserID:       userID,
		StarterSetID: &draftID,
		SessionType:  domain.SessionTypeListing,
		Status:       "active",
		ProgressStep: 1,
	}
	sessionID, err := u.sessionRepo.Create(session)
	if err != nil {
		return nil, err
	}

	// Start AI listing support
	result, err := u.aiClient.StartListingSupport(hobbyText)
	if err != nil {
		return nil, err
	}

	// Save system + assistant messages
	_, _ = u.messageRepo.Create(&domain.AIMessage{
		SessionID: sessionID,
		Sender:    "system",
		Message:   fmt.Sprintf("趣味: %s", hobbyText),
	})
	_, _ = u.messageRepo.Create(&domain.AIMessage{
		SessionID: sessionID,
		Sender:    "assistant",
		Message:   result.Message,
	})

	return &domain.CreateDraftResponse{
		DraftSetID:     draftID,
		SessionID:      sessionID,
		Message:        result.Message,
		SuggestedChips: result.SuggestedChips,
		Progress:       result.Progress,
	}, nil
}

func (u *sellUsecase) SendMessage(userID, draftID int64, message string) (*domain.SellMessageResponse, error) {
	// Verify ownership
	draft, err := u.setRepo.FindByID(draftID)
	if err != nil {
		return nil, err
	}
	if draft == nil {
		return nil, errors.New("draft not found")
	}
	if draft.SellerID != userID {
		return nil, errors.New("unauthorized")
	}

	// Find session for this draft
	session, err := u.sessionRepo.FindBySetID(draftID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	// Save user message first
	_, err = u.messageRepo.Create(&domain.AIMessage{
		SessionID: session.ID,
		Sender:    "user",
		Message:   message,
	})
	if err != nil {
		return nil, err
	}

	// Get full message history (including the user message just saved)
	history, err := u.messageRepo.FindBySession(session.ID)
	if err != nil {
		return nil, err
	}

	// Convert to AI session messages
	sessionMsgs := make([]ai.SessionMessage, 0, len(history))
	for _, m := range history {
		sessionMsgs = append(sessionMsgs, ai.SessionMessage{
			Sender:  m.Sender,
			Message: m.Message,
		})
	}

	// Get AI response
	result, err := u.aiClient.NextListingStep(sessionMsgs, message)
	if err != nil {
		return nil, err
	}

	// Save assistant response
	_, _ = u.messageRepo.Create(&domain.AIMessage{
		SessionID: session.ID,
		Sender:    "assistant",
		Message:   result.Message,
	})

	// Update session progress
	session.ProgressStep = result.Progress.Current
	if result.Done {
		session.Status = "completed"
	}
	_ = u.sessionRepo.Update(session)

	// If done, update draft with generated content
	if result.Done {
		u.applyAIResultToDraft(draft, draftID, result)
	}

	return &domain.SellMessageResponse{
		Message:        result.Message,
		SuggestedChips: result.SuggestedChips,
		Progress:       result.Progress,
		Done:           result.Done,
	}, nil
}

func (u *sellUsecase) applyAIResultToDraft(draft *domain.StarterSet, draftID int64, result *ai.ListingSupportResult) {
	draft.Title = result.Title
	draft.Description = result.Description
	draft.BeginnerScore = result.BeginnerScore
	draft.ReadinessScore = result.ReadinessScore
	draft.PreviousOwnerNote = result.PreviousOwnerNote
	draft.StartableSummary = result.StartableSummary

	_ = u.setRepo.Update(draft)

	// Delete existing items
	_ = u.setRepo.DeleteItems(draftID)
	_ = u.setRepo.DeleteRecommendedItems(draftID)

	// Insert new items
	for _, item := range result.Items {
		_ = u.setRepo.AddItem(&domain.SetItem{
			StarterSetID:   draftID,
			Name:           item.Name,
			ConditionLabel: item.ConditionLabel,
			Quantity:       1,
			IsEssential:    item.IsEssential,
		})
	}

	for _, rec := range result.RecommendedItems {
		_ = u.setRepo.AddRecommendedItem(&domain.RecommendedItem{
			StarterSetID: draftID,
			Name:         rec.Name,
			Importance:   rec.Importance,
			Reason:       rec.Reason,
		})
	}
}

func (u *sellUsecase) GetDraft(userID, draftID int64) (*domain.DraftDetail, error) {
	draft, err := u.setRepo.FindByID(draftID)
	if err != nil {
		return nil, err
	}
	if draft == nil {
		return nil, errors.New("draft not found")
	}
	if draft.SellerID != userID {
		return nil, errors.New("unauthorized")
	}

	items := make([]domain.SetItemDTO, 0, len(draft.Items))
	for _, item := range draft.Items {
		items = append(items, domain.SetItemDTO{
			ID:             item.ID,
			Name:           item.Name,
			ConditionLabel: item.ConditionLabel,
			Quantity:       item.Quantity,
			IsEssential:    item.IsEssential,
			Note:           item.Note,
		})
	}

	recItems := make([]domain.RecommendedItemDTO, 0, len(draft.RecommendedItems))
	for _, rec := range draft.RecommendedItems {
		recItems = append(recItems, domain.RecommendedItemDTO{
			ID:         rec.ID,
			Name:       rec.Name,
			Importance: rec.Importance,
			Reason:     rec.Reason,
		})
	}

	imageURL := ""
	if len(draft.Images) > 0 {
		imageURL = draft.Images[0].ImageURL
	}

	return &domain.DraftDetail{
		ID:                draft.ID,
		Title:             draft.Title,
		Description:       draft.Description,
		Price:             draft.Price,
		CategoryID:        draft.CategoryID,
		HobbyID:           draft.HobbyID,
		BeginnerScore:     draft.BeginnerScore,
		ReadinessScore:    draft.ReadinessScore,
		StartableSummary:  draft.StartableSummary,
		PreviousOwnerNote: draft.PreviousOwnerNote,
		Items:             items,
		RecommendedItems:  recItems,
		ImageURL:          imageURL,
	}, nil
}

func (u *sellUsecase) UpdateDraft(userID, draftID int64, req domain.UpdateDraftRequest) error {
	draft, err := u.setRepo.FindByID(draftID)
	if err != nil {
		return err
	}
	if draft == nil {
		return errors.New("draft not found")
	}
	if draft.SellerID != userID {
		return errors.New("unauthorized")
	}
	if draft.Status != domain.SetStatusDraft {
		return errors.New("only drafts can be updated this way")
	}

	draft.Title = req.Title
	draft.Description = req.Description
	draft.Price = req.Price
	draft.BeginnerScore = req.BeginnerScore
	draft.ReadinessScore = req.ReadinessScore
	draft.StartableSummary = req.StartableSummary
	draft.PreviousOwnerNote = req.PreviousOwnerNote

	return u.setRepo.Update(draft)
}

func (u *sellUsecase) PublishDraft(userID, draftID int64) error {
	draft, err := u.setRepo.FindByID(draftID)
	if err != nil {
		return err
	}
	if draft == nil {
		return errors.New("draft not found")
	}
	if draft.SellerID != userID {
		return errors.New("unauthorized")
	}
	if draft.Status != domain.SetStatusDraft {
		return errors.New("only drafts can be published")
	}

	// Validation
	if draft.Title == "" {
		return errors.New("title is required")
	}
	if draft.Price <= 0 {
		return errors.New("price must be greater than 0")
	}

	return u.setRepo.UpdateStatus(draftID, domain.SetStatusOnSale)
}
