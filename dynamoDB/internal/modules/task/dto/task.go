package dto

import (
	"time"
	"github.com/ramon/goals-tasks-api/internal/modules/task/entity"
)

type CreateTaskRequest struct {
	GoalID      string    `json:"goal_id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	IsRecurring bool      `json:"is_recurring"`
}

type UpdateTaskRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	DueDate     time.Time         `json:"due_date"`
	Status      entity.TaskStatus `json:"status"`
}
