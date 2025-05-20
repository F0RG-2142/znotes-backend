package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/F0RG-2142/capstone-1/backend/internal/auth"
	"github.com/F0RG-2142/capstone-1/backend/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db       *database.Queries
	platform string
	secret   string
}

var Cfg apiConfig

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load env:", err)
		return
	}
	dbURL := os.Getenv("DB_URL")
	Cfg.platform = os.Getenv("PLATFORM")
	Cfg.secret = os.Getenv("JWT_SECRET")

	db, _ := sql.Open("postgres", dbURL)
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	mux := http.NewServeMux()
	//Utility and admin
	mux.Handle("/", http.FileServer(http.Dir("../app/")))
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("../app/")))) // Serve app landing page //Done
	mux.Handle("GET /api/v1/healthz", http.HandlerFunc(readiness))                       //Check if server is ready //Done
	mux.Handle("GET /api/v1/admin/metrics", http.HandlerFunc(metrics))                   //Server metrics endpoint //---
	mux.Handle("POST /api/v1/payment/webhooks", http.HandlerFunc(payment))               //Payment platform webhook //---
	//Users and auth
	mux.Handle("POST /api/v1/register", http.HandlerFunc(newUser))      //New User Registration //Done
	mux.Handle("POST /api/v1/login", http.HandlerFunc(login))           //Login to profile //Done
	mux.Handle("POST /api/v1/logout", http.HandlerFunc(revoke))         //Revoke refresh token Done
	mux.Handle("POST /api/v1/token/refresh", http.HandlerFunc(refresh)) //Refresh JWT //Done
	mux.Handle("GET /api/v1/user/me", http.HandlerFunc(refresh))        //Get current user login details //Done
	mux.Handle("PUT /api/v1/user/me", http.HandlerFunc(updateUser))     //Update user details //Done
	//Private Notes
	mux.Handle("POST /api/v1/notes", http.HandlerFunc(notes))                 //Post Private Note //Done
	mux.Handle("GET /api/v1/notes", http.HandlerFunc(getNotes))               //Get all private notes //Done
	mux.Handle("GET /api/v1/notes/{noteID}", http.HandlerFunc(getNote))       //Get one private note //Done
	mux.Handle("PUT /api/v1/notes/{noteID}", http.HandlerFunc(updateNote))    //Update private note //Done
	mux.Handle("DELETE /api/{userID}/{noteID}", http.HandlerFunc(deleteNote)) //Delete note based on id //Done
	//Teams
	mux.Handle("POST /api/v1/teams", http.HandlerFunc(newTeam))                                          //Create new team //Done
	mux.Handle("GET /api/v1/teams", http.HandlerFunc(teams))                                             //List all teams a user is part of //Done
	mux.Handle("GET /api/v1/teams/{teamID}", http.HandlerFunc(team))                                     //Get specific team details //WIP
	mux.Handle("DELETE /api/v1/teams/{teamID}", http.HandlerFunc(deleteTeam))                            //Delete team //Done
	mux.Handle("POST /api/v1/teams/{teamID}/members", http.HandlerFunc(addUserToTeam))                   //Add new user to team //Done
	mux.Handle("DELETE /api/v1/teams/{teamID}/members/{memberID}", http.HandlerFunc(removeUserFromTeam)) //Remove user from team //Done
	mux.Handle("GET /api/v1/teams/{teamID}/members", http.HandlerFunc(getTeamMembers))                   //Get all users in team //Done
	//Team Notes
	mux.Handle("POST /api/v1/teams/{teamID}/notes", http.HandlerFunc(teamNotes))                 //Post team Note //Done
	mux.Handle("GET /api/v1/teams/{teamID}/notes", http.HandlerFunc(getTeamNotes))               //Get all team notes //Done
	mux.Handle("GET /api/v1/teams/{teamID}/notes/{noteID}", http.HandlerFunc(getTeamNote))       //Get one team note //Done
	mux.Handle("PUT /api/v1/teams/{teamID}/notes/{noteID}", http.HandlerFunc(updateTeamNote))    //Update team Note //Done
	mux.Handle("DELETE /api/v1/teams/{teamID}/notes/{noteID}", http.HandlerFunc(deleteTeamNote)) //Delete team note based on id //Done

	server := &http.Server{Handler: mux, Addr: ":8080", ReadHeaderTimeout: time.Second * 10}
	fmt.Println("Listening on http://localhost:8080/")
	if err = server.ListenAndServe(); err != nil {
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

	err = Cfg.db.GivePremium(r.Context(), req.Data.UserId)
	if err != nil {
		http.Error(w, "User Not Found", http.StatusNotFound)
	}
	w.WriteHeader(http.StatusNoContent)
}
