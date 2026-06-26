package repository

import (
	"context"

	"github.com/ramon/goals-tasks-api/internal/modules/goal/entity"
)

type GoalRepository interface {
	Create(ctx context.Context, goal *entity.Goal) error
	Update(ctx context.Context, goal *entity.Goal) error
	Delete(ctx context.Context, userID, goalID string) error
	GetByID(ctx context.Context, userID, goalID string) (*entity.Goal, error)
	ListActive(ctx context.Context, userID string) ([]*entity.Goal, error)
	ListAll(ctx context.Context, userID string) ([]*entity.Goal, error)
}
