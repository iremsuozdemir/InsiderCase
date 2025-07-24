package database

import (
	"fmt"
	"insider-league/Models"
	"insider-league/Services"
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
	// Try to update first, if no rows affected, insert
	result, err := DB.Exec(`
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
	
	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	// If no rows were updated, insert a new record
	if rowsAffected == 0 {
		_, err = DB.Exec(`
			INSERT INTO team_stats (league_id, team_id, played, won, drawn, lost, goals_for, goals_against, points)
			VALUES ($1, (SELECT id FROM teams WHERE name = $2), $3, $4, $5, $6, $7, $8, $9)`,
			leagueID, teamName, stats.Played, stats.Won, stats.Drawn, stats.Lost,
			stats.GoalsFor, stats.GoalsAgainst, stats.Points)
		
		if err != nil {
			return fmt.Errorf("failed to insert team stats: %v", err)
		}
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
	// Start a transaction
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // Rollback if not committed
	
	// Delete matches
	_, err = tx.Exec("DELETE FROM matches WHERE league_id = $1", leagueID)
	if err != nil {
		return fmt.Errorf("failed to delete matches: %v", err)
	}
	
	// Delete team stats
	_, err = tx.Exec("DELETE FROM team_stats WHERE league_id = $1", leagueID)
	if err != nil {
		return fmt.Errorf("failed to delete team stats: %v", err)
	}
	
	// Delete league_teams associations
	_, err = tx.Exec("DELETE FROM league_teams WHERE league_id = $1", leagueID)
	if err != nil {
		return fmt.Errorf("failed to delete league teams: %v", err)
	}
	
	// Delete the league itself
	_, err = tx.Exec("DELETE FROM leagues WHERE id = $1", leagueID)
	if err != nil {
		return fmt.Errorf("failed to delete league: %v", err)
	}
	
	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	
	return nil
} 

// PlayWeek plays a single week for a league
func (r *TeamRepository) PlayWeek(leagueID int) ([]models.Match, error) {
	// Get current week
	var currentWeek int
	err := DB.QueryRow("SELECT current_week FROM leagues WHERE id = $1", leagueID).Scan(&currentWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get current week: %v", err)
	}
	
	// Get total weeks
	var totalWeeks int
	err = DB.QueryRow("SELECT total_weeks FROM leagues WHERE id = $1", leagueID).Scan(&totalWeeks)
	if err != nil {
		return nil, fmt.Errorf("failed to get total weeks: %v", err)
	}
	
	// Check if season is complete
	if currentWeek >= totalWeeks {
		return nil, fmt.Errorf("season is complete, no more weeks to play")
	}
	
	// Get all teams for the league
	teams, err := r.GetAllTeams()
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %v", err)
	}
	
	// Create in-memory league for simulation
	league := services.NewGenerateLeague(teams)
	
	// Check if fixtures exist in database, if not, store them
	var fixtureCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM matches WHERE league_id = $1", leagueID).Scan(&fixtureCount)
	if err != nil {
		return nil, fmt.Errorf("failed to check fixtures: %v", err)
	}
	
	if fixtureCount == 0 {
		// Store fixtures in database
		if err := r.StoreFixtures(leagueID, league.Fixtures); err != nil {
			return nil, fmt.Errorf("failed to store fixtures: %v", err)
		}
	} else {
		// Load fixtures from database
		fixtures, err := r.GetMatchSchedule(leagueID)
		if err != nil {
			return nil, fmt.Errorf("failed to load fixtures: %v", err)
		}
		
		// Convert map to slice format expected by league
		var fixturesSlice [][]models.Match
		for week := 1; week <= len(fixtures); week++ {
			if weekMatches, exists := fixtures[week]; exists {
				fixturesSlice = append(fixturesSlice, weekMatches)
			}
		}
		league.Fixtures = fixturesSlice
		
		// Load existing matches and stats from database
		existingMatches, err := r.GetMatches(leagueID)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing matches: %v", err)
		}
		
		// Add existing matches to the league
		for _, match := range existingMatches {
			league.Results = append(league.Results, match)
		}
		
		// Load existing team stats from database
		existingStats, err := r.GetLeagueTable(leagueID)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing stats: %v", err)
		}
		
		// Update in-memory stats with database stats
		for _, stat := range existingStats {
			if leagueStat, exists := league.TeamStats[stat.TeamName]; exists {
				leagueStat.Played = stat.Played
				leagueStat.Won = stat.Won
				leagueStat.Drawn = stat.Drawn
				leagueStat.Lost = stat.Lost
				leagueStat.GoalsFor = stat.GoalsFor
				leagueStat.GoalsAgainst = stat.GoalsAgainst
				leagueStat.Points = stat.Points
				leagueStat.GoalDiff = stat.GoalDiff

			}
		}
	}
	
	// Set the current week to match the database
	league.CurrentWeek = currentWeek
	
	// Play only the next week
	nextWeek := currentWeek + 1
	
	// Safety check to prevent index out of range
	if nextWeek-1 >= len(league.Fixtures) {
		return nil, fmt.Errorf("no fixtures available for week %d", nextWeek)
	}
	
	weekFixtures := league.Fixtures[nextWeek-1] // week-1 because fixtures are 0-indexed
	
	var weekMatches []models.Match
	
	// Play each match in this week
	for _, fixture := range weekFixtures {
		// Find the actual team objects
		var homeTeam, awayTeam models.Team
		for _, team := range teams {
			if team.Name == fixture.HomeTeam {
				homeTeam = team
			}
			if team.Name == fixture.AwayTeam {
				awayTeam = team
			}
		}
		
		// Play the match
		match, err := services.PlayMatch(homeTeam, awayTeam, league)
		if err != nil {
			return nil, fmt.Errorf("failed to play match: %v", err)
		}
		
		// Set the week number
		match.Week = nextWeek
		
		// Save match to database
		if err := r.SaveMatch(leagueID, match); err != nil {
			return nil, fmt.Errorf("failed to save match: %v", err)
		}
		
		// Update team stats
		homeStats := league.TeamStats[match.HomeTeam]
		if homeStats != nil {
			// Convert services.TeamStats to models.TeamStats
			modelStats := models.TeamStats{
				TeamName:     homeStats.TeamName,
				Played:       homeStats.Played,
				Won:          homeStats.Won,
				Drawn:        homeStats.Drawn,
				Lost:         homeStats.Lost,
				GoalsFor:     homeStats.GoalsFor,
				GoalsAgainst: homeStats.GoalsAgainst,
				Points:       homeStats.Points,
				GoalDiff:     homeStats.GoalDiff,
			}

			if err := r.UpdateTeamStats(leagueID, match.HomeTeam, modelStats); err != nil {
				return nil, fmt.Errorf("failed to update home team stats: %v", err)
			}
		}
		
		awayStats := league.TeamStats[match.AwayTeam]
		if awayStats != nil {
			// Convert services.TeamStats to models.TeamStats
			modelStats := models.TeamStats{
				TeamName:     awayStats.TeamName,
				Played:       awayStats.Played,
				Won:          awayStats.Won,
				Drawn:        awayStats.Drawn,
				Lost:         awayStats.Lost,
				GoalsFor:     awayStats.GoalsFor,
				GoalsAgainst: awayStats.GoalsAgainst,
				Points:       awayStats.Points,
				GoalDiff:     awayStats.GoalDiff,
			}

			if err := r.UpdateTeamStats(leagueID, match.AwayTeam, modelStats); err != nil {
				return nil, fmt.Errorf("failed to update away team stats: %v", err)
			}
		}
		
		weekMatches = append(weekMatches, match)
	}
	
	// Update league current week
	_, err = DB.Exec("UPDATE leagues SET current_week = $1 WHERE id = $2", nextWeek, leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to update league week: %v", err)
	}
	
	return weekMatches, nil
}

// PlayAllWeeks plays all remaining weeks in the league
func (r *TeamRepository) PlayAllWeeks(leagueID int) ([]models.Match, error) {
	// Get total weeks for the league
	var totalWeeks int
	err := DB.QueryRow("SELECT total_weeks FROM leagues WHERE id = $1", leagueID).Scan(&totalWeeks)
	if err != nil {
		return nil, fmt.Errorf("failed to get total weeks: %v", err)
	}
	
	// Get current week
	var currentWeek int
	err = DB.QueryRow("SELECT current_week FROM leagues WHERE id = $1", leagueID).Scan(&currentWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get current week: %v", err)
	}
	
	// Get all teams for the league
	teams, err := r.GetAllTeams()
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %v", err)
	}
	
	// Create in-memory league for simulation
	league := services.NewGenerateLeague(teams)
	
	// Check if fixtures exist in database, if not, store them
	var fixtureCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM matches WHERE league_id = $1", leagueID).Scan(&fixtureCount)
	if err != nil {
		return nil, fmt.Errorf("failed to check fixtures: %v", err)
	}
	
	if fixtureCount == 0 {
		// Store fixtures in database
		if err := r.StoreFixtures(leagueID, league.Fixtures); err != nil {
			return nil, fmt.Errorf("failed to store fixtures: %v", err)
		}
	}
	
	// Play all remaining weeks
	var allMatches []models.Match
	for week := currentWeek + 1; week <= totalWeeks; week++ {
		// Safety check to prevent index out of range
		if week-1 >= len(league.Fixtures) {
			return nil, fmt.Errorf("no fixtures available for week %d", week)
		}
		
		// Get the fixtures for this week
		weekFixtures := league.Fixtures[week-1] // week-1 because fixtures are 0-indexed
		
		// Play each match in this week
		for _, fixture := range weekFixtures {
			// Find the actual team objects
			var homeTeam, awayTeam models.Team
			for _, team := range teams {
				if team.Name == fixture.HomeTeam {
					homeTeam = team
				}
				if team.Name == fixture.AwayTeam {
					awayTeam = team
				}
			}
			
			// Play the match
			match, err := services.PlayMatch(homeTeam, awayTeam, league)
			if err != nil {
				return nil, fmt.Errorf("failed to play match: %v", err)
			}
			
			// Set the week number
			match.Week = week
			
			// Save match to database
			if err := r.SaveMatch(leagueID, match); err != nil {
				return nil, fmt.Errorf("failed to save match: %v", err)
			}
			
			// Update team stats
			homeStats := league.TeamStats[match.HomeTeam]
			if homeStats != nil {
				// Convert services.TeamStats to models.TeamStats
				modelStats := models.TeamStats{
					TeamName:     homeStats.TeamName,
					Played:       homeStats.Played,
					Won:          homeStats.Won,
					Drawn:        homeStats.Drawn,
					Lost:         homeStats.Lost,
					GoalsFor:     homeStats.GoalsFor,
					GoalsAgainst: homeStats.GoalsAgainst,
					Points:       homeStats.Points,
					GoalDiff:     homeStats.GoalDiff,
				}
				if err := r.UpdateTeamStats(leagueID, match.HomeTeam, modelStats); err != nil {
					return nil, fmt.Errorf("failed to update home team stats: %v", err)
				}
			}
			
			awayStats := league.TeamStats[match.AwayTeam]
			if awayStats != nil {
				// Convert services.TeamStats to models.TeamStats
				modelStats := models.TeamStats{
					TeamName:     awayStats.TeamName,
					Played:       awayStats.Played,
					Won:          awayStats.Won,
					Drawn:        awayStats.Drawn,
					Lost:         awayStats.Lost,
					GoalsFor:     awayStats.GoalsFor,
					GoalsAgainst: awayStats.GoalsAgainst,
					Points:       awayStats.Points,
					GoalDiff:     awayStats.GoalDiff,
				}
				if err := r.UpdateTeamStats(leagueID, match.AwayTeam, modelStats); err != nil {
					return nil, fmt.Errorf("failed to update away team stats: %v", err)
				}
			}
			
			allMatches = append(allMatches, match)
		}
	}
	
	// Update league current week to total weeks
	_, err = DB.Exec("UPDATE leagues SET current_week = $1 WHERE id = $2", totalWeeks, leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to update league week: %v", err)
	}
	
	return allMatches, nil
} 

// GetMatchesByWeek retrieves matches for a specific week
func (r *TeamRepository) GetMatchesByWeek(leagueID int, weekNumber int) ([]models.Match, error) {
	rows, err := DB.Query(`
		SELECT ht.name, at.name, m.home_score, m.away_score, m.week_number
		FROM matches m
		JOIN teams ht ON m.home_team_id = ht.id
		JOIN teams at ON m.away_team_id = at.id
		WHERE m.league_id = $1 AND m.week_number = $2 AND m.played = true
		ORDER BY m.id`,
		leagueID, weekNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query matches for week %d: %v", weekNumber, err)
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

// InitializeDatabase runs the schema.sql to set up the database
func (r *TeamRepository) InitializeDatabase() error {
	// First clear existing data in correct order (respecting foreign key constraints)
	_, err := DB.Exec("DELETE FROM matches")
	if err != nil {
		return fmt.Errorf("failed to clear existing matches: %v", err)
	}
	
	_, err = DB.Exec("DELETE FROM team_stats")
	if err != nil {
		return fmt.Errorf("failed to clear existing team stats: %v", err)
	}
	
	_, err = DB.Exec("DELETE FROM league_teams")
	if err != nil {
		return fmt.Errorf("failed to clear existing league teams: %v", err)
	}
	
	_, err = DB.Exec("DELETE FROM leagues")
	if err != nil {
		return fmt.Errorf("failed to clear existing leagues: %v", err)
	}
	
	_, err = DB.Exec("DELETE FROM teams")
	if err != nil {
		return fmt.Errorf("failed to clear existing teams: %v", err)
	}
	
	// Read and execute the schema.sql file
	schemaSQL := `
	-- Teams table
	CREATE TABLE IF NOT EXISTS teams (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL UNIQUE,
		strength INTEGER NOT NULL CHECK (strength >= 1 AND strength <= 100)
	);

	-- Leagues table  
	CREATE TABLE IF NOT EXISTS leagues (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		current_week INTEGER DEFAULT 0,
		total_weeks INTEGER NOT NULL,
		status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'paused'))
	);

	-- League teams (many-to-many relationship)
	CREATE TABLE IF NOT EXISTS league_teams (
		id SERIAL PRIMARY KEY,
		league_id INTEGER REFERENCES leagues(id) ON DELETE CASCADE,
		team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
		UNIQUE(league_id, team_id)
	);

	-- Matches table
	CREATE TABLE IF NOT EXISTS matches (
		id SERIAL PRIMARY KEY,
		league_id INTEGER REFERENCES leagues(id) ON DELETE CASCADE,
		week_number INTEGER NOT NULL,
		home_team_id INTEGER REFERENCES teams(id),
		away_team_id INTEGER REFERENCES teams(id),
		home_score INTEGER DEFAULT NULL,
		away_score INTEGER DEFAULT NULL,
		played BOOLEAN DEFAULT FALSE,
		played_at INTEGER DEFAULT NULL,
		CHECK (home_team_id != away_team_id)
	);

	-- Team statistics table
	CREATE TABLE IF NOT EXISTS team_stats (
		id SERIAL PRIMARY KEY,
		league_id INTEGER REFERENCES leagues(id) ON DELETE CASCADE,
		team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
		played INTEGER DEFAULT 0,
		won INTEGER DEFAULT 0,
		drawn INTEGER DEFAULT 0,
		lost INTEGER DEFAULT 0,
		goals_for INTEGER DEFAULT 0,
		goals_against INTEGER DEFAULT 0,
		points INTEGER DEFAULT 0,
		goal_difference INTEGER GENERATED ALWAYS AS (goals_for - goals_against) STORED,
		UNIQUE(league_id, team_id)
	);

	-- Insert sample teams (4 teams for smaller league)
	INSERT INTO teams (name, strength) VALUES 
		('Arsenal', 90),
		('Chelsea', 85),
		('Liverpool', 88),
		('Manchester City', 92);
	`
	
	_, err = DB.Exec(schemaSQL)
	return err
}

// AddTeam adds a new team to the database
func (r *TeamRepository) AddTeam(name string, strength int) error {
	_, err := DB.Exec("INSERT INTO teams (name, strength) VALUES ($1, $2)", name, strength)
	return err
}

// ClearTeams removes all teams from the database
func (r *TeamRepository) ClearTeams() error {
	_, err := DB.Exec("DELETE FROM teams")
	return err
}

// StoreFixtures stores the generated fixtures in the database
func (r *TeamRepository) StoreFixtures(leagueID int, fixtures [][]models.Match) error {
	// Get team IDs for mapping
	teams, err := r.GetAllTeams()
	if err != nil {
		return fmt.Errorf("failed to get teams: %v", err)
	}
	
	// Create team name to ID mapping
	teamMap := make(map[string]int)
	for _, team := range teams {
		teamMap[team.Name] = team.ID
	}
	
	// Store each fixture
	for weekIndex, weekMatches := range fixtures {
		weekNumber := weekIndex + 1 // Convert to 1-based week numbers
		for _, match := range weekMatches {
			homeTeamID := teamMap[match.HomeTeam]
			awayTeamID := teamMap[match.AwayTeam]
			
			_, err := DB.Exec(`
				INSERT INTO matches (league_id, week_number, home_team_id, away_team_id, played)
				VALUES ($1, $2, $3, $4, false)`,
				leagueID, weekNumber, homeTeamID, awayTeamID)
			if err != nil {
				return fmt.Errorf("failed to store fixture: %v", err)
			}
		}
	}
	
	return nil
}

// GetMatchSchedule gets the stored match schedule for a league
func (r *TeamRepository) GetMatchSchedule(leagueID int) (map[int][]models.Match, error) {
	// Get stored fixtures from database
	rows, err := DB.Query(`
		SELECT ht.name, at.name, m.week_number
		FROM matches m
		JOIN teams ht ON m.home_team_id = ht.id
		JOIN teams at ON m.away_team_id = at.id
		WHERE m.league_id = $1
		ORDER BY m.week_number, m.id`,
		leagueID)
	if err != nil {
		return nil, fmt.Errorf("failed to query match schedule: %v", err)
	}
	defer rows.Close()

	// Group matches by week
	schedule := make(map[int][]models.Match)
	for rows.Next() {
		var match models.Match
		err := rows.Scan(&match.HomeTeam, &match.AwayTeam, &match.Week)
		if err != nil {
			return nil, fmt.Errorf("failed to scan match: %v", err)
		}
		schedule[match.Week] = append(schedule[match.Week], match)
	}

	return schedule, nil
} 