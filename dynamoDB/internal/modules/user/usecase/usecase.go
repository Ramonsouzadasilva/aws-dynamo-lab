package usecase

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/modules/user/dto"
	"github.com/ramon/goals-tasks-api/internal/modules/user/entity"
	"github.com/ramon/goals-tasks-api/internal/modules/user/repository"
	"github.com/ramon/goals-tasks-api/internal/utils"
)

type AuthUseCase struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthUseCase(userRepo repository.UserRepository, jwtSecret string) *AuthUseCase {
	return &AuthUseCase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (u *AuthUseCase) Register(ctx context.Context, req dto.RegisterRequest) (*entity.User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, appErrors.ErrBadRequest
	}

	// Create user domain entity (handles password hashing)
	id := utils.GenerateUUID()
	user, err := entity.NewUser(id, req.Email, req.Password)
	if err != nil {
		return nil, appErrors.NewAppError("VALIDATION_ERROR", err.Error(), 400)
	}

	// Store in repository
	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u *AuthUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponseData, error) {
	if req.Email == "" || req.Password == "" {
		return nil, appErrors.ErrBadRequest
	}

	// Fetch user from repo
	user, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Do not leak exact reason (security best practice)
		return nil, appErrors.ErrInvalidCredentials
	}

	// Validate password
	if !user.CheckPassword(req.Password) {
		return nil, appErrors.ErrInvalidCredentials
	}

	// Generate JWT
	tokenString, err := u.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponseData{
		Token: tokenString,
		Email: user.Email,
		ID:    user.ID,
	}, nil
}

func (u *AuthUseCase) generateToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", appErrors.ErrInternalServer
	}

	return tokenString, nil
}
