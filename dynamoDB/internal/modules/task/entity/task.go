package entity

import (
	"errors"
	"time"
)

type TaskStatus string

const (
	StatusPending    TaskStatus = "PENDING"
	StatusInProgress TaskStatus = "IN_PROGRESS"
	StatusCompleted  TaskStatus = "COMPLETED"
	StatusCancelled  TaskStatus = "CANCELLED"
)

type Task struct {
	ID          string     `json:"id" dynamodbav:"id"`
	GoalID      string     `json:"goal_id,omitempty" dynamodbav:"goal_id,omitempty"`
	UserID      string     `json:"user_id" dynamodbav:"user_id"`
	Title       string     `json:"title" dynamodbav:"title"`
	Description string     `json:"description" dynamodbav:"description"`
	Status      TaskStatus `json:"status" dynamodbav:"status"`
	DueDate     time.Time  `json:"due_date" dynamodbav:"due_date"`
	IsRecurring bool       `json:"is_recurring" dynamodbav:"is_recurring"`
	CreatedAt   time.Time  `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" dynamodbav:"updated_at"`
}

func NewTask(id, goalID, userID, title, description string, dueDate time.Time, isRecurring bool) (*Task, error) {
	if title == "" {
		return nil, errors.New("o título da tarefa não pode ser vazio")
	}

	return &Task{
		ID:          id,
		GoalID:      goalID,
		UserID:      userID,
		Title:       title,
		Description: description,
		Status:      StatusPending,
		DueDate:     dueDate.UTC(),
		IsRecurring: isRecurring,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

func (t *Task) UpdateStatus(status TaskStatus) error {
	switch status {
	case StatusPending, StatusInProgress, StatusCompleted, StatusCancelled:
		t.Status = status
		t.UpdatedAt = time.Now().UTC()
		return nil
	default:
		return errors.New("status de tarefa inválido")
	}
}

func (t *Task) Update(title, description string, dueDate time.Time, status TaskStatus) error {
	if title == "" {
		return errors.New("o título da tarefa não pode ser vazio")
	}
	t.Title = title
	t.Description = description
	t.DueDate = dueDate.UTC()
	t.UpdatedAt = time.Now().UTC()
	return t.UpdateStatus(status)
}

// GetWeekdays returns Mon-Fri dates for the week of the given date
func GetWeekdays(t time.Time) []time.Time {
	wd := t.Weekday()
	var offset int
	if wd == time.Sunday {
		offset = -6
	} else {
		offset = -int(wd - time.Monday)
	}

	monday := t.AddDate(0, 0, offset)
	
	weekdays := make([]time.Time, 5)
	for i := 0; i < 5; i++ {
		// Set to midnight UTC for clean comparison
		m := monday.AddDate(0, 0, i)
		weekdays[i] = time.Date(m.Year(), m.Month(), m.Day(), 0, 0, 0, 0, time.UTC)
	}
	return weekdays
}
