package usecase

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"hobby-relay-backend/internal/domain"
	"hobby-relay-backend/internal/middleware"
	"hobby-relay-backend/internal/repository"
)

type AuthUsecase interface {
	Signup(req domain.SignupRequest) (*domain.AuthResponse, error)
	Login(req domain.LoginRequest) (*domain.AuthResponse, error)
}

type authUsecase struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthUsecase(userRepo repository.UserRepository, jwtSecret string) AuthUsecase {
	return &authUsecase{userRepo: userRepo, jwtSecret: jwtSecret}
}

func (u *authUsecase) Signup(req domain.SignupRequest) (*domain.AuthResponse, error) {
	if req.DisplayName == "" || req.Email == "" || req.Password == "" {
		return nil, errors.New("all fields are required")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	existing, err := u.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		DisplayName:   req.DisplayName,
		Email:         req.Email,
		PasswordHash:  string(hash),
		AvatarURL:     "",
		RatingAverage: 0,
		RatingCount:   0,
	}

	id, err := u.userRepo.Create(user)
	if err != nil {
		return nil, err
	}
	user.ID = id

	token, err := u.generateToken(id)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User: domain.UserResponse{
			ID:          id,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			AvatarURL:   user.AvatarURL,
			RatingAvg:   user.RatingAverage,
		},
	}, nil
}

func (u *authUsecase) Login(req domain.LoginRequest) (*domain.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := u.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := u.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User: domain.UserResponse{
			ID:          user.ID,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			AvatarURL:   user.AvatarURL,
			RatingAvg:   user.RatingAverage,
		},
	}, nil
}

func (u *authUsecase) generateToken(userID int64) (string, error) {
	claims := middleware.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.jwtSecret))
}
