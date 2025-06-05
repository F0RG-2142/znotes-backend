package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/F0RG-2142/capstone-1/handlers"
	"github.com/F0RG-2142/capstone-1/internal/auth"
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
	dbURL := os.Getenv("DB_URL")
	models.Cfg.Platform = os.Getenv("PLATFORM")
	models.Cfg.Secret = os.Getenv("JWT_SECRET")

	db, _ := sql.Open("postgres", dbURL)
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	mux := http.NewServeMux()
	//Frontend Routes
	mux.Handle("/", http.HandlerFunc(handlers.LandingPage))
	//Utility and admin
	mux.Handle("GET /api/v1/healthz", http.HandlerFunc(readiness))         //Check if server is ready //Done
	mux.Handle("GET /api/v1/admin/metrics", http.HandlerFunc(metrics))     //Server metrics endpoint //---
	mux.Handle("POST /api/v1/payment/webhooks", http.HandlerFunc(payment)) //Payment platform webhook //---
	//Users and auth
	mux.Handle("POST /api/v1/register", http.HandlerFunc(handlers.NewUser))          //New User Registration
	mux.Handle("POST /api/v1/login", http.HandlerFunc(handlers.Login))               //Login to profile
	mux.Handle("POST /api/v1/logout", http.HandlerFunc(handlers.RevokeRefreshToken)) //Revoke refresh tok
	mux.Handle("POST /api/v1/token/refresh", http.HandlerFunc(handlers.RefreshJWT))  //Refresh JWT
	mux.Handle("PUT /api/v1/user/me", http.HandlerFunc(handlers.UpdateUser))         //Update user details
	//Private Notes
	mux.Handle("POST /api/v1/notes", http.HandlerFunc(handlers.Notes))              //Post Private Note //Done
	mux.Handle("GET /api/v1/notes", http.HandlerFunc(handlers.GetNotes))            //Get all private notes //Done
	mux.Handle("GET /api/v1/notes/{noteID}", http.HandlerFunc(handlers.GetNote))    //Get one private note //Done
	mux.Handle("PUT /api/v1/notes/{noteID}", http.HandlerFunc(handlers.UpdateNote)) //Update private note //Done
	mux.Handle("DELETE /api/notes/{noteID}", http.HandlerFunc(handlers.DeleteNote)) //Delete note based on id //Done
	//Teams
	mux.Handle("POST /api/v1/teams", http.HandlerFunc(handlers.NewTeam))                                          //Create new team
	mux.Handle("GET /api/v1/teams", http.HandlerFunc(handlers.Teams))                                             //List all teams a user is part of
	mux.Handle("GET /api/v1/teams/{teamID}", http.HandlerFunc(handlers.Team))                                     //Get specific team details
	mux.Handle("DELETE /api/v1/teams/{teamID}", http.HandlerFunc(handlers.DeleteTeam))                            //Delete team
	mux.Handle("POST /api/v1/teams/{teamID}/members", http.HandlerFunc(handlers.AddUserToTeam))                   //Add new user to team
	mux.Handle("DELETE /api/v1/teams/{teamID}/members/{memberID}", http.HandlerFunc(handlers.RemoveUserFromTeam)) //Remove user from team
	mux.Handle("GET /api/v1/teams/{teamID}/members", http.HandlerFunc(handlers.GetTeamMembers))                   //Get all users in team
	//Team Notes
	mux.Handle("POST /api/v1/teams/{teamID}/notes", http.HandlerFunc(handlers.TeamNotes))                 //Post team Note
	mux.Handle("GET /api/v1/teams/{teamID}/notes", http.HandlerFunc(handlers.GetTeamNotes))               //Get all team notes
	mux.Handle("GET /api/v1/teams/{teamID}/notes/{noteID}", http.HandlerFunc(handlers.GetTeamNote))       //Get one team note
	mux.Handle("PUT /api/v1/teams/{teamID}/notes/{noteID}", http.HandlerFunc(handlers.UpdateTeamNote))    //Update team Note
	mux.Handle("DELETE /api/v1/teams/{teamID}/notes/{noteID}", http.HandlerFunc(handlers.DeleteTeamNote)) //Delete team note based on id

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

	err = models.Cfg.DB.GivePremium(r.Context(), req.Data.UserId)
	if err != nil {
		http.Error(w, "User Not Found", http.StatusNotFound)
	}
	w.WriteHeader(http.StatusNoContent)
}
