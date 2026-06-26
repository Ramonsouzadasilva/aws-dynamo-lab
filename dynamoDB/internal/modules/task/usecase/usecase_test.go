package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	goalEntity "github.com/ramon/goals-tasks-api/internal/modules/goal/entity"
	"github.com/ramon/goals-tasks-api/internal/modules/task/dto"
	"github.com/ramon/goals-tasks-api/internal/modules/task/entity"
	"github.com/ramon/goals-tasks-api/internal/modules/task/usecase"
)

// MockGoalRepository implements goal/repository.GoalRepository interface
type MockGoalRepository struct {
	Goals map[string]*goalEntity.Goal
}

func (m *MockGoalRepository) Create(ctx context.Context, goal *goalEntity.Goal) error {
	m.Goals[goal.UserID+"#"+goal.ID] = goal
	return nil
}
func (m *MockGoalRepository) Update(ctx context.Context, goal *goalEntity.Goal) error {
	m.Goals[goal.UserID+"#"+goal.ID] = goal
	return nil
}
func (m *MockGoalRepository) Delete(ctx context.Context, userID, goalID string) error {
	delete(m.Goals, userID+"#"+goalID)
	return nil
}
func (m *MockGoalRepository) GetByID(ctx context.Context, userID, goalID string) (*goalEntity.Goal, error) {
	g, ok := m.Goals[userID+"#"+goalID]
	if !ok {
		return nil, errors.New("meta nao encontrada")
	}
	return g, nil
}
func (m *MockGoalRepository) ListActive(ctx context.Context, userID string) ([]*goalEntity.Goal, error) {
	var active []*goalEntity.Goal
	for _, g := range m.Goals {
		if g.UserID == userID && g.IsActive {
			active = append(active, g)
		}
	}
	return active, nil
}
func (m *MockGoalRepository) ListAll(ctx context.Context, userID string) ([]*goalEntity.Goal, error) {
	var all []*goalEntity.Goal
	for _, g := range m.Goals {
		if g.UserID == userID {
			all = append(all, g)
		}
	}
	return all, nil
}

type MockTaskRepository struct {
	Tasks map[string]*entity.Task
}

func (m *MockTaskRepository) Create(ctx context.Context, task *entity.Task) error {
	m.Tasks[task.UserID+"#"+task.ID] = task
	return nil
}

func (m *MockTaskRepository) CreateMultiple(ctx context.Context, tasks []*entity.Task) error {
	for _, t := range tasks {
		_ = m.Create(ctx, t)
	}
	return nil
}

func (m *MockTaskRepository) Update(ctx context.Context, task *entity.Task) error {
	m.Tasks[task.UserID+"#"+task.ID] = task
	return nil
}

func (m *MockTaskRepository) Delete(ctx context.Context, userID, taskID string) error {
	delete(m.Tasks, userID+"#"+taskID)
	return nil
}

func (m *MockTaskRepository) GetByID(ctx context.Context, userID, taskID string) (*entity.Task, error) {
	t, ok := m.Tasks[userID+"#"+taskID]
	if !ok {
		return nil, errors.New("tarefa nao encontrada")
	}
	return t, nil
}

func (m *MockTaskRepository) ListByStatus(ctx context.Context, userID string, status entity.TaskStatus) ([]*entity.Task, error) {
	var filtered []*entity.Task
	for _, t := range m.Tasks {
		if t.UserID == userID && t.Status == status {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

func (m *MockTaskRepository) ListByPeriod(ctx context.Context, userID string, start, end time.Time) ([]*entity.Task, error) {
	var filtered []*entity.Task
	for _, t := range m.Tasks {
		if t.UserID == userID && !t.DueDate.Before(start) && !t.DueDate.After(end) {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

func (m *MockTaskRepository) ListAllCompleted(ctx context.Context, userID string) ([]*entity.Task, error) {
	return m.ListByStatus(ctx, userID, entity.StatusCompleted)
}

func (m *MockTaskRepository) ListAll(ctx context.Context, userID string) ([]*entity.Task, error) {
	var all []*entity.Task
	for _, t := range m.Tasks {
		if t.UserID == userID {
			all = append(all, t)
		}
	}
	return all, nil
}

func (m *MockTaskRepository) CountTasksByGoal(ctx context.Context, goalID string) (completed, total int, err error) {
	for _, t := range m.Tasks {
		if t.GoalID == goalID {
			total++
			if t.Status == entity.StatusCompleted {
				completed++
			}
		}
	}
	return completed, total, nil
}

func TestTaskUseCase_CreateStandard(t *testing.T) {
	taskRepo := &MockTaskRepository{Tasks: make(map[string]*entity.Task)}
	goalRepo := &MockGoalRepository{Goals: make(map[string]*goalEntity.Goal)}
	uc := usecase.NewTaskUseCase(taskRepo, goalRepo)

	req := dto.CreateTaskRequest{
		Title:       "Estudar Go",
		Description: "Focar em testes unitários",
		DueDate:     time.Now(),
		IsRecurring: false,
	}

	tasks, err := uc.Create(context.Background(), "user1", req)
	if err != nil {
		t.Fatalf("esperava sem erro, obteve %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("esperava 1 tarefa, obteve %d", len(tasks))
	}

	if tasks[0].Title != "Estudar Go" {
		t.Errorf("esperava titulo 'Estudar Go', obteve '%s'", tasks[0].Title)
	}
}

func TestTaskUseCase_CreateRecurring(t *testing.T) {
	taskRepo := &MockTaskRepository{Tasks: make(map[string]*entity.Task)}
	goalRepo := &MockGoalRepository{Goals: make(map[string]*goalEntity.Goal)}
	uc := usecase.NewTaskUseCase(taskRepo, goalRepo)

	targetDate := time.Date(2026, time.June, 14, 12, 0, 0, 0, time.UTC) // A Sunday

	req := dto.CreateTaskRequest{
		Title:       "Fazer Exercícios",
		Description: "Treino de 30 mins",
		DueDate:     targetDate,
		IsRecurring: true,
	}

	tasks, err := uc.Create(context.Background(), "user1", req)
	if err != nil {
		t.Fatalf("esperava sem erro, obteve %v", err)
	}

	// Should create 5 tasks (Mon-Fri)
	if len(tasks) != 5 {
		t.Fatalf("esperava 5 tarefas, obteve %d", len(tasks))
	}

	expectedWeekdays := []string{
		"Fazer Exercícios (Segunda-feira)",
		"Fazer Exercícios (Terça-feira)",
		"Fazer Exercícios (Quarta-feira)",
		"Fazer Exercícios (Quinta-feira)",
		"Fazer Exercícios (Sexta-feira)",
	}

	for i, task := range tasks {
		if task.Title != expectedWeekdays[i] {
			t.Errorf("esperava titulo '%s', obteve '%s'", expectedWeekdays[i], task.Title)
		}
		if !task.IsRecurring {
			t.Errorf("esperava que a tarefa fosse recorrente")
		}
		// Confirm Monday of that week was June 8th, 2026
		expectedDay := 8 + i
		if task.DueDate.Day() != expectedDay {
			t.Errorf("esperava dia %d, obteve %d", expectedDay, task.DueDate.Day())
		}
	}
}
