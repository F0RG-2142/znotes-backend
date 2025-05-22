package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/F0RG-2142/capstone-1/internal/auth"
	"github.com/F0RG-2142/capstone-1/internal/database"
	"github.com/google/uuid"
)

// Creates a new team using the following parameters:
//
//	{
//		"team_name":"string",
//		"user_id":"uuid",
//		"is_private":"bool",
//	}
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

// Gets one team from database bassed on team id given in url and returns:
//
//	{
//	   "team_id":"uuid"
//	   "created_at":"timestamp"
//	   "updated_at" "timestamp"
//	   "team_name":"string"
//	   "created_by":"uuid"
//	   "is_private":"bool"
//	}
func team(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	teamId, err := uuid.Parse(r.URL.Query().Get("team_id"))
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
	params := database.GetTeamByIdParams{
		UserID: userId,
		TeamID: teamId,
	}
	team, err = Cfg.db.GetTeamById(r.Context(), params)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
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

// Gets all teams from database and returns:
//
//	{
//		{
//		   "team_id":"uuid"
//		   "created_at":"timestamp"
//		   "updated_at" "timestamp"
//		   "team_name":"string"
//		   "created_by":"uuid"
//		   "is_private":"bool"
//		}
//
// ...
// }
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

// Deletes team from database based on team id given in url
func deleteTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get team id
	teamId, err := uuid.Parse(r.URL.Query().Get("team_id"))
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

// Adds user to team in database based on url provided and needs the following parameters:
//
//	{
//		"user_id":"uuid"
//		"role":"string"
//	}
func addUserToTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get teamID
	teamId, err := uuid.Parse(r.URL.Query().Get("team_id"))
	if err != nil {
		http.Error(w, "Could not parse uuid", http.StatusBadRequest)
		return
	}
	//decode req to add user and what their role should be
	var req struct {
		UserID uuid.UUID `json:"user_id"`
		Role   string    `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//Set the parameters for what team member to get to see if they are authorized to add someone (be admin on specified team)
	getMemberParams := database.GetTeamMemberParams{
		UserID: userId,
		TeamID: teamId,
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
	addParams := database.AddUserToTeamParams{
		UserID: req.UserID,
		TeamID: teamId,
		Role:   req.Role,
	}
	err = Cfg.db.AddUserToTeam(r.Context(), addParams)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusNoContent)
}

// Remove a user from the team
func removeUserFromTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//get team and member id
	teamId, err := uuid.Parse(r.URL.Query().Get("team_id"))
	if err != nil {
		http.Error(w, "Could not parse uuid", http.StatusBadRequest)
		return
	}
	memberId, err := uuid.Parse(r.URL.Query().Get("memberID"))
	if err != nil {
		http.Error(w, "Could not parse uuid", http.StatusBadRequest)
		return
	}

	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//See if requester is authorized to remove someone (must be admin on specified team)
	getMemberParams := database.GetTeamMemberParams{
		UserID: userId,
		TeamID: teamId,
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
	//remove specified member
	removeUserParams := database.RemoveUserFromTeamParams{
		UserID: memberId,
		TeamID: teamId,
	}
	if err = Cfg.db.RemoveUserFromTeam(r.Context(), removeUserParams); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Get all the members of a specified group
func getTeamMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get team ID
	teamId, err := uuid.Parse(r.URL.Query().Get("team_id"))
	if err != nil {
		http.Error(w, "Could not parse uuid", http.StatusBadRequest)
		return
	}
	//See if requester is authorized to view this team
	getMemberParams := database.GetTeamMemberParams{
		UserID: userId,
		TeamID: teamId,
	}
	//Doesnt need to make member var as we just need to  see if they are in the team, anyone in a team can view members
	if _, err = Cfg.db.GetTeamMember(r.Context(), getMemberParams); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get all members
	var members []database.Team
	if members, err = Cfg.db.GetTeamMembers(r.Context(), teamId); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
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
