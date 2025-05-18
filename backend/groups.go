package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/F0RG-2142/capstone-1/backend/internal/database"
	"github.com/google/uuid"
)

func newTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//req struct and decoding
	var req struct {
		TeamName  string    `json:"team_name"`
		UserId    uuid.UUID `json:"user_id"`
		IsPrivate bool      `json:"is_private"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()
	if req.TeamName == "" {
		http.Error(w, `{"error":"Please enter a team name"}`, http.StatusNotAcceptable)
		return
	}
	params := database.NewTeamParams{
		TeamName:  req.TeamName,
		CreatedBy: req.UserId,
		IsPrivate: req.IsPrivate,
	}
	err := Cfg.db.NewTeam(r.Context(), params)
	if err != nil {
		log.Printf("Error creating team: %v", err)
		http.Error(w, `{"error":"Failed to create team"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func team(w http.ResponseWriter, r *http.Request) {
}

func teams(w http.ResponseWriter, r *http.Request) {
}

func deleteTeam(w http.ResponseWriter, r *http.Request) {
}

func addUserToTeam(w http.ResponseWriter, r *http.Request) {
}

func removeUserFromTeam(w http.ResponseWriter, r *http.Request) {
}

func getTeamMembers(w http.ResponseWriter, r *http.Request) {
}
