package usecase_test

import (
	"context"
	"errors"
	"testing"

	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/modules/user/dto"
	"github.com/ramon/goals-tasks-api/internal/modules/user/entity"
	"github.com/ramon/goals-tasks-api/internal/modules/user/usecase"
)

type MockUserRepository struct {
	Users map[string]*entity.User
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	for _, u := range m.Users {
		if u.Email == user.Email {
			return appErrors.ErrUserAlreadyExists
		}
	}
	m.Users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	for _, u := range m.Users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, appErrors.ErrUserNotFound
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	u, ok := m.Users[id]
	if !ok {
		return nil, appErrors.ErrUserNotFound
	}
	return u, nil
}

func TestAuthUseCase_RegisterAndLogin(t *testing.T) {
	repo := &MockUserRepository{Users: make(map[string]*entity.User)}
	uc := usecase.NewAuthUseCase(repo, "test-secret-key-very-secure")

	// 1. Test Register
	regReq := dto.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := uc.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("esperava sem erro no registro, obteve %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("esperava email 'test@example.com', obteve '%s'", user.Email)
	}

	// 2. Test Register duplicate email
	_, err = uc.Register(context.Background(), regReq)
	if !errors.Is(err, appErrors.ErrUserAlreadyExists) {
		t.Errorf("esperava erro de usuario duplicado, obteve %v", err)
	}

	// 3. Test Login successful
	loginReq := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	loginResp, err := uc.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("esperava sem erro no login, obteve %v", err)
	}

	if loginResp.Token == "" {
		t.Errorf("esperava token JWT, obteve vazio")
	}

	if loginResp.Email != "test@example.com" {
		t.Errorf("esperava email 'test@example.com', obteve '%s'", loginResp.Email)
	}

	// 4. Test Login invalid credentials
	badLoginReq := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	_, err = uc.Login(context.Background(), badLoginReq)
	if !errors.Is(err, appErrors.ErrInvalidCredentials) {
		t.Errorf("esperava erro de credenciais invalidas, obteve %v", err)
	}
}
