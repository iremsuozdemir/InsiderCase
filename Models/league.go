package models

type Team struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Strength int    `json:"strength"`
}

type Match struct {
	Week      int    `json:"week"`
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	HomeScore int    `json:"home_score"`
	AwayScore int    `json:"away_score"`
}

type TeamStats struct {
	TeamName     string `json:"team_name"`
	Played       int    `json:"played"`
	Won          int    `json:"won"`
	Drawn        int    `json:"drawn"`
	Lost         int    `json:"lost"`
	GoalsFor     int    `json:"goals_for"`
	GoalsAgainst int    `json:"goals_against"`
	Points       int    `json:"points"`
	GoalDiff     int    `json:"goal_difference"`
	Position     int    `json:"position"`
}

type LeagueResponse struct {
	CurrentWeek int    `json:"current_week"`
	TotalWeeks  int    `json:"total_weeks"`
	Status      string `json:"status"`
	Progress    string `json:"progress"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// New request/response models for team management
type AddTeamRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=50"`
	Strength int    `json:"strength" validate:"required,min=1,max=100"`
}

type UpdateTeamRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=50"`
	Strength int    `json:"strength" validate:"required,min=1,max=100"`
}

type TeamResponse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Strength int    `json:"strength"`
	Message  string `json:"message,omitempty"`
}

type TeamsResponse struct {
	Teams   []Team `json:"teams"`
	Count   int    `json:"count"`
	Message string `json:"message,omitempty"`
}

