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
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
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
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
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
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get team id
	teamId, err := uuid.Parse(r.URL.Query().Get("teamID"))
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	deleteParams := database.DeleteTeamParams{
		UserID: userId,
		ID:     teamId,
	}
	err = Cfg.db.DeleteTeam(r.Context(), deleteParams)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func addUserToTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//decode req to add user and what their role should be
	var req struct {
		UserID uuid.UUID `json:"userID"`
		TeamID uuid.UUID `json:"teamID"`
		Role   string    `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//Set the parameters for what team member to get to see if they are authorized to add someone (be admin on specified team)
	getMemberParams := database.GetTeamMemberParams{
		UserID: userId,
		TeamID: req.TeamID,
	}
	var member database.UserTeam
	if member, err = Cfg.db.GetTeamMember(r.Context(), getMemberParams); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	if member.Role != "admin" {
		http.Error(w, `{"error":"You are not authorized to add people to this group"}`, http.StatusBadRequest)
		return
	}
	//add user to team
	addParams := database.AddToTeamParams{
		UserID: req.UserID,
		TeamID: req.TeamID,
		Role:   req.Role,
	}
	err = Cfg.db.AddToTeam(r.Context(), addParams)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusNoContent)
}

func removeUserFromTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//decode req to add user and what their role should be
	var req struct {
		UserID uuid.UUID `json:"userID"`
		TeamID uuid.UUID `json:"teamID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//See if requester is authorized to remove someone (must be admin on specified team)
	getMemberParams := database.GetTeamMemberParams{
		UserID: userId,
		TeamID: req.TeamID,
	}
	var member database.UserTeam
	if member, err = Cfg.db.GetTeamMember(r.Context(), getMemberParams); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	if member.Role != "admin" {
		http.Error(w, `{"error":"You are not authorized to add people to this group"}`, http.StatusBadRequest)
		return
	}
	//remove user
	removeUserParams := database.RemoveUserParams{
		UserID: userId,
		TeamID: req.TeamID,
	}
	if err = Cfg.db.RemoveUser(r.Context(), removeUserParams); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusNoContent)
}

func getTeamMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//decode req to add user and what their role should be
	var req struct {
		TeamID uuid.UUID `json:"teamID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//See if requester is authorized to view this team
	getMemberParams := database.GetTeamMemberParams{
		UserID: userId,
		TeamID: req.TeamID,
	}
	//Doesnt need to make member var as we just need to  see if they are in the team, anyone in a team can view members
	if _, err = Cfg.db.GetTeamMember(r.Context(), getMemberParams); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get all members
	var members []database.Team
	if members, err = Cfg.db.GetTeamMembers(r.Context(), req.TeamID); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
	}

	membersJSON, err := json.Marshal(members)
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(membersJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}
