package usecase

import (
	"context"
	"fmt"
	"time"

	appErrors "github.com/ramon/goals-tasks-api/internal/errors"
	goalRepo "github.com/ramon/goals-tasks-api/internal/modules/goal/repository"
	"github.com/ramon/goals-tasks-api/internal/modules/task/dto"
	"github.com/ramon/goals-tasks-api/internal/modules/task/entity"
	"github.com/ramon/goals-tasks-api/internal/modules/task/repository"
	"github.com/ramon/goals-tasks-api/internal/utils"
)

type TaskUseCase struct {
	taskRepo repository.TaskRepository
	goalRepo goalRepo.GoalRepository
}

func NewTaskUseCase(taskRepo repository.TaskRepository, goalRepo goalRepo.GoalRepository) *TaskUseCase {
	return &TaskUseCase{
		taskRepo: taskRepo,
		goalRepo: goalRepo,
	}
}

var weekdayNames = []string{
	"Segunda-feira",
	"Terça-feira",
	"Quarta-feira",
	"Quinta-feira",
	"Sexta-feira",
}

func (u *TaskUseCase) Create(ctx context.Context, userID string, req dto.CreateTaskRequest) ([]*entity.Task, error) {
	if req.Title == "" {
		return nil, appErrors.NewAppError("VALIDATION_ERROR", "o título da tarefa não pode ser vazio", 400)
	}

	// Verify Goal exists and belongs to user if GoalID is supplied
	if req.GoalID != "" {
		_, err := u.goalRepo.GetByID(ctx, userID, req.GoalID)
		if err != nil {
			return nil, err
		}
	}

	if req.IsRecurring {
		weekdays := entity.GetWeekdays(req.DueDate)
		tasks := make([]*entity.Task, 5)
		for i, day := range weekdays {
			id := utils.GenerateUUID()
			title := fmt.Sprintf("%s (%s)", req.Title, weekdayNames[i])
			task, err := entity.NewTask(id, req.GoalID, userID, title, req.Description, day, true)
			if err != nil {
				return nil, err
			}
			tasks[i] = task
		}

		err := u.taskRepo.CreateMultiple(ctx, tasks)
		if err != nil {
			return nil, err
		}

		return tasks, nil
	}

	id := utils.GenerateUUID()
	task, err := entity.NewTask(id, req.GoalID, userID, req.Title, req.Description, req.DueDate, false)
	if err != nil {
		return nil, err
	}

	err = u.taskRepo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	return []*entity.Task{task}, nil
}

func (u *TaskUseCase) Update(ctx context.Context, userID, taskID string, req dto.UpdateTaskRequest) (*entity.Task, error) {
	task, err := u.taskRepo.GetByID(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}

	err = task.Update(req.Title, req.Description, req.DueDate, req.Status)
	if err != nil {
		return nil, appErrors.NewAppError("VALIDATION_ERROR", err.Error(), 400)
	}

	err = u.taskRepo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (u *TaskUseCase) Delete(ctx context.Context, userID, taskID string) error {
	return u.taskRepo.Delete(ctx, userID, taskID)
}

func (u *TaskUseCase) Get(ctx context.Context, userID, taskID string) (*entity.Task, error) {
	return u.taskRepo.GetByID(ctx, userID, taskID)
}

func (u *TaskUseCase) ListByPeriod(ctx context.Context, userID string, start, end time.Time) ([]*entity.Task, error) {
	return u.taskRepo.ListByPeriod(ctx, userID, start, end)
}

func (u *TaskUseCase) ListWeekly(ctx context.Context, userID string) ([]*entity.Task, error) {
	now := time.Now().UTC()
	weekdays := entity.GetWeekdays(now)
	
	// Start from Monday 00:00:00 UTC
	start := weekdays[0]
	// End on Sunday 23:59:59 UTC (6 days after Monday)
	end := weekdays[0].AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	return u.taskRepo.ListByPeriod(ctx, userID, start, end)
}

func (u *TaskUseCase) ListCompletedHistory(ctx context.Context, userID string) ([]*entity.Task, error) {
	return u.taskRepo.ListAllCompleted(ctx, userID)
}

func (u *TaskUseCase) ListByStatus(ctx context.Context, userID string, status entity.TaskStatus) ([]*entity.Task, error) {
	return u.taskRepo.ListByStatus(ctx, userID, status)
}
