package main

import (
	"backend/db"
	"backend/managers/server"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

var database *db.DB

///////////////////////////////////////
//main initializes the database, sets up HTTP routes, and starts the server.
///////////////////////////////////////

func main() {
	var err error
	database, err = db.NewDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	serverManager := server.GetServerManager()

	http.HandleFunc("/api/leaderboard", handleLeaderboard)
	http.HandleFunc("/api/player", handlePlayer)
	http.HandleFunc("/api/test/update-stats", handleTestUpdateStats)

	log.Println("Server started on :8080")
	serverManager.StartServer()
}

// /////////////////////////////////////
// handleLeaderboard handles the leaderboard API endpoint.
// /////////////////////////////////////

func handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 10
		}
	}

	players, err := database.GetLeaderboard(limit)
	if err != nil {
		log.Printf("Error getting leaderboard: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(players); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

///////////////////////////////////////
// handlePlayer handles the player API endpoint.
///////////////////////////////////////

func handlePlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username parameter is required", http.StatusBadRequest)
		return
	}

	player, err := database.GetPlayerByUsername(username)
	if err != nil {
		player, err = database.CreateOrUpdatePlayer(username)
		if err != nil {
			log.Printf("Error creating player: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(player); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

///////////////////////////////////////
// handleTestUpdateStats is a test endpoint to manually update player stats.
///////////////////////////////////////

func handleTestUpdateStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	winner := r.URL.Query().Get("winner")
	loser := r.URL.Query().Get("loser")

	if winner == "" || loser == "" {
		http.Error(w, "Winner and loser parameters are required", http.StatusBadRequest)
		return
	}

	_, err := database.CreateOrUpdatePlayer(winner)
	if err != nil {
		log.Printf("Error creating winner: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_, err = database.CreateOrUpdatePlayer(loser)
	if err != nil {
		log.Printf("Error creating loser: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = database.UpdateGameResult(winner, loser)
	if err != nil {
		log.Printf("Error updating game result: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	winnerPlayer, err := database.GetPlayerByUsername(winner)
	if err != nil {
		log.Printf("Error getting winner: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	loserPlayer, err := database.GetPlayerByUsername(loser)
	if err != nil {
		log.Printf("Error getting loser: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"winner":  winnerPlayer,
		"loser":   loserPlayer,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
