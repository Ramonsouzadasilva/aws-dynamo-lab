package entity

import (
	"errors"
	"time"
)

type Goal struct {
	ID          string    `json:"id" dynamodbav:"id"`
	UserID      string    `json:"user_id" dynamodbav:"user_id"`
	Title       string    `json:"title" dynamodbav:"title"`
	Description string    `json:"description" dynamodbav:"description"`
	StartDate   time.Time `json:"start_date" dynamodbav:"start_date"`
	EndDate     time.Time `json:"end_date" dynamodbav:"end_date"`
	IsActive    bool      `json:"is_active" dynamodbav:"is_active"`
	Progress    float64   `json:"progress" dynamodbav:"progress"` // Dynamically calculated attribute
	CreatedAt   time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

func NewGoal(id, userID, title, description string, startDate, endDate time.Time) (*Goal, error) {
	if title == "" {
		return nil, errors.New("o título da meta não pode ser vazio")
	}
	if startDate.After(endDate) {
		return nil, errors.New("a data de início deve ser anterior ou igual à data de fim")
	}

	return &Goal{
		ID:          id,
		UserID:      userID,
		Title:       title,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		IsActive:    true,
		Progress:    0.0,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

func (g *Goal) Update(title, description string, startDate, endDate time.Time, isActive bool) error {
	if title == "" {
		return errors.New("o título da meta não pode ser vazio")
	}
	if startDate.After(endDate) {
		return errors.New("a data de início deve ser anterior ou igual à data de fim")
	}

	g.Title = title
	g.Description = description
	g.StartDate = startDate
	g.EndDate = endDate
	g.IsActive = isActive
	g.UpdatedAt = time.Now().UTC()
	return nil
}

func (g *Goal) ComputeProgress(completedTasks, totalTasks int) {
	if totalTasks <= 0 {
		g.Progress = 0.0
		return
	}
	g.Progress = (float64(completedTasks) / float64(totalTasks)) * 100
}
