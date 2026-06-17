package usecase_test

import (
	"testing"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/testutil"
	"hobby-relay-backend/internal/usecase"
)

func newTxUC() (usecase.TransactionUsecase, *testutil.MockSetRepository) {
	setRepo := testutil.NewMockSetRepositoryPublic()
	txRepo := testutil.NewMockTransactionRepository()
	startPlanRepo := testutil.NewMockStartPlanRepository()
	aiClient := testutil.NewMockAIClientControlled()
	uc := usecase.NewTransactionUsecase(txRepo, setRepo, startPlanRepo, aiClient)
	return uc, setRepo
}

func seedSet(t *testing.T, setRepo *testutil.MockSetRepository, sellerID int64, status string, price int) int64 {
	t.Helper()
	set := &domain.StarterSet{
		SellerID: sellerID,
		Title:    "テストセット",
		Price:    price,
		Status:   status,
	}
	id, err := setRepo.Create(set)
	if err != nil {
		t.Fatalf("seedSet failed: %v", err)
	}
	return id
}

func TestCreateTransaction_OnSaleSet_Succeeds(t *testing.T) {
	uc, setRepo := newTxUC()
	setID := seedSet(t, setRepo, 2, domain.SetStatusOnSale, 8000)

	resp, err := uc.CreateTransaction(1, setID) // buyer=1, seller=2
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}
	if resp.TransactionID == 0 {
		t.Error("expected non-zero transaction ID")
	}
	if resp.Status != domain.TxStatusReserved {
		t.Errorf("expected status %s, got %s", domain.TxStatusReserved, resp.Status)
	}
}

func TestCreateTransaction_ReservedSet_ReturnsError(t *testing.T) {
	uc, setRepo := newTxUC()
	setID := seedSet(t, setRepo, 2, domain.SetStatusReserved, 8000)

	_, err := uc.CreateTransaction(1, setID)
	if err == nil {
		t.Error("expected error for reserved set")
	}
}

func TestCreateTransaction_SoldSet_ReturnsError(t *testing.T) {
	uc, setRepo := newTxUC()
	setID := seedSet(t, setRepo, 2, domain.SetStatusSold, 8000)

	_, err := uc.CreateTransaction(1, setID)
	if err == nil {
		t.Error("expected error for sold set")
	}
}

func TestCreateTransaction_BuyerIsSeller_ReturnsError(t *testing.T) {
	uc, setRepo := newTxUC()
	setID := seedSet(t, setRepo, 1, domain.SetStatusOnSale, 5000) // seller=1

	_, err := uc.CreateTransaction(1, setID) // buyer=1 = seller=1
	if err == nil {
		t.Error("expected error when buyer equals seller")
	}
}

func TestCreateTransaction_DoubleBuy_SecondFails(t *testing.T) {
	uc, setRepo := newTxUC()
	setID := seedSet(t, setRepo, 2, domain.SetStatusOnSale, 8000)

	// First purchase
	_, err := uc.CreateTransaction(1, setID)
	if err != nil {
		t.Fatalf("First CreateTransaction failed: %v", err)
	}

	// Second purchase should fail (set is now reserved)
	_, err = uc.CreateTransaction(3, setID)
	if err == nil {
		t.Error("expected error on double purchase")
	}
}
