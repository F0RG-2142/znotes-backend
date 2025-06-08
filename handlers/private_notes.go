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

func HandleUpdateNote(w http.ResponseWriter, r *http.Request) {
	//loads note from db, replaces old body with new one. Way to optimise?
	w.Header().Set("Content-Type", "application/json")
	//Get auth get & validate token ->
	userId, err := auth.GetAndValidateToken(r.Header, models.Cfg.Secret)
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

	note, err := models.Cfg.DB.GetNoteByID(r.Context(), getParams)
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
	err = models.Cfg.DB.UpdateNote(r.Context(), updateParams)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func HandleDeleteNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//get and validate token
	userId, err := auth.GetAndValidateToken(r.Header, models.Cfg.Secret)
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
	err = models.Cfg.DB.DeleteNote(r.Context(), deleteParams)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func HandleGetNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := uuid.Parse(r.URL.Query().Get("noteID"))
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
	}
	userId, err := auth.GetAndValidateToken(r.Header, models.Cfg.Secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	params := database.GetNoteByIDParams{
		ID:     id,
		UserID: userId,
	}
	note, err := models.Cfg.DB.GetNoteByID(r.Context(), params)
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

func HandleGetNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var notes []database.Note
	userId, err := auth.GetAndValidateToken(r.Header, models.Cfg.Secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	if userId != uuid.Nil {
		notes, err = models.Cfg.DB.GetAllNotes(r.Context(), userId)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	}
	notesJSON, err := json.Marshal(notes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(notesJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

}

func HandleNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//req struct
	var req struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	//decode req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()
	//get bearer token
	userId, err := auth.GetAndValidateToken(r.Header, models.Cfg.Secret)
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
	_, err = models.Cfg.DB.NewNote(r.Context(), params)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, `{"error":"Failed to create note"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
