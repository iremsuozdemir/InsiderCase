package database

import (
	"fmt"
	"insider-league/Models"
)

// TeamRepository handles team-related database operations
type TeamRepository struct{}

// GetAllTeams retrieves all teams from the database
func (r *TeamRepository) GetAllTeams() ([]models.Team, error) {
	rows, err := DB.Query("SELECT id, name, strength FROM teams ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %v", err)
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var team models.Team
		if err := rows.Scan(&team.ID, &team.Name, &team.Strength); err != nil {
			return nil, fmt.Errorf("failed to scan team: %v", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// CreateLeague creates a new league in the database
func (r *TeamRepository) CreateLeague(name string, totalWeeks int) (int, error) {
	var leagueID int
	err := DB.QueryRow("INSERT INTO leagues (name, total_weeks) VALUES ($1, $2) RETURNING id", 
		name, totalWeeks).Scan(&leagueID)
	if err != nil {
		return 0, fmt.Errorf("failed to create league: %v", err)
	}
	return leagueID, nil
}

// AddTeamsToLeague adds teams to a league
func (r *TeamRepository) AddTeamsToLeague(leagueID int, teamIDs []int) error {
	for _, teamID := range teamIDs {
		_, err := DB.Exec("INSERT INTO league_teams (league_id, team_id) VALUES ($1, $2)", 
			leagueID, teamID)
		if err != nil {
			return fmt.Errorf("failed to add team to league: %v", err)
		}
	}
	return nil
}

// InitializeTeamStats initializes team statistics for a league
func (r *TeamRepository) InitializeTeamStats(leagueID int, teamIDs []int) error {
	for _, teamID := range teamIDs {
		_, err := DB.Exec("INSERT INTO team_stats (league_id, team_id) VALUES ($1, $2)", 
			leagueID, teamID)
		if err != nil {
			return fmt.Errorf("failed to initialize team stats: %v", err)
		}
	}
	return nil
}

// SaveMatch saves a match result to the database
func (r *TeamRepository) SaveMatch(leagueID int, match models.Match) error {
	// Get team IDs by name
	var homeTeamID, awayTeamID int
	err := DB.QueryRow("SELECT id FROM teams WHERE name = $1", match.HomeTeam).Scan(&homeTeamID)
	if err != nil {
		return fmt.Errorf("failed to get home team ID: %v", err)
	}
	
	err = DB.QueryRow("SELECT id FROM teams WHERE name = $1", match.AwayTeam).Scan(&awayTeamID)
	if err != nil {
		return fmt.Errorf("failed to get away team ID: %v", err)
	}

	// Insert match
	_, err = DB.Exec(`
		INSERT INTO matches (league_id, week_number, home_team_id, away_team_id, home_score, away_score, played) 
		VALUES ($1, $2, $3, $4, $5, $6, true)`,
		leagueID, match.Week, homeTeamID, awayTeamID, match.HomeScore, match.AwayScore)
	
	if err != nil {
		return fmt.Errorf("failed to save match: %v", err)
	}

	return nil
}

// UpdateTeamStats updates team statistics after a match
func (r *TeamRepository) UpdateTeamStats(leagueID int, teamName string, stats models.TeamStats) error {
	_, err := DB.Exec(`
		UPDATE team_stats 
		SET played = $1, won = $2, drawn = $3, lost = $4, 
		    goals_for = $5, goals_against = $6, points = $7
		WHERE league_id = $8 AND team_id = (SELECT id FROM teams WHERE name = $9)`,
		stats.Played, stats.Won, stats.Drawn, stats.Lost,
		stats.GoalsFor, stats.GoalsAgainst, stats.Points,
		leagueID, teamName)
	
	if err != nil {
		return fmt.Errorf("failed to update team stats: %v", err)
	}
	
	return nil
}

// GetLeagueTable retrieves the current league table
func (r *TeamRepository) GetLeagueTable(leagueID int) ([]models.TeamStats, error) {
	rows, err := DB.Query(`
		SELECT t.name, ts.played, ts.won, ts.drawn, ts.lost, 
		       ts.goals_for, ts.goals_against, ts.points, ts.goal_difference
		FROM team_stats ts
		JOIN teams t ON ts.team_id = t.id
		WHERE ts.league_id = $1
		ORDER BY ts.points DESC, ts.goal_difference DESC, ts.goals_for DESC`,
		leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to query league table: %v", err)
	}
	defer rows.Close()

	var table []models.TeamStats
	position := 1
	for rows.Next() {
		var stats models.TeamStats
		err := rows.Scan(&stats.TeamName, &stats.Played, &stats.Won, &stats.Drawn, &stats.Lost,
			&stats.GoalsFor, &stats.GoalsAgainst, &stats.Points, &stats.GoalDiff)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team stats: %v", err)
		}
		stats.Position = position
		table = append(table, stats)
		position++
	}

	return table, nil
}

// GetMatches retrieves all matches for a league
func (r *TeamRepository) GetMatches(leagueID int) ([]models.Match, error) {
	rows, err := DB.Query(`
		SELECT ht.name, at.name, m.home_score, m.away_score, m.week_number
		FROM matches m
		JOIN teams ht ON m.home_team_id = ht.id
		JOIN teams at ON m.away_team_id = at.id
		WHERE m.league_id = $1 AND m.played = true
		ORDER BY m.week_number, m.id`,
		leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to query matches: %v", err)
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var match models.Match
		err := rows.Scan(&match.HomeTeam, &match.AwayTeam, &match.HomeScore, &match.AwayScore, &match.Week)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %v", err)
		}
		matches = append(matches, match)
	}

	return matches, nil
}

// GetLeagueStatus retrieves the current league status
func (r *TeamRepository) GetLeagueStatus(leagueID int) (models.LeagueResponse, error) {
	var response models.LeagueResponse
	
	err := DB.QueryRow(`
		SELECT current_week, total_weeks, status
		FROM leagues WHERE id = $1`, leagueID).Scan(&response.CurrentWeek, &response.TotalWeeks, &response.Status)
	if err != nil {
		return response, fmt.Errorf("failed to get league status: %v", err)
	}
	
	// Calculate progress percentage
	if response.TotalWeeks > 0 {
		progress := float64(response.CurrentWeek) / float64(response.TotalWeeks) * 100
		response.Progress = fmt.Sprintf("%.1f%%", progress)
	}
	
	return response, nil
}

// ClearLeague clears all data for a league
func (r *TeamRepository) ClearLeague(leagueID int) error {
	// Delete matches
	_, err := DB.Exec("DELETE FROM matches WHERE league_id = $1", leagueID)
	if err != nil {
		return fmt.Errorf("failed to delete matches: %v", err)
	}
	
	// Reset team stats
	_, err = DB.Exec(`
		UPDATE team_stats 
		SET played = 0, won = 0, drawn = 0, lost = 0, 
		    goals_for = 0, goals_against = 0, points = 0
		WHERE league_id = $1`, leagueID)
	if err != nil {
		return fmt.Errorf("failed to reset team stats: %v", err)
	}
	
	// Reset league current week
	_, err = DB.Exec("UPDATE leagues SET current_week = 0 WHERE id = $1", leagueID)
	if err != nil {
		return fmt.Errorf("failed to reset league week: %v", err)
	}
	
	return nil
} 