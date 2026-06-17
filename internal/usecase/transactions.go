package usecase

import (
	"errors"
	"fmt"

	"hobby-relay-backend/internal/ai"
	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type TransactionUsecase interface {
	CreateTransaction(buyerID, setID int64) (*domain.CreateTransactionResponse, error)
	GetTransaction(userID, txID int64) (*domain.TransactionDetail, error)
	UpdateStatus(userID, txID int64, status string) error
	GetStartPlan(userID, txID int64) (*domain.StartPlanDTO, error)
}

type transactionUsecase struct {
	txRepo        repository.TransactionRepository
	setRepo       repository.SetRepository
	startPlanRepo repository.StartPlanRepository
	aiClient      ai.Client
}

func NewTransactionUsecase(
	txRepo repository.TransactionRepository,
	setRepo repository.SetRepository,
	startPlanRepo repository.StartPlanRepository,
	aiClient ai.Client,
) TransactionUsecase {
	return &transactionUsecase{
		txRepo:        txRepo,
		setRepo:       setRepo,
		startPlanRepo: startPlanRepo,
		aiClient:      aiClient,
	}
}

func (u *transactionUsecase) CreateTransaction(buyerID, setID int64) (*domain.CreateTransactionResponse, error) {
	// Pre-validate (the real repo also validates under FOR UPDATE lock)
	set, err := u.setRepo.FindByID(setID)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, errors.New("set not found")
	}
	if set.Status != domain.SetStatusOnSale {
		return nil, fmt.Errorf("set is not available for purchase (status: %s)", set.Status)
	}
	if set.SellerID == buyerID {
		return nil, errors.New("cannot purchase your own set")
	}

	tx := &domain.Transaction{
		StarterSetID: setID,
		BuyerID:      buyerID,
		SellerID:     set.SellerID,
		Price:        set.Price,
		Status:       domain.TxStatusReserved,
	}

	id, err := u.txRepo.Create(tx)
	if err != nil {
		return nil, err
	}

	// Update set status (real repo does this in DB transaction)
	_ = u.setRepo.UpdateStatus(setID, domain.SetStatusReserved)

	return &domain.CreateTransactionResponse{
		TransactionID: id,
		Status:        domain.TxStatusReserved,
	}, nil
}

func (u *transactionUsecase) GetTransaction(userID, txID int64) (*domain.TransactionDetail, error) {
	tx, err := u.txRepo.FindByID(txID)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, errors.New("transaction not found")
	}
	if tx.BuyerID != userID && tx.SellerID != userID {
		return nil, errors.New("unauthorized")
	}

	set, err := u.setRepo.FindByID(tx.StarterSetID)
	if err != nil {
		return nil, err
	}

	var card domain.StarterSetCard
	if set != nil {
		card = toSetCard(*set)
	}

	// Check for existing start plan
	var planDTO *domain.StartPlanDTO
	plan, _ := u.startPlanRepo.FindByTransaction(txID)
	if plan != nil {
		steps := make([]domain.StartPlanStepDTO, 0, len(plan.Steps))
		for _, s := range plan.Steps {
			steps = append(steps, domain.StartPlanStepDTO{
				DayNo: s.DayNo,
				Title: s.Title,
				Body:  s.Body,
			})
		}
		planDTO = &domain.StartPlanDTO{
			ID:    plan.ID,
			Title: plan.Title,
			Steps: steps,
		}
	}

	return &domain.TransactionDetail{
		ID:         tx.ID,
		Status:     tx.Status,
		Price:      tx.Price,
		StarterSet: card,
		StartPlan:  planDTO,
		CreatedAt:  tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (u *transactionUsecase) UpdateStatus(userID, txID int64, status string) error {
	tx, err := u.txRepo.FindByID(txID)
	if err != nil {
		return err
	}
	if tx == nil {
		return errors.New("transaction not found")
	}
	if tx.BuyerID != userID && tx.SellerID != userID {
		return errors.New("unauthorized")
	}

	validStatuses := map[string]bool{
		domain.TxStatusHandoverWaiting: true,
		domain.TxStatusShipped:         true,
		domain.TxStatusReceived:        true,
		domain.TxStatusCompleted:       true,
		domain.TxStatusCancelled:       true,
	}
	if !validStatuses[status] {
		return errors.New("invalid status")
	}

	if err := u.txRepo.UpdateStatus(txID, status); err != nil {
		return err
	}

	// If completed, mark set as sold
	if status == domain.TxStatusCompleted {
		_ = u.setRepo.UpdateStatus(tx.StarterSetID, domain.SetStatusSold)
	}
	// If cancelled, restore set to on_sale
	if status == domain.TxStatusCancelled {
		_ = u.setRepo.UpdateStatus(tx.StarterSetID, domain.SetStatusOnSale)
	}

	return nil
}

func (u *transactionUsecase) GetStartPlan(userID, txID int64) (*domain.StartPlanDTO, error) {
	tx, err := u.txRepo.FindByID(txID)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, errors.New("transaction not found")
	}
	if tx.BuyerID != userID {
		return nil, errors.New("unauthorized: only buyer can get start plan")
	}

	// Check if already exists
	existing, err := u.startPlanRepo.FindByTransaction(txID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		steps := make([]domain.StartPlanStepDTO, 0, len(existing.Steps))
		for _, s := range existing.Steps {
			steps = append(steps, domain.StartPlanStepDTO{
				DayNo: s.DayNo,
				Title: s.Title,
				Body:  s.Body,
			})
		}
		return &domain.StartPlanDTO{
			ID:    existing.ID,
			Title: existing.Title,
			Steps: steps,
		}, nil
	}

	// Generate new plan
	set, err := u.setRepo.FindByID(tx.StarterSetID)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, errors.New("set not found")
	}

	hobbyName := ""
	if set.Hobby != nil {
		hobbyName = set.Hobby.Name
	}

	stepDTOs, err := u.aiClient.GenerateStartPlan(set.Title, hobbyName)
	if err != nil {
		return nil, err
	}

	title := hobbyName + " 7日間スタートプラン"
	if hobbyName == "" {
		title = set.Title + " スタートプラン"
	}

	plan := &domain.StartPlan{
		TransactionID: txID,
		StarterSetID:  tx.StarterSetID,
		UserID:        userID,
		Title:         title,
	}

	steps := make([]domain.StartPlanStep, 0, len(stepDTOs))
	for _, dto := range stepDTOs {
		steps = append(steps, domain.StartPlanStep{
			DayNo: dto.DayNo,
			Title: dto.Title,
			Body:  dto.Body,
		})
	}

	planID, err := u.startPlanRepo.Create(plan, steps)
	if err != nil {
		return nil, err
	}

	return &domain.StartPlanDTO{
		ID:    planID,
		Title: plan.Title,
		Steps: stepDTOs,
	}, nil
}
