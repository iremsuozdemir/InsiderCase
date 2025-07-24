package services


import (
	"insider-league/Models"
	"math"
	"math/rand"
	"errors"
	"sort"
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

func GenerateFixture(teams []models.Team) [][]models.Match {
	// Input validation
	if len(teams) < 2 {
		return [][]models.Match{} // Return empty fixtures for invalid input
	}
	
	var fixtures [][]models.Match
	teamCount := len(teams)
	isOdd := teamCount%2 != 0
	
	// For odd numbers, we need to handle scheduling differently
	if isOdd {
		// Create a round-robin schedule where each team plays every other team twice
		// One team gets a bye each week
		
		// Calculate total weeks needed
		totalWeeks := teamCount * (teamCount - 1) / ((teamCount - 1) / 2)
		
		// Create fixtures week by week
		for week := 1; week <= totalWeeks; week++ {
			var weekMatches []models.Match
			teamsUsedThisWeek := make(map[string]bool)
			
			// Determine which team gets the bye this week
			// Rotate through teams for bye
			byeTeamIndex := (week - 1) % teamCount
			byeTeam := teams[byeTeamIndex].Name
			teamsUsedThisWeek[byeTeam] = true // Mark bye team as used
			
			// Create matches for the remaining teams
			matchesThisWeek := 0
			maxMatchesPerWeek := (teamCount - 1) / 2
			
			// Try to create matches between teams that haven't played this week
			for i := 0; i < teamCount && matchesThisWeek < maxMatchesPerWeek; i++ {
				for j := i + 1; j < teamCount && matchesThisWeek < maxMatchesPerWeek; j++ {
					team1 := teams[i].Name
					team2 := teams[j].Name
					
					// Skip if either team is the bye team or already used
					if team1 == byeTeam || team2 == byeTeam || 
					   teamsUsedThisWeek[team1] || teamsUsedThisWeek[team2] {
						continue
					}
					
					// Create match
					match := models.Match{
						HomeTeam: team1,
						AwayTeam: team2,
						Week:     week,
					}
					weekMatches = append(weekMatches, match)
					teamsUsedThisWeek[team1] = true
					teamsUsedThisWeek[team2] = true
					matchesThisWeek++
				}
			}
			
			// Add week to fixtures if it has matches
			if len(weekMatches) > 0 {
				fixtures = append(fixtures, weekMatches)
			}
		}
		
		return fixtures
	}
	
	// Even number of teams - use round-robin algorithm
	// For even teams, we can use a simple round-robin schedule
	// Each team plays every other team twice (home and away)
	
	// Calculate total weeks needed
	totalWeeks := teamCount * (teamCount - 1) / (teamCount / 2)
	
	// Create a round-robin schedule
	for week := 1; week <= totalWeeks; week++ {
		var weekMatches []models.Match
		teamsUsedThisWeek := make(map[string]bool)
		
		// For even teams, we can schedule all teams in pairs
		matchesThisWeek := 0
		maxMatchesPerWeek := teamCount / 2
		
		// Create matches for this week
		for i := 0; i < teamCount && matchesThisWeek < maxMatchesPerWeek; i++ {
			for j := i + 1; j < teamCount && matchesThisWeek < maxMatchesPerWeek; j++ {
				team1 := teams[i].Name
				team2 := teams[j].Name
				
				// Skip if either team is already used this week
				if teamsUsedThisWeek[team1] || teamsUsedThisWeek[team2] {
					continue
				}
				
				// Determine home/away based on week number for variety
				var homeTeam, awayTeam string
				if (week + i + j) % 2 == 0 {
					homeTeam, awayTeam = team1, team2
				} else {
					homeTeam, awayTeam = team2, team1
				}
				
				// Create match
				match := models.Match{
					HomeTeam: homeTeam,
					AwayTeam: awayTeam,
					Week:     week,
				}
				weekMatches = append(weekMatches, match)
				teamsUsedThisWeek[homeTeam] = true
				teamsUsedThisWeek[awayTeam] = true
				matchesThisWeek++
			}
		}
		
		// Add week to fixtures if it has matches
		if len(weekMatches) > 0 {
			fixtures = append(fixtures, weekMatches)
		}
	}
	
	return fixtures
}


// ValidateWeek ensures that each week has the correct number of teams playing
func ValidateWeek(week []models.Match, totalTeams int) bool {
	if len(week) != totalTeams/2 {
		return false
	}
	
	// Check that each team plays exactly once in this week
	teamsInWeek := make(map[string]bool)
	for _, match := range week {
		if teamsInWeek[match.HomeTeam] || teamsInWeek[match.AwayTeam] {
			return false // Team already playing in this week
		}
		teamsInWeek[match.HomeTeam] = true
		teamsInWeek[match.AwayTeam] = true
	}
	
	return len(teamsInWeek) == totalTeams
}




func NewGenerateLeague(teams []models.Team)*GenerateLeague {
	engine := GenerateLeague{
		Teams: teams,
		Fixtures: GenerateFixture(teams),
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
	// Base draw chance of 25%, exponentially decreases with strength difference
	drawChance := 0.25 * math.Exp(-strengthDifference/50)
	if drawChance < 0.05 {
		drawChance = 0.05 // Minimum 5%
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
	
	// Update league table with the match result
	league.updateLeagueTable(match)
	
	return match, nil
}

// calculateTeamStrength calculates dynamic team strength based on form and home advantage
func calculateTeamStrength(team models.Team, league *GenerateLeague, isHome bool) float64 {
	baseStrength := float64(team.Strength)
	
	// Home advantage (3% boost)
	if isHome {
		baseStrength *= 1.03
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
    // Base from team strength
    baseGoals := int(teamStrength / 32)
    
    // Exponential impact of strength difference
    // Small differences have minimal impact, large differences have big impact
    diffMultiplier := 1.0 + (strengthDiff / 100) * 0.5
    if diffMultiplier < 0.3 {
        diffMultiplier = 0.3 // Minimum 30% of base
    }
    
    baseGoals = int(float64(baseGoals) * diffMultiplier)
    
    // Add randomness
    totalGoals := baseGoals + rand.Intn(2)
    
    // Range check
    if totalGoals < 0 {
        totalGoals = 0
    } else if totalGoals > 5 {
        totalGoals = 5
    }
    
    return totalGoals
}

// generateDrawScore generates score for a draw
func generateDrawScore(avgStrength float64) int {
	baseGoals := int(avgStrength / 25)
	randomGoals := rand.Intn(1)
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


