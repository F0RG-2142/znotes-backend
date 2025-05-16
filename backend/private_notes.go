package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/F0RG-2142/capstone-1/backend/internal/auth"
	"github.com/F0RG-2142/capstone-1/backend/internal/database"
	"github.com/google/uuid"
)

func updateNote(w http.ResponseWriter, r *http.Request) {
	//Do it
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
	id, err := uuid.Parse(r.URL.Query().Get("yapId"))
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	yap, err := Cfg.db.GetNoteByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	if yap.UserID != user_id {
		http.Error(w, "This is not your yap", http.StatusForbidden)
		return
	}
	err = Cfg.db.DeleteNote(r.Context(), yap.ID)
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
	yap, err := Cfg.db.GetNoteByID(r.Context(), uuid.UUID(id))
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(err.Error()))
	}
	yapJSON, err := json.Marshal(yap)
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(yapJSON)
}

func getNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var yaps []database.Note
	id, err := uuid.Parse(r.URL.Query().Get("authorId"))
	if err != nil {
		http.Error(w, "Could not parse uuid", http.StatusBadRequest)
		return
	}
	if id != uuid.Nil {
		yaps, err = Cfg.db.GetNotesByAuthor(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	} else {
		yaps, err = Cfg.db.GetAllNotes(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	}

	yapsJSON, err := json.Marshal(yaps)
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

	//If body too long (>140) return error
	if len(req.Body) > 140 {
		respBody := returnValues{
			Err: "Chirp is too long",
		}
		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write(data)
		return
	}
	//Clean profanities 1.0
	cleaned_body := req.Body
	cleaned_body = strings.Replace(cleaned_body, "Lol", "****", -1)
	cleaned_body = strings.Replace(cleaned_body, "fortnite", "****", -1)
	cleaned_body = strings.Replace(cleaned_body, "damn", "****", -1)
	//save chirp to db
	params := database.NewNoteParams{
		Body:   cleaned_body,
		UserID: req.UserId,
	}
	chirp, err := Cfg.db.NewNote(r.Context(), params)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, `{"error":"Failed to create chirp"}`, http.StatusInternalServerError)
		return
	}

	respBody := returnValues{
		Id:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID.String(),
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
