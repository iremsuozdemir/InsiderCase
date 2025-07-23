package services


import (
	"insider-league/Models"
	"math"
	"math/rand"
	"errors"
	"sort"
	"fmt"
)

// LeagueSimulator defines the core league operations
type LeagueSimulator interface {
	PlayWeek() error
	GetLeagueTable() []*TeamStats
	PrintLeagueTable() string
}

type GenerateLeague struct {
	Teams []models.Team
	Fixtures [][]models.Match
	Results []models.Match
	CurrentWeek int
	TeamStats map[string]*TeamStats
}

// TeamStats tracks individual team performance
type TeamStats struct {
	TeamName     string
	Played       int
	Won          int
	Drawn        int
	Lost         int
	GoalsFor     int
	GoalsAgainst int
	Points       int
	GoalDiff     int
}

func generateFixture(teams []models.Team) [][]models.Match {
	// Input validation
	if len(teams) < 2 {
		return [][]models.Match{} // Return empty fixtures for invalid input
	}
	
	var fixtures [][]models.Match
	var allMatches []models.Match

	// Generate all possible matches (each team vs each other twice)
	for i := 0; i < len(teams); i++ {
		for j := i + 1; j < len(teams); j++ {
			// First leg: team i home vs team j away
			match1 := models.Match{
				HomeTeam: teams[i].Name,
				AwayTeam: teams[j].Name,
			}
			// Second leg: team j home vs team i away
			match2 := models.Match{
				HomeTeam: teams[j].Name,
				AwayTeam: teams[i].Name,
			}
			allMatches = append(allMatches, match1, match2)
		}
	}

	// Shuffle all matches randomly
	for i := len(allMatches) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		allMatches[i], allMatches[j] = allMatches[j], allMatches[i]
	}

	// Group matches into weeks (2 matches per week)
	weekNumber := 1
	for index := 0; index < len(allMatches); index += 2 {
		match1 := allMatches[index]
		match2 := allMatches[index+1]

		match1.Week = weekNumber
		match2.Week = weekNumber

		week := []models.Match{match1, match2}
		fixtures = append(fixtures, week)

		weekNumber++
	}
	
	return fixtures
}

// CalculateTotalWeeks calculates the total number of weeks needed for a league
func CalculateTotalWeeks(teamCount int) int {
	if teamCount < 2 {
		return 0
	}
	// Formula: n * (n-1) / 2 where n = number of teams
	// Each team plays every other team twice (home and away)
	totalMatches := teamCount * (teamCount - 1)
	// 2 matches per week
	return totalMatches / 2
}

func NewGenerateLeague(teams []models.Team)*GenerateLeague {
	engine := GenerateLeague{
		Teams: teams,
		Fixtures: generateFixture(teams),
		CurrentWeek: 0,
		TeamStats: make(map[string]*TeamStats),
	}

	// Initialize TeamStats for all teams
	for _, team := range teams {
		engine.TeamStats[team.Name] = &TeamStats{
			TeamName: team.Name,
			Played: 0,
			Won: 0,
			Drawn: 0,
			Lost: 0,
			GoalsFor: 0,
			GoalsAgainst: 0,
			Points: 0,
			GoalDiff: 0,
		}
	}
	return &engine
}

func(l *GenerateLeague) PlayWeek() error {
	if l.CurrentWeek >= len(l.Fixtures){
		return errors.New("End of season, no more matches to play.")
	}

	currentMatches := l.Fixtures[l.CurrentWeek]

	for i := range currentMatches{
		homeTeamName := currentMatches[i].HomeTeam
		awayTeamName := currentMatches[i].AwayTeam 

		// Find the actual team objects
		var homeTeam, awayTeam models.Team
		for _, team := range l.Teams {
			if team.Name == homeTeamName {
				homeTeam = team
			}
			if team.Name == awayTeamName {
				awayTeam = team
			}
		}

		match, err := PlayMatch(homeTeam, awayTeam, l)
		if err != nil {
			return err
		}
		l.Results = append(l.Results, match)
		
		// Update league table after each match
		l.updateLeagueTable(match)
	}

	l.CurrentWeek++
	return nil
}

