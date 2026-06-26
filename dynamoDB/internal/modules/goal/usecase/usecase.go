package usecase

import (
	"context"

	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	"github.com/ramon/goals-tasks-api/internal/modules/goal/dto"
	"github.com/ramon/goals-tasks-api/internal/modules/goal/entity"
	"github.com/ramon/goals-tasks-api/internal/modules/goal/repository"
	"github.com/ramon/goals-tasks-api/internal/utils"
)

// TaskCounter defines the minimal interface needed from TaskRepository to calculate Goal progress
type TaskCounter interface {
	CountTasksByGoal(ctx context.Context, goalID string) (completed, total int, err error)
}

type GoalUseCase struct {
	goalRepo    repository.GoalRepository
	taskCounter TaskCounter
}

func NewGoalUseCase(goalRepo repository.GoalRepository, taskCounter TaskCounter) *GoalUseCase {
	return &GoalUseCase{
		goalRepo:    goalRepo,
		taskCounter: taskCounter,
	}
}

func (u *GoalUseCase) Create(ctx context.Context, userID string, req dto.CreateGoalRequest) (*entity.Goal, error) {
	id := utils.GenerateUUID()
	goal, err := entity.NewGoal(id, userID, req.Title, req.Description, req.StartDate, req.EndDate)
	if err != nil {
		return nil, appErrors.NewAppError("VALIDATION_ERROR", err.Error(), 400)
	}

	err = u.goalRepo.Create(ctx, goal)
	if err != nil {
		return nil, err
	}

	return goal, nil
}

func (u *GoalUseCase) Update(ctx context.Context, userID, goalID string, req dto.UpdateGoalRequest) (*entity.Goal, error) {
	goal, err := u.goalRepo.GetByID(ctx, userID, goalID)
	if err != nil {
		return nil, err
	}

	err = goal.Update(req.Title, req.Description, req.StartDate, req.EndDate, req.IsActive)
	if err != nil {
		return nil, appErrors.NewAppError("VALIDATION_ERROR", err.Error(), 400)
	}

	err = u.goalRepo.Update(ctx, goal)
	if err != nil {
		return nil, err
	}

	// Compute progress dynamically
	completed, total, err := u.taskCounter.CountTasksByGoal(ctx, goalID)
	if err == nil {
		goal.ComputeProgress(completed, total)
	}

	return goal, nil
}

func (u *GoalUseCase) Delete(ctx context.Context, userID, goalID string) error {
	// Verify goal existence and ownership first
	_, err := u.goalRepo.GetByID(ctx, userID, goalID)
	if err != nil {
		return err
	}

	return u.goalRepo.Delete(ctx, userID, goalID)
}

func (u *GoalUseCase) Get(ctx context.Context, userID, goalID string) (*entity.Goal, error) {
	goal, err := u.goalRepo.GetByID(ctx, userID, goalID)
	if err != nil {
		return nil, err
	}

	// Calculate progress on the fly
	completed, total, err := u.taskCounter.CountTasksByGoal(ctx, goalID)
	if err != nil {
		return nil, err
	}

	goal.ComputeProgress(completed, total)
	return goal, nil
}

func (u *GoalUseCase) ListActive(ctx context.Context, userID string) ([]*entity.Goal, error) {
	goals, err := u.goalRepo.ListActive(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate progress for each active goal
	for _, goal := range goals {
		completed, total, err := u.taskCounter.CountTasksByGoal(ctx, goal.ID)
		if err == nil {
			goal.ComputeProgress(completed, total)
		}
	}

	return goals, nil
}

func (u *GoalUseCase) ListAll(ctx context.Context, userID string) ([]*entity.Goal, error) {
	goals, err := u.goalRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate progress for each goal
	for _, goal := range goals {
		completed, total, err := u.taskCounter.CountTasksByGoal(ctx, goal.ID)
		if err == nil {
			goal.ComputeProgress(completed, total)
		}
	}

	return goals, nil
}
