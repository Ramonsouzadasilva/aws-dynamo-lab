package usecase

import (
	"context"

	"github.com/ramon/goals-tasks-api/internal/modules/dashboard/dto"
	goalRepo "github.com/ramon/goals-tasks-api/internal/modules/goal/repository"
	taskRepo "github.com/ramon/goals-tasks-api/internal/modules/task/repository"
	"github.com/ramon/goals-tasks-api/internal/modules/task/entity"
)

type DashboardUseCase struct {
	goalRepo goalRepo.GoalRepository
	taskRepo taskRepo.TaskRepository
}

func NewDashboardUseCase(goalRepo goalRepo.GoalRepository, taskRepo taskRepo.TaskRepository) *DashboardUseCase {
	return &DashboardUseCase{
		goalRepo: goalRepo,
		taskRepo: taskRepo,
	}
}

func (u *DashboardUseCase) GetDashboard(ctx context.Context, userID string) (*dto.DashboardResponse, error) {
	// 1. Get active goals
	goals, err := u.goalRepo.ListActive(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Get all tasks for user
	tasks, err := u.taskRepo.ListAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	activeGoalsCount := len(goals)
	totalTasksCount := len(tasks)
	completedTasksCount := 0

	tasksByStatus := map[string]int{
		string(entity.StatusPending):    0,
		string(entity.StatusInProgress): 0,
		string(entity.StatusCompleted):  0,
		string(entity.StatusCancelled):  0,
	}

	for _, task := range tasks {
		statusStr := string(task.Status)
		tasksByStatus[statusStr]++
		if task.Status == entity.StatusCompleted {
			completedTasksCount++
		}
	}

	// Compute overall progress
	var overallProgress float64
	if activeGoalsCount > 0 {
		var totalProgress float64
		for _, goal := range goals {
			completed, total, err := u.taskRepo.CountTasksByGoal(ctx, goal.ID)
			if err == nil {
				goal.ComputeProgress(completed, total)
			}
			totalProgress += goal.Progress
		}
		overallProgress = totalProgress / float64(activeGoalsCount)
	} else if totalTasksCount > 0 {
		overallProgress = (float64(completedTasksCount) / float64(totalTasksCount)) * 100
	} else {
		overallProgress = 0.0
	}

	return &dto.DashboardResponse{
		ActiveGoalsCount:    activeGoalsCount,
		TotalTasksCount:     totalTasksCount,
		CompletedTasksCount: completedTasksCount,
		TasksByStatus:       tasksByStatus,
		OverallProgress:     overallProgress,
	}, nil
}
