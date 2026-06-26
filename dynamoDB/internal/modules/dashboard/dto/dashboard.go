package dto

type DashboardResponse struct {
	ActiveGoalsCount    int            `json:"active_goals_count"`
	TotalTasksCount     int            `json:"total_tasks_count"`
	CompletedTasksCount int            `json:"completed_tasks_count"`
	TasksByStatus       map[string]int `json:"tasks_by_status"`
	OverallProgress     float64        `json:"overall_progress"`
}
