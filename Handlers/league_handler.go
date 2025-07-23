package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"insider-league/Models"
	"insider-league/Services"
	"insider-league/database"
)

type LeagueHandler struct {
	league *services.GenerateLeague
	repo   *database.TeamRepository
}

func NewLeagueHandler() *LeagueHandler {
	return &LeagueHandler{
		repo: &database.TeamRepository{},
	}
}

// CreateLeague - POST /api/league
func (h *LeagueHandler) CreateLeague(w http.ResponseWriter, r *http.Request) {
	var teams []models.Team
	
	if err := json.NewDecoder(r.Body).Decode(&teams); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	
	if len(teams) < 2 {
		http.Error(w, "At least 2 teams required", http.StatusBadRequest)
		return
	}
	
	// Get teams from database instead of request
	dbTeams, err := h.repo.GetAllTeams()
	if err != nil {
		http.Error(w, "Failed to get teams from database", http.StatusInternalServerError)
		return
	}
	
	// Create league in database
	leagueID, err := h.repo.CreateLeague("New League", services.CalculateTotalWeeks(len(dbTeams)))
	if err != nil {
		http.Error(w, "Failed to create league", http.StatusInternalServerError)
		return
	}
	
	// Add teams to league
	var teamIDs []int
	for _, team := range dbTeams {
		teamIDs = append(teamIDs, team.ID)
	}
	
	if err := h.repo.AddTeamsToLeague(leagueID, teamIDs); err != nil {
		http.Error(w, "Failed to add teams to league", http.StatusInternalServerError)
		return
	}
	
	// Initialize team stats
	if err := h.repo.InitializeTeamStats(leagueID, teamIDs); err != nil {
		http.Error(w, "Failed to initialize team stats", http.StatusInternalServerError)
		return
	}
	
	// Create in-memory league for simulation
	h.league = services.NewGenerateLeague(dbTeams)
	
	response := models.LeagueResponse{
		CurrentWeek: h.league.CurrentWeek,
		TotalWeeks:  services.CalculateTotalWeeks(len(dbTeams)),
		Status:      "League created successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PlayWeek - POST /api/league/play-week
func (h *LeagueHandler) PlayWeek(w http.ResponseWriter, r *http.Request) {
	if h.league == nil {
		http.Error(w, "League not initialized", http.StatusBadRequest)
		return
	}
	
	// Store current results count to know which matches are new
	previousResultsCount := len(h.league.Results)
	
	if err := h.league.PlayWeek(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	for i := previousResultsCount; i < len(h.league.Results); i++ {
		match := h.league.Results[i]
		if err := h.repo.SaveMatch(1, match); err != nil {
			http.Error(w, "Failed to save match to database", http.StatusInternalServerError)
			return
		}
		
		// Update team stats in database
		homeStats := h.league.TeamStats[match.HomeTeam]
		if homeStats != nil {
			if err := h.repo.UpdateTeamStats(1, match.HomeTeam, *homeStats); err != nil {
				http.Error(w, "Failed to update home team stats", http.StatusInternalServerError)
				return
			}
		}
		
		awayStats := h.league.TeamStats[match.AwayTeam]
		if awayStats != nil {
			if err := h.repo.UpdateTeamStats(1, match.AwayTeam, *awayStats); err != nil {
				http.Error(w, "Failed to update away team stats", http.StatusInternalServerError)
				return
			}
		}
	}
	
	// Calculate total weeks based on team count
	totalWeeks := services.CalculateTotalWeeks(len(h.league.Teams))
	
	response := models.LeagueResponse{
		CurrentWeek: h.league.CurrentWeek,
		TotalWeeks:  totalWeeks,
		Status:      "Week played successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLeagueTable - GET /api/league/table
func (h *LeagueHandler) GetLeagueTable(w http.ResponseWriter, r *http.Request) {
	// Get league table from database (assuming league ID 1 for now)
	standings, err := h.repo.GetLeagueTable(1)
	if err != nil {
		http.Error(w, "Failed to get league table", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(standings)
}

// GetMatches - GET /api/league/matches
func (h *LeagueHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	// Get matches from database (assuming league ID 1 for now)
	matches, err := h.repo.GetMatches(1)
	if err != nil {
		http.Error(w, "Failed to get matches", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

// GetWeekMatches - GET /api/league/matches/week/{week}
func (h *LeagueHandler) GetWeekMatches(w http.ResponseWriter, r *http.Request) {
	if h.league == nil {
		http.Error(w, "League not initialized", http.StatusBadRequest)
		return
	}
	
	// Extract week number from URL path: /api/league/matches/week/1
	path := strings.TrimPrefix(r.URL.Path, "/api/league/matches/week/")
	weekStr := strings.Split(path, "/")[0]
	
	if weekStr == "" {
		http.Error(w, "Week number required", http.StatusBadRequest)
		return
	}
	
	week, err := strconv.Atoi(weekStr)
	if err != nil {
		http.Error(w, "Invalid week number", http.StatusBadRequest)
		return
	}
	
	if week < 1 || week > len(h.league.Fixtures) {
		http.Error(w, "Week not found", http.StatusNotFound)
		return
	}
	
	weekMatches := h.league.Fixtures[week-1] // Convert to 0-based index
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weekMatches)
}

// GetLeagueStatus - GET /api/league/status
func (h *LeagueHandler) GetLeagueStatus(w http.ResponseWriter, r *http.Request) {
	if h.league == nil {
		http.Error(w, "League not initialized", http.StatusBadRequest)
		return
	}
	
	status := "In Progress"
	if h.league.CurrentWeek >= len(h.league.Fixtures) {
		status = "Season Complete"
	}
	
	response := models.LeagueResponse{
		CurrentWeek: h.league.CurrentWeek,
		TotalWeeks:  len(h.league.Fixtures),
		Status:      status,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ClearLeague - DELETE /api/league
func (h *LeagueHandler) ClearLeague(w http.ResponseWriter, r *http.Request) {
	// Clear league data from database 
	if err := h.repo.ClearLeague(1); err != nil {
		http.Error(w, "Failed to clear league", http.StatusInternalServerError)
		return
	}
	
	// Reset in-memory league
	h.league = nil
	
	response := models.LeagueResponse{
		CurrentWeek: 0,
		TotalWeeks:  0,
		Status:      "League cleared successfully",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
