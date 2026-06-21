package usecase

import (
	"errors"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/repository"
)

type MypageUsecase interface {
	GetMe(userID int64) (*domain.UserResponse, error)
	GetMyPage(userID int64) (*domain.MyPageResponse, error)
	GetSelling(userID int64) ([]domain.StarterSetCard, error)
	GetPurchases(userID int64) ([]domain.TransactionDetail, error)
	GetFavorites(userID int64) ([]domain.StarterSetCard, error)
}

type mypageUsecase struct {
	userRepo     repository.UserRepository
	setRepo      repository.SetRepository
	txRepo       repository.TransactionRepository
	favoriteRepo repository.FavoriteRepository
}

func NewMypageUsecase(
	userRepo repository.UserRepository,
	setRepo repository.SetRepository,
	txRepo repository.TransactionRepository,
	favoriteRepo repository.FavoriteRepository,
) MypageUsecase {
	return &mypageUsecase{
		userRepo:     userRepo,
		setRepo:      setRepo,
		txRepo:       txRepo,
		favoriteRepo: favoriteRepo,
	}
}

func (u *mypageUsecase) GetMe(userID int64) (*domain.UserResponse, error) {
	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return &domain.UserResponse{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		AvatarURL:   user.AvatarURL,
		RatingAvg:   user.RatingAverage,
	}, nil
}

func (u *mypageUsecase) GetMyPage(userID int64) (*domain.MyPageResponse, error) {
	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	sellingSets, err := u.setRepo.FindBySeller(userID)
	if err != nil {
		return nil, err
	}
	sellingCount := 0
	for _, s := range sellingSets {
		if s.Status == domain.SetStatusOnSale || s.Status == domain.SetStatusReserved {
			sellingCount++
		}
	}

	purchases, err := u.txRepo.FindByBuyer(userID)
	if err != nil {
		return nil, err
	}

	favorites, err := u.setRepo.FindFavorites(userID)
	if err != nil {
		return nil, err
	}

	return &domain.MyPageResponse{
		User: domain.UserResponse{
			ID:          user.ID,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			AvatarURL:   user.AvatarURL,
			RatingAvg:   user.RatingAverage,
		},
		SellingCount:   sellingCount,
		PurchasesCount: len(purchases),
		FavoritesCount: len(favorites),
	}, nil
}

func (u *mypageUsecase) GetSelling(userID int64) ([]domain.StarterSetCard, error) {
	sets, err := u.setRepo.FindBySeller(userID)
	if err != nil {
		return nil, err
	}

	cards := make([]domain.StarterSetCard, 0, len(sets))
	for _, s := range sets {
		if s.Status == domain.SetStatusOnSale || s.Status == domain.SetStatusReserved {
			cards = append(cards, toSetCard(s))
		}
	}
	return cards, nil
}

func (u *mypageUsecase) GetPurchases(userID int64) ([]domain.TransactionDetail, error) {
	txs, err := u.txRepo.FindByBuyer(userID)
	if err != nil {
		return nil, err
	}

	details := make([]domain.TransactionDetail, 0, len(txs))
	for _, tx := range txs {
		set, _ := u.setRepo.FindByID(tx.StarterSetID)
		var card domain.StarterSetCard
		if set != nil {
			card = toSetCard(*set)
		}
		details = append(details, domain.TransactionDetail{
			ID:         tx.ID,
			Status:     tx.Status,
			Price:      tx.Price,
			StarterSet: card,
			CreatedAt:  tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return details, nil
}

func (u *mypageUsecase) GetFavorites(userID int64) ([]domain.StarterSetCard, error) {
	sets, err := u.setRepo.FindFavorites(userID)
	if err != nil {
		return nil, err
	}

	cards := make([]domain.StarterSetCard, 0, len(sets))
	for _, s := range sets {
		card := toSetCard(s)
		card.IsFavorite = true
		cards = append(cards, card)
	}
	return cards, nil
}
