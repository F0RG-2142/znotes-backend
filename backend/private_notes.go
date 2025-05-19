package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/F0RG-2142/capstone-1/backend/internal/auth"
	"github.com/F0RG-2142/capstone-1/backend/internal/database"
	"github.com/google/uuid"
)

func updateNote(w http.ResponseWriter, r *http.Request) {
	//loads note from db, replaces old body with new one. Way to optimise?
	w.Header().Set("Content-Type", "application/json")
	//Get auth get & validate token ->
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(r.URL.Query().Get("noteID")) //loads entire note struct from db
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	getParams := database.GetNoteByIDParams{
		ID:     id,
		UserID: userId,
	}

	note, err := Cfg.db.GetNoteByID(r.Context(), getParams)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	if note.UserID != userId {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	//decode req after auth
	req := struct {
		NoteID uuid.UUID `json:"noteID"`
		Body   string    `json:"body"`
	}{
		NoteID: note.ID,
		Body:   "",
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()

	updateParams := database.UpdateNoteParams{
		ID:   req.NoteID,
		Body: req.Body,
	}
	err = Cfg.db.UpdateNote(r.Context(), updateParams)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(r.URL.Query().Get("noteID"))
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	deleteParams := database.DeleteNoteParams{
		ID:     id,
		UserID: userId,
	}
	err = Cfg.db.DeleteNote(r.Context(), deleteParams)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := uuid.Parse(r.URL.Query().Get("noteID"))
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
	}
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	params := database.GetNoteByIDParams{
		ID:     id,
		UserID: userId,
	}
	note, err := Cfg.db.GetNoteByID(r.Context(), params)
	if err != nil {
		w.WriteHeader(404)
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	}
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

func getNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var notes []database.Note
	id, err := uuid.Parse(r.URL.Query().Get("authorId"))
	if err != nil {
		http.Error(w, "Could not parse uuid", http.StatusBadRequest)
		return
	}
	if id != uuid.Nil {
		notes, err = Cfg.db.GetAllNotes(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	yapsJSON, err := json.Marshal(notes)
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(yapsJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

func notes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
	//get bearer token
	userId, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	if userId != req.UserId {
		w.WriteHeader(http.StatusForbidden)
	}
	//save note to db
	params := database.NewNoteParams{
		Body:   req.Body,
		UserID: req.UserId,
	}
	err = Cfg.db.NewNote(r.Context(), params)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, `{"error":"Failed to create chirp"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
