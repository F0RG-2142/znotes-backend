package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/F0RG-2142/capstone-1/handlers"
	"github.com/F0RG-2142/capstone-1/internal/auth"
	"github.com/F0RG-2142/capstone-1/internal/database"
	"github.com/F0RG-2142/capstone-1/models"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load env:", err)
		return
	}
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal("Failed to connect to db:", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	queries := database.New(db)
	models.Cfg.DB = queries
	models.Cfg.Platform = os.Getenv("PLATFORM")
	models.Cfg.Secret = os.Getenv("JWT_SECRET")

	mux := http.NewServeMux()
	//Utility and admin
	mux.Handle("GET /api/v1/healthz", http.HandlerFunc(readiness))         //Check if server is ready //Done
	mux.Handle("GET /api/v1/admin/metrics", http.HandlerFunc(metrics))     //Server metrics endpoint //---
	mux.Handle("POST /api/v1/payment/webhooks", http.HandlerFunc(payment)) //Payment platform webhook //---
	//Users and auth
	mux.Handle("POST /api/v1/register", Chain(http.HandlerFunc(handlers.HandleNewUser)))          //New User Registration
	mux.Handle("POST /api/v1/login", Chain(http.HandlerFunc(handlers.HandleLogin)))               //Login to profile
	mux.Handle("POST /api/v1/logout", Chain(http.HandlerFunc(handlers.HandleRevokeRefreshToken))) //Revoke refresh tok
	mux.Handle("POST /api/v1/token/refresh", Chain(http.HandlerFunc(handlers.HandleRefreshJWT)))  //Refresh JWT
	mux.Handle("PUT /api/v1/user/me", Chain(http.HandlerFunc(handlers.HandleUpdateUser)))         //Update user details
	//Private Notes
	mux.Handle("POST /api/v1/notes", Chain(http.HandlerFunc(handlers.HandleNotes)))                 //Post Private Note //Done
	mux.Handle("GET /api/v1/notes", Chain(http.HandlerFunc(handlers.HandleGetNotes)))               //Get all private notes //Done
	mux.Handle("GET /api/v1/notes/{noteID}", Chain(http.HandlerFunc(handlers.HandleGetNote)))       //Get one private note //Done
	mux.Handle("PUT /api/v1/notes/{noteID}", Chain(http.HandlerFunc(handlers.HandleUpdateNote)))    //Update private note //Done
	mux.Handle("DELETE /api/v1/notes/{noteID}", Chain(http.HandlerFunc(handlers.HandleDeleteNote))) //Delete note based on id //Done
	//Teams
	mux.Handle("POST /api/v1/teams", Chain(http.HandlerFunc(handlers.HandleNewTeam)))                                          //Create new team
	mux.Handle("GET /api/v1/teams", Chain(http.HandlerFunc(handlers.HandleGetTeams)))                                          //List all teams a user is part of
	mux.Handle("GET /api/v1/teams/{teamID}", Chain(http.HandlerFunc(handlers.HandleGetTeam)))                                  //Get specific team details
	mux.Handle("DELETE /api/v1/teams/{teamID}", Chain(http.HandlerFunc(handlers.HandleDeleteTeam)))                            //Delete team
	mux.Handle("POST /api/v1/teams/{teamID}/members", Chain(http.HandlerFunc(handlers.HandleAddUserToTeam)))                   //Add new user to team
	mux.Handle("DELETE /api/v1/teams/{teamID}/members/{memberID}", Chain(http.HandlerFunc(handlers.HandleRemoveUserFromTeam))) //Remove user from team
	mux.Handle("GET /api/v1/teams/{teamID}/members", Chain(http.HandlerFunc(handlers.HandleGetTeamMembers)))                   //Get all users in team
	//Team Notes
	mux.Handle("POST /api/v1/teams/{teamID}/notes", Chain(http.HandlerFunc(handlers.HandleTeamNotes)))                 //Post team Note
	mux.Handle("GET /api/v1/teams/{teamID}/notes", Chain(http.HandlerFunc(handlers.HandleGetTeamNotes)))               //Get all team notes
	mux.Handle("GET /api/v1/teams/{teamID}/notes/{noteID}", Chain(http.HandlerFunc(handlers.HandleGetTeamNote)))       //Get one team note
	mux.Handle("PUT /api/v1/teams/{teamID}/notes/{noteID}", Chain(http.HandlerFunc(handlers.HandleUpdateTeamNote)))    //Update team Note
	mux.Handle("DELETE /api/v1/teams/{teamID}/notes/{noteID}", Chain(http.HandlerFunc(handlers.HandleDeleteTeamNote))) //Delete team note based on id

	fmt.Println("Listening on http://localhost:8080/")
	if err = http.ListenAndServe(":8080", corsMiddleware(mux)); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func readiness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Server is good to go"))
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
}

func metrics(w http.ResponseWriter, r *http.Request) {
}

func payment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusFailedDependency)
	}
	if apiKey != os.Getenv("PP_KEY") {
		http.Error(w, "Unauthorized Endpoint", http.StatusUnauthorized)
	}
	req := struct {
		Event string `json:"event"`
		Data  struct {
			UserId uuid.UUID `json:"user_id"`
		} `json:"data"`
	}{
		Event: "",
		Data: struct {
			UserId uuid.UUID `json:"user_id"`
		}{
			UserId: uuid.Nil,
		},
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Event != "user.upgraded" {
		http.Error(w, "", http.StatusNoContent)
	}

	err = models.Cfg.DB.GivePremium(r.Context(), req.Data.UserId)
	if err != nil {
		http.Error(w, "User Not Found", http.StatusNotFound)
	}
	w.WriteHeader(http.StatusNoContent)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if strings.HasPrefix(strings.ToLower(origin), "http://localhost") ||
			strings.HasPrefix(strings.ToLower(origin), "https://localhost") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Chain(h http.Handler, middlewares ...models.Middleware) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}