func PlayMatch(homeTeam, awayTeam models.Team, league *GenerateLeague) (models.Match, error) {
	// Calculate dynamic strengths based on form and home advantage
	homeStrength := calculateTeamStrength(homeTeam, league, true)  // true = home team
	awayStrength := calculateTeamStrength(awayTeam, league, false) // false = away team

	// Calculate win probability based on adjusted strengths
	totalStrength := homeStrength + awayStrength
	homeWinProb := homeStrength / totalStrength
	
	// Generate random outcome
	randomResult := rand.Float64()
	
	var homeScore, awayScore int
	
	// Determine match result based on adjusted team strength
	if randomResult < homeWinProb {
		// Home team wins - score based on strength difference
		strengthDiff := homeStrength - awayStrength
		homeScore = generateScore(homeStrength, strengthDiff, true)
		awayScore = generateScore(awayStrength, -strengthDiff, false)
	} else {
		// Away team wins - score based on strength difference
		strengthDiff := awayStrength - homeStrength
		awayScore = generateScore(awayStrength, strengthDiff, true)
		homeScore = generateScore(homeStrength, -strengthDiff, false)
	}
	
	// Handle potential draw (small chance, more likely if teams are close in strength)
	strengthDifference := math.Abs(homeStrength - awayStrength)
	drawChance := 0.20 - (strengthDifference * 0.05) // 20% base, decreases with strength difference
	if drawChance < 0.05 {
		drawChance = 0.05 // Minimum 5% chance
	}
	
	if rand.Float64() < drawChance {
		// Draw - both teams score similar amounts
		avgStrength := (homeStrength + awayStrength) / 2
		drawScore := generateDrawScore(avgStrength)
		homeScore = drawScore
		awayScore = drawScore
	}
	
	match := models.Match{
		HomeTeam: homeTeam.Name,
		AwayTeam: awayTeam.Name,
		HomeScore: homeScore,
		AwayScore: awayScore,
	}
	
	return match, nil
}

// calculateTeamStrength calculates dynamic team strength based on form and home advantage
func calculateTeamStrength(team models.Team, league *GenerateLeague, isHome bool) float64 {
	baseStrength := float64(team.Strength)
	
	// Home advantage (10% boost)
	if isHome {
		baseStrength *= 1.1
	}
	
	// Form factor based on recent results
	formBonus := calculateFormBonus(team.Name, league)
	baseStrength *= (1.0 + formBonus)
	
	return baseStrength
}

// calculateFormBonus calculates bonus based on the most recent match result
func calculateFormBonus(teamName string, league *GenerateLeague) float64 {
	lastResult := getLastResult(teamName, league)
	
	switch lastResult {
	case "W":
		return 0.05 // 5% boost after a win
	case "L":
		return -0.05 // 5% penalty after a loss
	case "D":
		return 0.0 // No effect after a draw
	default:
		return 0.0 // No recent matches
	}
}



// generateScore generates realistic score based on team strength and strength difference
func generateScore(teamStrength, strengthDiff float64, isWinner bool) int {
	baseGoals := int(teamStrength / 20) // Base goals from strength
	
	if isWinner {
		// Winner gets bonus goals based on strength difference
		bonusGoals := int(strengthDiff / 10)
		if bonusGoals < 0 {
			bonusGoals = 0
		}
		totalGoals := baseGoals + bonusGoals + rand.Intn(3) // Add some randomness
		
		// Ensure realistic score range
		if totalGoals < 1 {
			totalGoals = 1
		} else if totalGoals > 5 {
			totalGoals = 5
		}
		return totalGoals
	} else {
		// Loser gets fewer goals
		penaltyGoals := int(strengthDiff / 15)
		if penaltyGoals < 0 {
			penaltyGoals = 0
		}
		totalGoals := baseGoals - penaltyGoals + rand.Intn(2)
		
		// Ensure realistic score range
		if totalGoals < 0 {
			totalGoals = 0
		} else if totalGoals > 3 {
			totalGoals = 3
		}
		return totalGoals
	}
}

