package usecase_test

import (
	"testing"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/testutil"
	"hobby-relay-backend/internal/usecase"
)

func newSellUC() usecase.SellUsecase {
	setRepo := testutil.NewMockSetRepository()
	sessionRepo := testutil.NewMockAIChatSessionRepository()
	messageRepo := testutil.NewMockAIMessageRepository()
	aiClient := testutil.NewMockAIClientControlled()
	return usecase.NewSellUsecase(setRepo, sessionRepo, messageRepo, aiClient)
}

func TestCreateDraft_WithHobbyText(t *testing.T) {
	uc := newSellUC()
	resp, err := uc.CreateDraft(1, "ギター")
	if err != nil {
		t.Fatalf("CreateDraft failed: %v", err)
	}
	if resp.DraftSetID == 0 {
		t.Error("expected non-zero draftSetId")
	}
	if resp.SessionID == 0 {
		t.Error("expected non-zero sessionId")
	}
	if resp.Message == "" {
		t.Error("expected non-empty message")
	}
	if resp.Progress.Total != 5 {
		t.Errorf("expected progress total 5, got %d", resp.Progress.Total)
	}
}

func TestCreateDraft_EmptyHobbyText(t *testing.T) {
	uc := newSellUC()
	_, err := uc.CreateDraft(1, "")
	if err == nil {
		t.Error("expected error for empty hobbyText")
	}
}

func TestCreateDraft_StatusIsDraft(t *testing.T) {
	uc := newSellUC()
	resp, err := uc.CreateDraft(1, "ギター")
	if err != nil {
		t.Fatalf("CreateDraft failed: %v", err)
	}

	draft, err := uc.GetDraft(1, resp.DraftSetID)
	if err != nil {
		t.Fatalf("GetDraft failed: %v", err)
	}
	if draft.ID != resp.DraftSetID {
		t.Errorf("expected draft ID %d, got %d", resp.DraftSetID, draft.ID)
	}
}

func TestSendMessage_AdvancesProgress(t *testing.T) {
	uc := newSellUC()
	resp, err := uc.CreateDraft(1, "ギター")
	if err != nil {
		t.Fatalf("CreateDraft failed: %v", err)
	}

	msgResp, err := uc.SendMessage(1, resp.DraftSetID, "ギター本体、チューナー、ピック")
	if err != nil {
		t.Fatalf("SendMessage failed: %v", err)
	}
	if msgResp.Progress.Current <= 1 {
		t.Errorf("expected progress > 1, got %d", msgResp.Progress.Current)
	}
	if msgResp.Done {
		t.Error("expected Done=false after first message")
	}
}

func TestSendMessage_After5Steps_DoneIsTrue(t *testing.T) {
	uc := newSellUC()
	resp, err := uc.CreateDraft(1, "ギター")
	if err != nil {
		t.Fatalf("CreateDraft failed: %v", err)
	}

	messages := []string{
		"ギター本体、チューナー、ピック",
		"ほぼ新品",
		"8000円",
		"大切に使っていました",
	}
	for _, msg := range messages {
		_, err := uc.SendMessage(1, resp.DraftSetID, msg)
		if err != nil {
			t.Fatalf("SendMessage failed: %v", err)
		}
	}

	// 5th message
	finalResp, err := uc.SendMessage(1, resp.DraftSetID, "よろしくお願いします")
	if err != nil {
		t.Fatalf("SendMessage step 5 failed: %v", err)
	}
	if !finalResp.Done {
		t.Error("expected Done=true after step 5")
	}
	if finalResp.Progress.Current != 5 {
		t.Errorf("expected progress 5, got %d", finalResp.Progress.Current)
	}
}

func TestPublishDraft_ValidDraft_Succeeds(t *testing.T) {
	uc := newSellUC()
	resp, _ := uc.CreateDraft(1, "ギター")

	// Update draft with required fields
	err := uc.UpdateDraft(1, resp.DraftSetID, domain.UpdateDraftRequest{
		Title: "ギター スターターセット一式",
		Price: 8000,
	})
	if err != nil {
		t.Fatalf("UpdateDraft failed: %v", err)
	}

	err = uc.PublishDraft(1, resp.DraftSetID)
	if err != nil {
		t.Fatalf("PublishDraft failed: %v", err)
	}
}

func TestPublishDraft_WithoutTitle_ReturnsError(t *testing.T) {
	uc := newSellUC()
	resp, _ := uc.CreateDraft(1, "ギター")

	// Update draft: clear title, keep price at 0 → should fail publish
	err := uc.UpdateDraft(1, resp.DraftSetID, domain.UpdateDraftRequest{
		Title: "",
		Price: 0,
	})
	if err != nil {
		t.Fatalf("UpdateDraft should succeed: %v", err)
	}

	err = uc.PublishDraft(1, resp.DraftSetID)
	if err == nil {
		t.Error("expected error when publishing draft without title")
	}
}

func TestPublishDraft_Unauthorized(t *testing.T) {
	uc := newSellUC()
	resp, _ := uc.CreateDraft(1, "ギター")
	err := uc.PublishDraft(2, resp.DraftSetID) // different user
	if err == nil {
		t.Error("expected unauthorized error")
	}
}
