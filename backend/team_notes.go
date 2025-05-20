package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/F0RG-2142/capstone-1/backend/internal/auth"
	"github.com/F0RG-2142/capstone-1/backend/internal/database"
	"github.com/google/uuid"
)

// Func to post new team note
func teamNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//get teamID
	teamId, err := uuid.Parse(r.URL.Query().Get("teamID"))
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
	noteId, err := Cfg.db.NewNote(r.Context(), newNoteParams)
	if err != nil {
		http.Error(w, `{"error":"Failed to create note"}`, http.StatusInternalServerError)
		return
	}
	teamNoteParams := database.AddNoteToTeamParams{
		NoteID: noteId,
		ID:     teamId,
		UserID: userId,
	}
	err = Cfg.db.AddNoteToTeam(r.Context(), teamNoteParams)
	if err != nil {
		http.Error(w, `{"error":"Failed to create note"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func getTeamNotes(w http.ResponseWriter, r *http.Request) {
}

func getTeamNote(w http.ResponseWriter, r *http.Request) {
}

func deleteTeamNote(w http.ResponseWriter, r *http.Request) {
}

func updateTeamNote(w http.ResponseWriter, r *http.Request) {
}
