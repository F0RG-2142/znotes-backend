package handlers

import (

	"net/http"

	"github.com/F0RG-2142/capstone-1/components"
)
func LandingPage(w http.ResponseWriter, r *http.Request){
	component := components.LandingPage()
	w.WriteHeader(http.StatusOK)
	component.Render(r.Context(), w)
}