package main

import (
	"encoding/json"
	handlers "insider-league/Handlers"
	"net/http"
)

func SetupRoutes() {
	// Create league handler
	leagueHandler := handlers.NewLeagueHandler()
		http.HandleFunc("/api/league", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			leagueHandler.CreateLeague(w, r)
		case http.MethodDelete:
			leagueHandler.ClearLeague(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	http.HandleFunc("/api/league/play-week", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			leagueHandler.PlayWeek(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	http.HandleFunc("/api/league/table", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			leagueHandler.GetLeagueTable(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	http.HandleFunc("/api/league/matches", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			leagueHandler.GetMatches(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	http.HandleFunc("/api/league/status", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			leagueHandler.GetLeagueStatus(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	// Health endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			
			healthStatus := map[string]interface{}{
				"status":    "OK",
				"message":   "Football League API is running",
				"timestamp": "2024-01-15T10:30:00Z",
				"version":   "1.0.0",
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			
			response, _ := json.Marshal(healthStatus)
			w.Write(response)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
