package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/F0RG-2142/capstone-1/internal/auth"
	"github.com/F0RG-2142/capstone-1/internal/database"
	"github.com/google/uuid"
)

// Updatse user username and/or password, needs theese parameters:
//
//	{
//			"email":"string"
//			"password":"string"
//	};
//
// and returns:
//
//	{
//		"user_id":"uuid"
//		"created_at":"timestamp"
//		"updated_at":"timsetamp"
//		"user_email":"string"
//		"has_notes_premium":"bool"
//	}
func updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//get and validate auth token
	user_id, err := auth.GetAndValidateToken(r.Header, Cfg.secret)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}
	//decode request
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	//hash passw and update user
	hashed_pass, err := auth.HashPassword(req.Password)
	if err != nil {

	}
	params := database.UpdateUserParams{
		Email:          req.Email,
		HashedPassword: hashed_pass,
		ID:             user_id,
	}
	err = Cfg.db.UpdateUser(r.Context(), params)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
		return
	}
	//get updated user
	user, err := Cfg.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
		return
	}
	//create response struct, marshal, and respond
	resp := database.User{
		ID:              user.ID,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		Email:           user.Email,
		HasNotesPremium: user.HasNotesPremium,
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, `{"error":"Failed to create response"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResp)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

// Needs JWT in Authorization header
//
// Refreshes the JWT. This will be called every 55 mins by the client as the JWT expires every hour
//
// Returns:
//
//	{
//		"token":"string"
//	}
func refreshJWT(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}
	refreshToken, err := Cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("Error fetching refresh token: %v", err)
		http.Error(w, `{"error":"Invalid refresh token"}`, http.StatusForbidden)
		return
	}
	if !refreshToken.RevokedAt.Valid {
		http.Error(w, `{"error":"Refresh token is revoked"}`, http.StatusForbidden)
		return
	}
	if time.Now().After(refreshToken.ExpiresAt) {
		http.Error(w, `{"error":"Refresh token is expired"}`, http.StatusForbidden)
		return
	}
	tokenSecret := Cfg.secret
	if tokenSecret == "" {
		log.Println("JWT_SECRET not set")
		http.Error(w, `{"error":"Server configuration error"}`, http.StatusInternalServerError)
		return
	}
	accessToken, err := auth.MakeJWT(refreshToken.UserID, tokenSecret, time.Hour)
	if err != nil {
		log.Printf("Error generating access token: %v", err)
		http.Error(w, `{"error":"Failed to generate access token"}`, http.StatusInternalServerError)
		return
	}
	resp := struct {
		AccessToken string `json:"token"`
	}{
		AccessToken: accessToken,
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		http.Error(w, `{"error":"Failed to create response"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResp)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

// Creatse a new user and needs the following params:
//
//	{
//		"user_email":"string"
//		"user_password":"string"
//	}
//
// and returns:
//
//	{
//		"user_id":"uuid"
//		"created_at":"timestamp"
//		"updated_at":"timsetamp"
//		"user_email":"string"
//		"has_notes_premium":"bool"
//	}
func newUser(w http.ResponseWriter, r *http.Request) {
	//decode request body
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		Email    string `json:"user_email"`
		Password string `json:"user_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//validate email and password
	if req.Email == "" {
		http.Error(w, `{"error":"Email is required"}`, http.StatusBadRequest)
		return
	}
	if req.Password == "" {
		http.Error(w, `{"error":"Password is required"}`, http.StatusBadRequest)
		return
	}
	//hash passw
	hashedPass, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"Failed to hash password"}`, http.StatusFailedDependency)
	}
	//Create user and resepond with created user
	params := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPass,
	}

	user, err := Cfg.db.CreateUser(r.Context(), params)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, `{"error":"Failed to create user"}`, http.StatusFailedDependency)
		return
	}
	resp := database.User{
		ID:              user.ID,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		Email:           user.Email,
		HasNotesPremium: user.HasNotesPremium,
	}
	userJSON, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling user to JSON: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusFailedDependency)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(userJSON)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

// Log into specified account and needs the following params:
//
//	{
//		"user_email":"string"
//		"user_password":"string"
//	}
//
// returns the following:
//
//	{
//		"id":"uuid"
//		"created_at":"timestamp"
//		"updated_at":"timestamp"
//		"email":"string"
//		"token":"string"
//		"refresh_token":"string"
//		"has_notes_premium":"bool"
//	}
func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//parse req
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	//verify usern and passw
	user, err := Cfg.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, `{"error":"Incorrect username or password"}`, http.StatusBadRequest)
	}
	err = auth.CheckPasswordHash(user.HashedPassword, req.Password)
	if err != nil {

		http.Error(w, `{"error":"Incorrect username or password"}`, http.StatusBadRequest)
	}
	//make jwt
	Token, err := auth.MakeJWT(user.ID, Cfg.secret, time.Hour)
	if err != nil {
		log.Printf("Error generating JWT for user %q: %v", user.ID, err)
		http.Error(w, `{"error":"Failed to generate access token"}`, http.StatusInternalServerError)
		return
	}
	refreshToken, _ := auth.MakeRefreshToken()
	params := database.NewRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	}
	_, err = Cfg.db.NewRefreshToken(r.Context(), params) //Wat gaan hier aan???
	if err != nil {
		w.WriteHeader(http.StatusFailedDependency)
		return
	}
	resp := struct {
		ID                uuid.UUID `json:"id"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
		Email             string    `json:"email"`
		Token             string    `json:"token"`
		RefreshToken      string    `json:"refresh_token"`
		Has_notes_premium bool      `json:"has_notes_premium"`
	}{
		ID:                user.ID,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
		Email:             user.Email,
		Token:             Token,
		RefreshToken:      refreshToken,
		Has_notes_premium: user.HasNotesPremium,
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, `{"error":"Failed to create response"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResp)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

// Revoke the refresh token from a user. Needs token in auth header to authorize
func revokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
		return
	}
	refreshToken, err := Cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("Error fetching refresh token: %v", err)
		http.Error(w, `{"error":"Invalid refresh token"}`, http.StatusForbidden)
		return
	}
	err = Cfg.db.RevokeRefreshToken(r.Context(), refreshToken.Token)
	if err != nil {
		http.Error(w, `"error":"Could not revoke Refresh Token"`, http.StatusFailedDependency)
	}
	w.WriteHeader(http.StatusNoContent)
}
