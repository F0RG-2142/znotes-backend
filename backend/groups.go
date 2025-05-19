package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/F0RG-2142/capstone-1/backend/internal/auth"
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
	w.Header().Set("Content-Type", "application/json")
	id, err := uuid.Parse(r.URL.Query().Get("teamID"))
	if err != nil {
		http.Error(w, "Could not parse uuid", http.StatusBadRequest)
		return
	}
	//get and validate token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	userId, err := auth.ValidateJWT(token, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}
	var team database.Team
	//add auth check to team call
	params := database.GetTeamByIdParams{
		UserID: userId,
		TeamID: id,
	}
	if id != uuid.Nil {
		team, err = Cfg.db.GetTeamById(r.Context(), params)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	//check return value to see if it returns valid json as there is no json tag in database.Team
	teamJSON, err := json.Marshal(team)
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(teamJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

func teams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var teams []database.Team
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	userId, err := auth.ValidateJWT(token, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}
	teams, err = Cfg.db.GetAllTeams(r.Context(), userId)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}
	teamsJSON, err := json.Marshal(teams)
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(teamsJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

func deleteTeam(w http.ResponseWriter, r *http.Request) {
}

func addUserToTeam(w http.ResponseWriter, r *http.Request) {
}

func removeUserFromTeam(w http.ResponseWriter, r *http.Request) {
}

func getTeamMembers(w http.ResponseWriter, r *http.Request) {
}
