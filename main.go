package main

import (
	"fmt"
	"log"
	"net/http"
	"math/rand"
	"time"
	"insider-league/database"
)

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
	
	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	
	fmt.Println("Database connected successfully!")
	
	// Setup routes
	SetupRoutes()
	
	// Start server
	port := ":8080"
	fmt.Println(" Football League API Server starting on port %s\n", port)
	fmt.Println(" Available endpoints:")
	fmt.Println("  POST   /api/league              - Create a new league")
	fmt.Println("  POST   /api/league/play-week    - Play next week")
	fmt.Println("  GET    /api/league/table        - Get league standings")
	fmt.Println("  GET    /api/league/matches      - Get all match results")
	fmt.Println("  GET    /api/league/matches/week/{week} - Get matches for specific week")
	fmt.Println("  GET    /api/league/status       - Get league status")
	fmt.Println("  GET    /api/health              - Health check")
	
	log.Fatal(http.ListenAndServe(port, nil))
}