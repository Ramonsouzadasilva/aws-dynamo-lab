package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ramon/goals-tasks-api/internal/modules/dashboard/usecase"
	goalEntity "github.com/ramon/goals-tasks-api/internal/modules/goal/entity"
	taskEntity "github.com/ramon/goals-tasks-api/internal/modules/task/entity"
)

// ---- Mock Goal Repository ----

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

// ---- Mock Task Repository ----

type MockTaskRepository struct {
	Tasks map[string]*taskEntity.Task
}

func (m *MockTaskRepository) Create(ctx context.Context, task *taskEntity.Task) error {
	m.Tasks[task.UserID+"#"+task.ID] = task
	return nil
}
func (m *MockTaskRepository) CreateMultiple(ctx context.Context, tasks []*taskEntity.Task) error {
	for _, t := range tasks {
		_ = m.Create(ctx, t)
	}
	return nil
}
func (m *MockTaskRepository) Update(ctx context.Context, task *taskEntity.Task) error {
	m.Tasks[task.UserID+"#"+task.ID] = task
	return nil
}
func (m *MockTaskRepository) Delete(ctx context.Context, userID, taskID string) error {
	delete(m.Tasks, userID+"#"+taskID)
	return nil
}
func (m *MockTaskRepository) GetByID(ctx context.Context, userID, taskID string) (*taskEntity.Task, error) {
	t, ok := m.Tasks[userID+"#"+taskID]
	if !ok {
		return nil, errors.New("tarefa nao encontrada")
	}
	return t, nil
}
func (m *MockTaskRepository) ListByStatus(ctx context.Context, userID string, status taskEntity.TaskStatus) ([]*taskEntity.Task, error) {
	var filtered []*taskEntity.Task
	for _, t := range m.Tasks {
		if t.UserID == userID && t.Status == status {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}
func (m *MockTaskRepository) ListByPeriod(ctx context.Context, userID string, start, end time.Time) ([]*taskEntity.Task, error) {
	var filtered []*taskEntity.Task
	for _, t := range m.Tasks {
		if t.UserID == userID && !t.DueDate.Before(start) && !t.DueDate.After(end) {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}
func (m *MockTaskRepository) ListAllCompleted(ctx context.Context, userID string) ([]*taskEntity.Task, error) {
	return m.ListByStatus(ctx, userID, taskEntity.StatusCompleted)
}
func (m *MockTaskRepository) ListAll(ctx context.Context, userID string) ([]*taskEntity.Task, error) {
	var all []*taskEntity.Task
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
			if t.Status == taskEntity.StatusCompleted {
				completed++
			}
		}
	}
	return completed, total, nil
}

// ---- Tests ----

func TestDashboardUseCase_GetDashboard(t *testing.T) {
	goalRepo := &MockGoalRepository{Goals: make(map[string]*goalEntity.Goal)}
	taskRepo := &MockTaskRepository{Tasks: make(map[string]*taskEntity.Task)}
	uc := usecase.NewDashboardUseCase(goalRepo, taskRepo)

	// Cria meta ativa
	g, _ := goalEntity.NewGoal("goal1", "user1", "Meta 1", "Desc", time.Now(), time.Now().Add(24*time.Hour))
	_ = goalRepo.Create(context.Background(), g)

	// Cria 3 tarefas: 2 concluídas, 1 pendente
	t1, _ := taskEntity.NewTask("task1", "goal1", "user1", "Tarefa 1", "Desc", time.Now(), false)
	_ = t1.UpdateStatus(taskEntity.StatusCompleted)
	_ = taskRepo.Create(context.Background(), t1)

	t2, _ := taskEntity.NewTask("task2", "goal1", "user1", "Tarefa 2", "Desc", time.Now(), false)
	_ = t2.UpdateStatus(taskEntity.StatusCompleted)
	_ = taskRepo.Create(context.Background(), t2)

	t3, _ := taskEntity.NewTask("task3", "goal1", "user1", "Tarefa 3", "Desc", time.Now(), false)
	_ = taskRepo.Create(context.Background(), t3)

	dash, err := uc.GetDashboard(context.Background(), "user1")
	if err != nil {
		t.Fatalf("esperava sem erro, obteve %v", err)
	}

	if dash.ActiveGoalsCount != 1 {
		t.Errorf("esperava 1 meta ativa, obteve %d", dash.ActiveGoalsCount)
	}
	if dash.TotalTasksCount != 3 {
		t.Errorf("esperava 3 tarefas no total, obteve %d", dash.TotalTasksCount)
	}
	if dash.CompletedTasksCount != 2 {
		t.Errorf("esperava 2 tarefas concluidas, obteve %d", dash.CompletedTasksCount)
	}

	expectedProgress := (2.0 / 3.0) * 100
	if dash.OverallProgress < expectedProgress-0.1 || dash.OverallProgress > expectedProgress+0.1 {
		t.Errorf("esperava progresso ~66.6%%, obteve %f%%", dash.OverallProgress)
	}
}
