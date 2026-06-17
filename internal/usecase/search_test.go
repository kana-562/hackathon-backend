package usecase_test

import (
	"testing"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/testutil"
	"hobby-relay-backend/internal/usecase"
)

func newSetsUC() usecase.SetsUsecase {
	setRepo := testutil.NewMockSetRepository()
	favoriteRepo := testutil.NewMockFavoriteRepository()
	aiClient := testutil.NewMockAIClientControlled()
	return usecase.NewSetsUsecase(setRepo, favoriteRepo, aiClient)
}

func TestSearch_NoQuery_ReturnsAll(t *testing.T) {
	uc := newSetsUC()
	resp, err := uc.Search(domain.SearchQuery{}, 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestSmartSearch_MaxPriceFilter(t *testing.T) {
	uc := newSetsUC()
	resp, err := uc.Search(domain.SearchQuery{
		Q:     "1万円以内",
		Smart: true,
	}, 0)
	if err != nil {
		t.Fatalf("Smart search failed: %v", err)
	}
	if resp.SmartMessage == "" {
		t.Error("expected non-empty SmartMessage for smart search")
	}
}

func TestSmartSearch_HomeHobbies(t *testing.T) {
	uc := newSetsUC()
	resp, err := uc.Search(domain.SearchQuery{
		Q:     "家でできる",
		Smart: true,
	}, 0)
	if err != nil {
		t.Fatalf("Smart search failed: %v", err)
	}
	if len(resp.RelatedChips) == 0 {
		t.Error("expected non-empty RelatedChips for '家でできる' query")
	}
	// Verify specific hobbies are returned
	found := false
	for _, chip := range resp.RelatedChips {
		if chip == "コーヒー" || chip == "イラスト" || chip == "ヨガ" || chip == "ウクレレ" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected indoor hobbies in RelatedChips, got %v", resp.RelatedChips)
	}
}

func TestInterpretSearchQuery_MaxPrice10000(t *testing.T) {
	aiClient := testutil.NewMockAIClientControlled()

	queries := []string{"1万円以内", "10000円以内"}
	for _, q := range queries {
		result, err := aiClient.InterpretSearchQuery(q)
		if err != nil {
			t.Fatalf("InterpretSearchQuery(%q) failed: %v", q, err)
		}
		if result.MaxPrice != 10000 {
			t.Errorf("InterpretSearchQuery(%q): expected MaxPrice=10000, got %d", q, result.MaxPrice)
		}
	}
}

func TestInterpretSearchQuery_HomeHobbies(t *testing.T) {
	aiClient := testutil.NewMockAIClientControlled()

	result, err := aiClient.InterpretSearchQuery("家でできる")
	if err != nil {
		t.Fatalf("InterpretSearchQuery failed: %v", err)
	}
	if len(result.RelatedHobbies) == 0 {
		t.Error("expected non-empty RelatedHobbies")
	}
	if result.SmartMessage == "" {
		t.Error("expected non-empty SmartMessage")
	}
}

func TestSearch_SmartFalse_NoSmartMessage(t *testing.T) {
	uc := newSetsUC()
	resp, err := uc.Search(domain.SearchQuery{
		Q:     "ギター",
		Smart: false,
	}, 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.SmartMessage != "" {
		t.Error("expected empty SmartMessage when smart=false")
	}
}
