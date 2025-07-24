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

	
	log.Fatal(http.ListenAndServe(port, nil))
}