// generateDrawScore generates score for a draw
func generateDrawScore(avgStrength float64) int {
	baseGoals := int(avgStrength / 25)
	randomGoals := rand.Intn(3)
	totalGoals := baseGoals + randomGoals
	
	// Ensure realistic draw score
	if totalGoals < 0 {
		totalGoals = 0
	} else if totalGoals > 2 {
		totalGoals = 2
	}
	return totalGoals
}

// getLastResult gets the most recent match result for a team
func getLastResult(teamName string, league *GenerateLeague) string {
	// Go through results in reverse order to find the most recent match
	for i := len(league.Results) - 1; i >= 0; i-- {
		match := league.Results[i]
		
		if match.HomeTeam == teamName {
			if match.HomeScore > match.AwayScore {
				return "W"
			} else if match.HomeScore == match.AwayScore {
				return "D"
			} else {
				return "L"
			}
		} else if match.AwayTeam == teamName {
			if match.AwayScore > match.HomeScore {
				return "W"
			} else if match.AwayScore == match.HomeScore {
				return "D"
			} else {
				return "L"
			}
		}
	}
	
	return "" // No previous matches found
}

// updateLeagueTable updates team statistics after a match
func (l *GenerateLeague) updateLeagueTable(match models.Match) {
	homeStats := l.TeamStats[match.HomeTeam]
	awayStats := l.TeamStats[match.AwayTeam]
	
	// Update goals
	homeStats.GoalsFor += match.HomeScore
	homeStats.GoalsAgainst += match.AwayScore
	awayStats.GoalsFor += match.AwayScore
	awayStats.GoalsAgainst += match.HomeScore
	
	// Update matches played
	homeStats.Played++
	awayStats.Played++
	
	// Update results and points based on Premier League rules
	if match.HomeScore > match.AwayScore {
		// Home team wins
		homeStats.Won++
		awayStats.Lost++
		homeStats.Points += 3 // 3 points for win
		awayStats.Points += 0 // 0 points for loss
	} else if match.HomeScore < match.AwayScore {
		// Away team wins
		awayStats.Won++
		homeStats.Lost++
		awayStats.Points += 3 // 3 points for win
		homeStats.Points += 0 // 0 points for loss
	} else {
		// Draw
		homeStats.Drawn++
		awayStats.Drawn++
		homeStats.Points += 1 // 1 point for draw
		awayStats.Points += 1 // 1 point for draw
	}
	
	// Update goal difference
	homeStats.GoalDiff = homeStats.GoalsFor - homeStats.GoalsAgainst
	awayStats.GoalDiff = awayStats.GoalsFor - awayStats.GoalsAgainst
}

// GetLeagueTable returns the current league standings sorted by points, goal difference, and goals scored
func (l *GenerateLeague) GetLeagueTable() []*TeamStats {
	var standings []*TeamStats
	
	// Convert map to slice
	for _, stats := range l.TeamStats {
		standings = append(standings, stats)
	}
	
	// Sort by Premier League rules: Points (desc), Goal Difference (desc), Goals For (desc)
	sort.Slice(standings, func(i, j int) bool {
		// First by points (descending)
		if standings[i].Points != standings[j].Points {
			return standings[i].Points > standings[j].Points
		}
		// Then by goal difference (descending)
		if standings[i].GoalDiff != standings[j].GoalDiff {
			return standings[i].GoalDiff > standings[j].GoalDiff
		}
		// Finally by goals scored (descending)
		return standings[i].GoalsFor > standings[j].GoalsFor
	})
	
	return standings
}

// GetTeamPosition returns the current league position of a team (1-based)
func (l *GenerateLeague) GetTeamPosition(teamName string) int {
	standings := l.GetLeagueTable()
	
	for i, stats := range standings {
		if stats.TeamName == teamName {
			return i + 1 // Return 1-based position
		}
	}
	
	return 0 // Team not found
}


