package repository

import (
	"context"
	"time"

	"github.com/ramon/goals-tasks-api/internal/modules/task/entity"
)

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error
	CreateMultiple(ctx context.Context, tasks []*entity.Task) error
	Update(ctx context.Context, task *entity.Task) error
	Delete(ctx context.Context, userID, taskID string) error
	GetByID(ctx context.Context, userID, taskID string) (*entity.Task, error)
	ListByStatus(ctx context.Context, userID string, status entity.TaskStatus) ([]*entity.Task, error)
	ListByPeriod(ctx context.Context, userID string, start, end time.Time) ([]*entity.Task, error)
	ListAllCompleted(ctx context.Context, userID string) ([]*entity.Task, error)
	ListAll(ctx context.Context, userID string) ([]*entity.Task, error)
	CountTasksByGoal(ctx context.Context, goalID string) (completed, total int, err error)
}
