package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/F0RG-2142/capstone-1/internal/auth"
	"github.com/F0RG-2142/capstone-1/internal/database"
	"github.com/F0RG-2142/capstone-1/models"
	"github.com/google/uuid"
)

// Func to post new team note and needs the following params:
//
//	{
//		"body":"string"
//		"user_id":"string"
//	};
func HandleTeamNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, models.Cfg.Secret)
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
	//req struct
	var req struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	//decode req
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		w.WriteHeader(500)
		return
	}
	//Create new note
	newNoteParams := database.NewNoteParams{
		Body:   req.Body,
		UserID: userId,
	}
	noteId, err := models.Cfg.DB.NewNote(r.Context(), newNoteParams)
	if err != nil {
		http.Error(w, `{"error":"Failed to create note"}`, http.StatusInternalServerError)
		return
	}
	teamNoteParams := database.AddNoteToTeamParams{
		NoteID: noteId,
		ID:     teamId,
		UserID: userId,
	}
	err = models.Cfg.DB.AddNoteToTeam(r.Context(), teamNoteParams)
	if err != nil {
		http.Error(w, `{"error":"Failed to create note"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// Func to get specific team note.
// Returns:
//
//	{
//		{
//			"note_id":"uuid"
//			"created_at":"timestamp"
//			"updated_at":"timestamp"
//			"body":"string"
//			"user_id":"uuid"
//		}
//	...
//
// }
func HandleGetTeamNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, models.Cfg.Secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get team id
	teamId, err := uuid.Parse(r.URL.Query().Get("team_id"))
	if err != nil {
		http.Error(w, "Could not parse team uuid", http.StatusBadRequest)
		return
	}
	getTeamNotesParams := database.GetTeamNotesParams{
		TeamID: teamId,
		UserID: userId,
	}
	notes, err := models.Cfg.DB.GetTeamNotes(r.Context(), getTeamNotesParams)
	if err != nil {
		http.Error(w, "Could not get notes, please reload", http.StatusFailedDependency)
		return
	}
	notesJSON, err := json.Marshal(notes)
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(notesJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

// Func to get specific team note.
// Returns:
//
//	{
//		"note_id":"uuid"
//		"created_at":"timestamp"
//		"updated_at":"timestamp"
//		"body":"string"
//		"user_id":"uuid"
//	}
func HandleGetTeamNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, models.Cfg.Secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get team id
	teamId, err := uuid.Parse(r.URL.Query().Get("team_id"))
	if err != nil {
		http.Error(w, "Could not parse team uuid", http.StatusBadRequest)
		return
	}
	//get note id
	noteId, err := uuid.Parse(r.URL.Query().Get("note_id"))
	if err != nil {
		http.Error(w, "Could not parse note uuid", http.StatusBadRequest)
		return
	}
	//Get team note
	getTeamNoteParams := database.GetTeamNoteParams{
		ID:     noteId,
		TeamID: teamId,
		UserID: userId,
	}
	note, err := models.Cfg.DB.GetTeamNote(r.Context(), getTeamNoteParams)
	if err != nil {
		http.Error(w, "Could note get note, please reload", http.StatusBadRequest)
		return
	}
	//marshal note to json
	noteJSON, err := json.Marshal(note)
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(noteJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

// Deletes the specified note from the team and database
func HandleDeleteTeamNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, models.Cfg.Secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get note id
	noteId, err := uuid.Parse(r.URL.Query().Get("note_id"))
	if err != nil {
		http.Error(w, "Could not parse note uuid", http.StatusBadRequest)
		return
	}
	//Get team note
	removeNoteFromTeamParams := database.RemoveNoteFromTeamParams{
		NoteID: noteId,
		UserID: userId,
	}
	err = models.Cfg.DB.RemoveNoteFromTeam(r.Context(), removeNoteFromTeamParams)
	if err != nil {
		http.Error(w, "Could note delete note, please try again", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Update note body and needs following params:
//
//	{
//		"body":"string"
//	}
//
// Returns:
//
//	{
//		"note_id":"uuid"
//		"created_at":"timestamp"
//		"updated_at":"timestamp"
//		"body":"string"
//		"user_id":"uuid"
//	}
func HandleUpdateTeamNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, models.Cfg.Secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get team id
	teamId, err := uuid.Parse(r.URL.Query().Get("team_id"))
	if err != nil {
		http.Error(w, "Could not parse team uuid", http.StatusBadRequest)
		return
	}
	//get note id
	noteId, err := uuid.Parse(r.URL.Query().Get("note_id"))
	if err != nil {
		http.Error(w, "Could not parse note uuid", http.StatusBadRequest)
		return
	}
	//req struct
	var req struct {
		Body string `json:"body"`
	}
	//decode req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()
	updateTeamNoteParams := database.UpdateTeamNoteParams{
		Body:   req.Body,
		ID:     noteId,
		TeamID: teamId,
		UserID: userId,
	}
	err = models.Cfg.DB.UpdateTeamNote(r.Context(), updateTeamNoteParams)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
