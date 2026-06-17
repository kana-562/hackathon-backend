package usecase

import (
	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type HomeUsecase interface {
	GetHome(userID int64) (*domain.HomeResponse, error)
	GetCategories() ([]domain.HobbyCategory, error)
}

type homeUsecase struct {
	categoryRepo repository.CategoryRepository
	setRepo      repository.SetRepository
}

func NewHomeUsecase(categoryRepo repository.CategoryRepository, setRepo repository.SetRepository) HomeUsecase {
	return &homeUsecase{categoryRepo: categoryRepo, setRepo: setRepo}
}

func (u *homeUsecase) GetHome(userID int64) (*domain.HomeResponse, error) {
	categories, err := u.categoryRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Featured: readiness > 70, on_sale
	featuredSets, err := u.setRepo.FindAll(repository.SetFilter{
		Status:            domain.SetStatusOnSale,
		MinReadinessScore: 70,
		Sort:              "readiness",
		Limit:             6,
	})
	if err != nil {
		return nil, err
	}

	// New: latest on_sale
	newSets, err := u.setRepo.FindAll(repository.SetFilter{
		Status: domain.SetStatusOnSale,
		Sort:   "new",
		Limit:  10,
	})
	if err != nil {
		return nil, err
	}

	featuredCards := make([]domain.StarterSetCard, 0, len(featuredSets))
	for _, s := range featuredSets {
		card := toSetCard(s)
		if userID > 0 {
			// isFavorite check could be added here
		}
		featuredCards = append(featuredCards, card)
	}

	newCards := make([]domain.StarterSetCard, 0, len(newSets))
	for _, s := range newSets {
		card := toSetCard(s)
		newCards = append(newCards, card)
	}

	return &domain.HomeResponse{
		FeaturedSets: featuredCards,
		NewSets:      newCards,
		Categories:   categories,
	}, nil
}

func (u *homeUsecase) GetCategories() ([]domain.HobbyCategory, error) {
	return u.categoryRepo.FindAll()
}

func toSetCard(s domain.StarterSet) domain.StarterSetCard {
	imageURL := ""
	if len(s.Images) > 0 {
		imageURL = s.Images[0].ImageURL
	}
	hobbyName := ""
	if s.Hobby != nil {
		hobbyName = s.Hobby.Name
	}
	categoryName := ""
	if s.Category != nil {
		categoryName = s.Category.Name
	}
	return domain.StarterSetCard{
		ID:             s.ID,
		Title:          s.Title,
		Price:          s.Price,
		BeginnerScore:  s.BeginnerScore,
		ReadinessScore: s.ReadinessScore,
		ImageURL:       imageURL,
		HobbyName:      hobbyName,
		CategoryName:   categoryName,
		Status:         s.Status,
	}
}
