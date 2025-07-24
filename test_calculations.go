package main

import (
	"fmt"
	"insider-league/Models"
	"insider-league/Services"
)

func main() {
	// Test with different team counts
	testCases := []int{2, 3, 4, 5, 6, 7, 8, 9, 10}
	
	fmt.Println("Comparing CalculateTotalWeeks vs GenerateFixture results:")
	fmt.Println("========================================================")
	
	for _, teamCount := range testCases {
		// Create test teams
		teams := make([]models.Team, teamCount)
		for i := 0; i < teamCount; i++ {
			teams[i] = models.Team{
				ID:   i + 1,
				Name: fmt.Sprintf("Team%d", i+1),
			}
		}
		
		// Calculate using CalculateTotalWeeks
		calculatedWeeks := Services.CalculateTotalWeeks(teamCount)
		
		// Calculate using GenerateFixture
		fixtures := Services.GenerateFixture(teams)
		actualWeeks := len(fixtures)
		
		// Compare results
		match := calculatedWeeks == actualWeeks
		status := "✓ MATCH"
		if !match {
			status = "✗ DIFFERENT"
		}
		
		fmt.Printf("Teams: %d | Calculated: %d | Actual: %d | %s\n", 
			teamCount, calculatedWeeks, actualWeeks, status)
		
		if !match {
			fmt.Printf("  Calculated formula: %d * (%d-1) / %d = %d\n", 
				teamCount, teamCount, teamCount/2, calculatedWeeks)
			fmt.Printf("  Actual fixtures generated: %d weeks\n", actualWeeks)
		}
	}
} 