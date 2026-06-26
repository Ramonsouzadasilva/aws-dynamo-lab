package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ramon/goals-tasks-api/internal/modules/goal/dto"
	"github.com/ramon/goals-tasks-api/internal/modules/goal/entity"
	"github.com/ramon/goals-tasks-api/internal/modules/goal/usecase"
)

type MockGoalRepository struct {
	Goals map[string]*entity.Goal
}

func (m *MockGoalRepository) Create(ctx context.Context, goal *entity.Goal) error {
	m.Goals[goal.UserID+"#"+goal.ID] = goal
	return nil
}

func (m *MockGoalRepository) Update(ctx context.Context, goal *entity.Goal) error {
	m.Goals[goal.UserID+"#"+goal.ID] = goal
	return nil
}

func (m *MockGoalRepository) Delete(ctx context.Context, userID, goalID string) error {
	delete(m.Goals, userID+"#"+goalID)
	return nil
}

func (m *MockGoalRepository) GetByID(ctx context.Context, userID, goalID string) (*entity.Goal, error) {
	g, ok := m.Goals[userID+"#"+goalID]
	if !ok {
		return nil, errors.New("meta nao encontrada")
	}
	return g, nil
}

func (m *MockGoalRepository) ListActive(ctx context.Context, userID string) ([]*entity.Goal, error) {
	var active []*entity.Goal
	for _, g := range m.Goals {
		if g.UserID == userID && g.IsActive {
			active = append(active, g)
		}
	}
	return active, nil
}

func (m *MockGoalRepository) ListAll(ctx context.Context, userID string) ([]*entity.Goal, error) {
	var all []*entity.Goal
	for _, g := range m.Goals {
		if g.UserID == userID {
			all = append(all, g)
		}
	}
	return all, nil
}

type MockTaskCounter struct {
	Completed int
	Total     int
}

func (m *MockTaskCounter) CountTasksByGoal(ctx context.Context, goalID string) (completed, total int, err error) {
	return m.Completed, m.Total, nil
}

func TestGoalUseCase_Create(t *testing.T) {
	repo := &MockGoalRepository{Goals: make(map[string]*entity.Goal)}
	counter := &MockTaskCounter{Completed: 0, Total: 0}
	uc := usecase.NewGoalUseCase(repo, counter)

	req := dto.CreateGoalRequest{
		Title:       "Minha Meta",
		Description: "Descrição da minha meta",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(24 * time.Hour),
	}

	goal, err := uc.Create(context.Background(), "user1", req)
	if err != nil {
		t.Fatalf("esperava sem erro, obteve %v", err)
	}

	if goal.Title != "Minha Meta" {
		t.Errorf("esperava titulo 'Minha Meta', obteve '%s'", goal.Title)
	}

	if goal.UserID != "user1" {
		t.Errorf("esperava user ID 'user1', obteve '%s'", goal.UserID)
	}
}

func TestGoalUseCase_GetWithProgress(t *testing.T) {
	repo := &MockGoalRepository{Goals: make(map[string]*entity.Goal)}
	counter := &MockTaskCounter{Completed: 2, Total: 5}
	uc := usecase.NewGoalUseCase(repo, counter)

	g, _ := entity.NewGoal("goal1", "user1", "Meta Teste", "Descricao", time.Now(), time.Now().Add(24*time.Hour))
	_ = repo.Create(context.Background(), g)

	goal, err := uc.Get(context.Background(), "user1", "goal1")
	if err != nil {
		t.Fatalf("esperava sem erro, obteve %v", err)
	}

	if goal.Progress != 40.0 {
		t.Errorf("esperava progresso 40%%, obteve %f%%", goal.Progress)
	}
}
