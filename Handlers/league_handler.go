package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"insider-league/Models"
	"insider-league/Services"
	"insider-league/database"
)

type LeagueHandler struct {
	league   *services.GenerateLeague
	repo     *database.TeamRepository
	leagueID int
}

func NewLeagueHandler() *LeagueHandler {
	return &LeagueHandler{
		repo: &database.TeamRepository{},
	}
}

// CreateLeague - POST /api/league
func (h *LeagueHandler) CreateLeague(w http.ResponseWriter, r *http.Request) {
	// Get teams from database 
	dbTeams, err := h.repo.GetAllTeams()
	if err != nil {
		http.Error(w, "Failed to get teams from database", http.StatusInternalServerError)
		return
	}
	
	if len(dbTeams) < 2 {
		http.Error(w, "At least 2 teams required in database", http.StatusBadRequest)
		return
	}
	
	// Create in-memory league for simulation first to get actual fixture count
	h.league = services.NewGenerateLeague(dbTeams)
	
	// Create league in database with actual number of weeks from fixtures
	leagueID, err := h.repo.CreateLeague("New League", len(h.league.Fixtures))
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
	
	// Store the league ID
	h.leagueID = leagueID
	
	// Store fixtures in database for consistency
	if err := h.repo.StoreFixtures(leagueID, h.league.Fixtures); err != nil {
		http.Error(w, "Failed to store fixtures", http.StatusInternalServerError)
		return
	}
	
	response := models.LeagueResponse{
		CurrentWeek: h.league.CurrentWeek,
		TotalWeeks:  len(h.league.Fixtures),
		Status:      "League created successfully",
	}
	
	// Set CORS headers (after Content-Type to ensure they're not overridden)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PlayWeek - POST /api/league/play-week
func (h *LeagueHandler) PlayWeek(w http.ResponseWriter, r *http.Request) {
	if h.leagueID == 0 {
		http.Error(w, "No league created yet. Please create a league first.", http.StatusBadRequest)
		return
	}
	
	// Play one week using the database method
	_, err := h.repo.PlayWeek(h.leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to play week: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Get updated league status
	leagueStatus, err := h.repo.GetLeagueStatus(h.leagueID)
	if err != nil {
		http.Error(w, "Failed to get league status", http.StatusInternalServerError)
		return
	}
	
	response := models.LeagueResponse{
		CurrentWeek: leagueStatus.CurrentWeek,
		TotalWeeks:  leagueStatus.TotalWeeks,
		Status:      fmt.Sprintf("Week %d played successfully", leagueStatus.CurrentWeek),
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PlayAllWeeks - POST /api/league/play-all
func (h *LeagueHandler) PlayAllWeeks(w http.ResponseWriter, r *http.Request) {
	if h.leagueID == 0 {
		http.Error(w, "No league created yet. Please create a league first.", http.StatusBadRequest)
		return
	}
	
	// Play all remaining weeks using the stored league ID
	matches, err := h.repo.PlayAllWeeks(h.leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to play all weeks: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Group matches by week for better display
	matchesByWeek := make(map[int][]models.Match)
	for _, match := range matches {
		matchesByWeek[match.Week] = append(matchesByWeek[match.Week], match)
	}
	
	response := map[string]interface{}{
		"status":        "All weeks played successfully",
		"total_matches": len(matches),
		"matches_by_week": matchesByWeek,
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetLeagueTable - GET /api/league/table
func (h *LeagueHandler) GetLeagueTable(w http.ResponseWriter, r *http.Request) {
	if h.leagueID == 0 {
		http.Error(w, "No league created yet. Please create a league first.", http.StatusBadRequest)
		return
	}
	
	// Get league table from database using the stored league ID
	standings, err := h.repo.GetLeagueTable(h.leagueID)
	if err != nil {
		http.Error(w, "Failed to get league table", http.StatusInternalServerError)
		return
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(standings)
}

// GetMatches - GET /api/league/matches
func (h *LeagueHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	if h.leagueID == 0 {
		http.Error(w, "No league created yet. Please create a league first.", http.StatusBadRequest)
		return
	}
	
	// Get matches from database using the stored league ID
	matches, err := h.repo.GetMatches(h.leagueID)
	if err != nil {
		http.Error(w, "Failed to get matches", http.StatusInternalServerError)
		return
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

// GetWeekMatches - GET /api/league/matches/week/{week}
func (h *LeagueHandler) GetWeekMatches(w http.ResponseWriter, r *http.Request) {
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
	
	if week < 1 {
		http.Error(w, "Week number must be positive", http.StatusBadRequest)
		return
	}
	
	// Get matches for the specific week from database using the stored league ID
	weekMatches, err := h.repo.GetMatchesByWeek(h.leagueID, week)
	if err != nil {
		http.Error(w, "Failed to get matches for week", http.StatusInternalServerError)
		return
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ClearLeague - DELETE /api/league
func (h *LeagueHandler) ClearLeague(w http.ResponseWriter, r *http.Request) {
	if h.leagueID == 0 {
		http.Error(w, "No league created yet. Please create a league first.", http.StatusBadRequest)
		return
	}
	
	// Clear league data from database using the stored league ID
	if err := h.repo.ClearLeague(h.leagueID); err != nil {
		http.Error(w, "Failed to clear league: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Reset in-memory league and league ID
	h.league = nil
	h.leagueID = 0
	
	response := models.LeagueResponse{
		CurrentWeek: 0,
		TotalWeeks:  0,
		Status:      "League cleared successfully",
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// InitializeDatabase - POST /api/init-db
func (h *LeagueHandler) InitializeDatabase(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.InitializeDatabase(); err != nil {
		http.Error(w, "Failed to initialize database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"status":  "Database initialized successfully",
		"message": "Default teams have been added to the database",
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ClearTeams - DELETE /api/clear-teams
func (h *LeagueHandler) ClearTeams(w http.ResponseWriter, r *http.Request) {
	// Clear all teams from database
	if err := h.repo.ClearTeams(); err != nil {
		http.Error(w, "Failed to clear teams: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"status":  "Teams cleared successfully",
		"message": "All teams have been removed from the database",
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTeams - GET /api/teams
func (h *LeagueHandler) GetTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.repo.GetAllTeams()
	if err != nil {
		http.Error(w, "Failed to get teams", http.StatusInternalServerError)
		return
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

// AddTeam - POST /api/teams
func (h *LeagueHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var teamRequest struct {
		Name     string `json:"name"`
		Strength int    `json:"strength"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&teamRequest); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	
	if teamRequest.Name == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}
	
	if teamRequest.Strength < 1 || teamRequest.Strength > 100 {
		http.Error(w, "Strength must be between 1 and 100", http.StatusBadRequest)
		return
	}
	
	if err := h.repo.AddTeam(teamRequest.Name, teamRequest.Strength); err != nil {
		http.Error(w, "Failed to add team: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"status":  "Team added successfully",
		"message": "Team " + teamRequest.Name + " has been added",
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMatchSchedule - GET /api/league/schedule
func (h *LeagueHandler) GetMatchSchedule(w http.ResponseWriter, r *http.Request) {
	if h.leagueID == 0 {
		http.Error(w, "No league created yet. Please create a league first.", http.StatusBadRequest)
		return
	}
	
	// Get match schedule from database using the stored league ID
	schedule, err := h.repo.GetMatchSchedule(h.leagueID)
	if err != nil {
		http.Error(w, "Failed to get match schedule", http.StatusInternalServerError)
		return
	}
	
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// GetChampionshipPredictions calculates and returns championship predictions
// GET /api/league/predictions

func (h *LeagueHandler) GetChampionshipPredictions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	// Get current league table
	standings, err := h.repo.GetLeagueTable(h.leagueID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get league table: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Calculate championship predictions
	predictions := h.calculateChampionshipPredictions(standings)
	
	// Return predictions as JSON
	if err := json.NewEncoder(w).Encode(predictions); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode predictions: %v", err), http.StatusInternalServerError)
		return
	}
}

// ChampionshipPrediction represents a team's championship probability
type ChampionshipPrediction struct {
	TeamName   string  `json:"team_name"`
	Percentage float64 `json:"percentage"`
	Points     int     `json:"points"`
	Position   int     `json:"position"`
}

// calculateChampionshipPredictions calculates championship percentages based on current points
func (h *LeagueHandler) calculateChampionshipPredictions(standings []models.TeamStats) []ChampionshipPrediction {
	var predictions []ChampionshipPrediction
	
	if len(standings) == 0 {
		return predictions
	}
	
	// Calculate total points across all teams
	totalPoints := 0
	for _, team := range standings {
		totalPoints += team.Points
	}
	
	// If no games played yet, give equal percentages
	if totalPoints == 0 {
		equalPercentage := 100.0 / float64(len(standings))
		for i, team := range standings {
			predictions = append(predictions, ChampionshipPrediction{
				TeamName:   team.TeamName,
				Percentage: equalPercentage,
				Points:     team.Points,
				Position:   i + 1,
			})
		}
		return predictions
	}
	
	// Group teams by points to handle ties properly
	teamsByPoints := make(map[int][]models.TeamStats)
	for _, team := range standings {
		teamsByPoints[team.Points] = append(teamsByPoints[team.Points], team)
	}
	
	// Sort points in descending order
	var sortedPoints []int
	for points := range teamsByPoints {
		sortedPoints = append(sortedPoints, points)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sortedPoints)))
	
	// Calculate predictions with tie handling
	position := 1
	for _, points := range sortedPoints {
		teamsWithSamePoints := teamsByPoints[points]
		
		// If multiple teams have same points, calculate equal share for their positions
		teamsCount := len(teamsWithSamePoints)
		positionRange := teamsCount
		
		// Calculate base percentage for this group of teams
		basePercentage := (float64(points) / float64(totalPoints)) * 100.0
		
		// Distribute percentage equally among tied teams
		percentagePerTeam := basePercentage / float64(teamsCount)
		
		for i, team := range teamsWithSamePoints {
			// Apply position-based adjustments
			positionMultiplier := 1.0
			currentPosition := position + i
			
			if currentPosition == 1 {
				positionMultiplier = 1.3 // Leader gets 30% bonus
			} else if currentPosition == 2 {
				positionMultiplier = 1.1 // Second place gets 10% bonus
			} else if currentPosition >= len(standings)-1 {
				positionMultiplier = 0.7 // Bottom teams get 30% penalty
			}
			
			finalPercentage := percentagePerTeam * positionMultiplier
			
			predictions = append(predictions, ChampionshipPrediction{
				TeamName:   team.TeamName,
				Percentage: finalPercentage,
				Points:     team.Points,
				Position:   currentPosition,
			})
		}
		
		position += positionRange
	}
	
	// Normalize percentages to sum to 100%
	totalPercentage := 0.0
	for _, pred := range predictions {
		totalPercentage += pred.Percentage
	}
	
	if totalPercentage > 0 {
		for i := range predictions {
			predictions[i].Percentage = (predictions[i].Percentage / totalPercentage) * 100.0
			// Round to 1 decimal place
			predictions[i].Percentage = float64(int(predictions[i].Percentage*10)) / 10.0
		}
	}
	
	return predictions
}
