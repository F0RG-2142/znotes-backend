package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/F0RG-2142/capstone-1/backend/internal/auth"
	"github.com/F0RG-2142/capstone-1/backend/internal/database"
	"github.com/google/uuid"
)

func updateNote(w http.ResponseWriter, r *http.Request) {
	//loads note from db, replaces old body with new one. Way to optimise?
	w.Header().Set("Content-Type", "application/json")
	//Get auth get & validate token ->
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	user_id, err := auth.ValidateJWT(token, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}
	id, err := uuid.Parse(r.URL.Query().Get("noteId")) //loads entire note struct from db
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	note, err := Cfg.db.GetNoteByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	if note.UserID != user_id {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	//decode req after auth
	req := struct {
		NoteID uuid.UUID `json:"noteId"`
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

	params := database.UpdateNoteParams{
		ID:   req.NoteID,
		Body: req.Body,
	}
	err = Cfg.db.UpdateNote(r.Context(), params)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//get JWT
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	//validate token
	user_id, err := auth.ValidateJWT(token, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}
	id, err := uuid.Parse(r.URL.Query().Get("noteId"))
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	note, err := Cfg.db.GetNoteByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	if note.UserID != user_id {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusForbidden)
		return
	}
	err = Cfg.db.DeleteNote(r.Context(), note.ID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := uuid.Parse(r.URL.Query().Get("yapId"))
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
	}
	note, err := Cfg.db.GetNoteByID(r.Context(), uuid.UUID(id))
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
	}
	noteJSON, err := json.Marshal(note)
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(noteJSON)
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
		notes, err = Cfg.db.GetNotesByAuthor(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	} else {
		notes, err = Cfg.db.GetAllNotes(r.Context())
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
	w.Write(yapsJSON)
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
	//response struct
	type returnValues struct {
		Id        string    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    string    `json:"user_id"`
		Err       string    `json:"error"`
		Valid     bool      `json:"valid"`
	}
	//get bearer token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
	}
	//validate token
	user_id, err := auth.ValidateJWT(token, Cfg.secret)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
	}
	if user_id != req.UserId {
		w.WriteHeader(http.StatusForbidden)
	}
	//save note to db
	params := database.NewNoteParams{
		Body:   req.Body,
		UserID: req.UserId,
	}
	note, err := Cfg.db.NewNote(r.Context(), params)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, `{"error":"Failed to create chirp"}`, http.StatusInternalServerError)
		return
	}

	respBody := returnValues{
		Id:        note.ID.String(),
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
		Body:      note.Body,
		UserId:    note.UserID.String(),
		Valid:     true,
	}
	//marshal and send reponse on successful creation
	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
