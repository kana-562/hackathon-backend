package usecase

import (
	"errors"

	"hobby-relay-backend/internal/ai"
	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type SetsUsecase interface {
	Search(query domain.SearchQuery, userID int64) (*domain.SearchResponse, error)
	GetSetDetail(id, userID int64) (*domain.StarterSetDetail, error)
	GetCategorySets(categoryID, userID int64) ([]domain.StarterSetCard, error)
	AddFavorite(userID, setID int64) error
	RemoveFavorite(userID, setID int64) error
	AskQuestion(setID int64, userMsg string) (string, error)
}

type setsUsecase struct {
	setRepo      repository.SetRepository
	favoriteRepo repository.FavoriteRepository
	aiClient     ai.Client
}

func NewSetsUsecase(setRepo repository.SetRepository, favoriteRepo repository.FavoriteRepository, aiClient ai.Client) SetsUsecase {
	return &setsUsecase{setRepo: setRepo, favoriteRepo: favoriteRepo, aiClient: aiClient}
}

func (u *setsUsecase) Search(query domain.SearchQuery, userID int64) (*domain.SearchResponse, error) {
	filter := repository.SetFilter{
		Status:            domain.SetStatusOnSale,
		CategoryID:        query.CategoryID,
		HobbyID:           query.HobbyID,
		MaxPrice:          query.MaxPrice,
		MinBeginnerScore:  query.MinBeginnerScore,
		MinReadinessScore: query.MinReadinessScore,
		Sort:              query.Sort,
	}

	resp := &domain.SearchResponse{}

	if query.Smart && query.Q != "" {
		interp, err := u.aiClient.InterpretSearchQuery(query.Q)
		if err == nil {
			resp.SmartMessage = interp.SmartMessage
			resp.RelatedChips = interp.RelatedHobbies
			if interp.MaxPrice > 0 && filter.MaxPrice == 0 {
				filter.MaxPrice = interp.MaxPrice
			}
			if interp.MinBeginnerScore > 0 && filter.MinBeginnerScore == 0 {
				filter.MinBeginnerScore = interp.MinBeginnerScore
			}
			if interp.MinReadinessScore > 0 && filter.MinReadinessScore == 0 {
				filter.MinReadinessScore = interp.MinReadinessScore
			}
		}
	}

	sets, err := u.setRepo.FindAll(filter)
	if err != nil {
		return nil, err
	}

	cards := make([]domain.StarterSetCard, 0, len(sets))
	for _, s := range sets {
		card := toSetCard(s)
		if userID > 0 {
			fav, _ := u.favoriteRepo.Exists(userID, s.ID)
			card.IsFavorite = fav
		}
		cards = append(cards, card)
	}
	resp.Sets = cards
	return resp, nil
}

func (u *setsUsecase) GetSetDetail(id, userID int64) (*domain.StarterSetDetail, error) {
	s, err := u.setRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, errors.New("set not found")
	}

	images := make([]string, 0, len(s.Images))
	for _, img := range s.Images {
		images = append(images, img.ImageURL)
	}

	items := make([]domain.SetItemDTO, 0, len(s.Items))
	for _, item := range s.Items {
		items = append(items, domain.SetItemDTO{
			ID:             item.ID,
			Name:           item.Name,
			ConditionLabel: item.ConditionLabel,
			Quantity:       item.Quantity,
			IsEssential:    item.IsEssential,
			Note:           item.Note,
		})
	}

	recItems := make([]domain.RecommendedItemDTO, 0, len(s.RecommendedItems))
	for _, rec := range s.RecommendedItems {
		recItems = append(recItems, domain.RecommendedItemDTO{
			ID:         rec.ID,
			Name:       rec.Name,
			Importance: rec.Importance,
			Reason:     rec.Reason,
		})
	}

	var seller domain.UserResponse
	if s.Seller != nil {
		seller = domain.UserResponse{
			ID:          s.Seller.ID,
			DisplayName: s.Seller.DisplayName,
			Email:       s.Seller.Email,
			AvatarURL:   s.Seller.AvatarURL,
			RatingAvg:   s.Seller.RatingAverage,
		}
	}

	hobbyName := ""
	if s.Hobby != nil {
		hobbyName = s.Hobby.Name
	}
	categoryName := ""
	if s.Category != nil {
		categoryName = s.Category.Name
	}

	isFavorite := false
	if userID > 0 {
		isFavorite, _ = u.favoriteRepo.Exists(userID, id)
	}

	return &domain.StarterSetDetail{
		ID:                s.ID,
		Title:             s.Title,
		Price:             s.Price,
		Description:       s.Description,
		Status:            s.Status,
		BeginnerScore:     s.BeginnerScore,
		ReadinessScore:    s.ReadinessScore,
		ValueScore:        s.ValueScore,
		EstimatedNewPrice: s.EstimatedNewPrice,
		PreviousOwnerNote: s.PreviousOwnerNote,
		StartableSummary:  s.StartableSummary,
		Images:            images,
		Items:             items,
		RecommendedItems:  recItems,
		Seller:            seller,
		HobbyName:         hobbyName,
		CategoryName:      categoryName,
		IsFavorite:        isFavorite,
	}, nil
}

func (u *setsUsecase) GetCategorySets(categoryID, userID int64) ([]domain.StarterSetCard, error) {
	sets, err := u.setRepo.FindAll(repository.SetFilter{
		CategoryID: categoryID,
		Status:     domain.SetStatusOnSale,
	})
	if err != nil {
		return nil, err
	}

	cards := make([]domain.StarterSetCard, 0, len(sets))
	for _, s := range sets {
		card := toSetCard(s)
		if userID > 0 {
			fav, _ := u.favoriteRepo.Exists(userID, s.ID)
			card.IsFavorite = fav
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func (u *setsUsecase) AddFavorite(userID, setID int64) error {
	set, err := u.setRepo.FindByID(setID)
	if err != nil {
		return err
	}
	if set == nil {
		return errors.New("set not found")
	}
	return u.favoriteRepo.Add(userID, setID)
}

func (u *setsUsecase) RemoveFavorite(userID, setID int64) error {
	return u.favoriteRepo.Remove(userID, setID)
}

func (u *setsUsecase) AskQuestion(setID int64, userMsg string) (string, error) {
	set, err := u.setRepo.FindByID(setID)
	if err != nil {
		return "", err
	}
	if set == nil {
		return "", errors.New("set not found")
	}

	itemNames := make([]string, 0, len(set.Items))
	for _, item := range set.Items {
		itemNames = append(itemNames, item.Name)
	}

	return u.aiClient.AnswerSetQuestion(set.Title, itemNames, userMsg)
}
